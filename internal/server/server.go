package server

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"breath/internal/autostart"
	"breath/internal/config"
	"breath/internal/timer"
	"breath/internal/web"
)

// Server HTTP 服务器，提供 Web UI 和 API
type Server struct {
	cfg     *config.Config
	tracker *timer.ActivityTracker

	mu      sync.RWMutex
	clients map[chan string]struct{} // SSE 客户端

	port int
}

// NewServer 创建 HTTP 服务器
func NewServer(cfg *config.Config, tracker *timer.ActivityTracker) *Server {
	return &Server{
		cfg:     cfg,
		tracker: tracker,
		clients: make(map[chan string]struct{}),
	}
}

// GetPort 获取服务器监听端口
func (s *Server) GetPort() int {
	return s.port
}

// Start 启动 HTTP 服务器（非阻塞）
func (s *Server) Start() (int, error) {
	mux := http.NewServeMux()

	// API 路由
	mux.HandleFunc("/api/events", s.handleSSE)
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/toggle-pause", s.handleTogglePause)
	mux.HandleFunc("/api/reset", s.handleReset)
	mux.HandleFunc("/api/postpone", s.handlePostpone)
	mux.HandleFunc("/api/start-break", s.handleStartBreak)

	// 静态文件服务
	staticFS, err := fs.Sub(web.StaticFS, "static")
	if err != nil {
		return 0, fmt.Errorf("加载静态资源失败: %w", err)
	}
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	// 监听随机可用端口
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("监听端口失败: %w", err)
	}

	s.port = listener.Addr().(*net.TCPAddr).Port
	log.Printf("HTTP 服务器启动在 http://127.0.0.1:%d", s.port)

	// 启动状态推送协程
	go s.statusBroadcaster()

	// 启动 HTTP 服务
	go func() {
		if err := http.Serve(listener, mux); err != nil {
			log.Printf("HTTP 服务器错误: %v", err)
		}
	}()

	return s.port, nil
}

// SendReminder 向所有 SSE 客户端发送提醒事件
func (s *Server) SendReminder(activeDuration float64, postponeCount int) {
	data := map[string]interface{}{
		"active_minutes": activeDuration,
		"postpone_count": postponeCount,
	}
	jsonData, _ := json.Marshal(data)
	s.broadcast("reminder", string(jsonData))
}

// SendBreak 向所有 SSE 客户端发送开始休息事件（由系统弹窗触发）
func (s *Server) SendBreak() {
	breakSeconds := int(s.cfg.BreakDuration().Seconds())
	data := map[string]interface{}{
		"break_duration_seconds": breakSeconds,
	}
	jsonData, _ := json.Marshal(data)
	s.broadcast("break", string(jsonData))
}

// SendDismissReminder 向所有 SSE 客户端发送关闭提醒弹窗事件
func (s *Server) SendDismissReminder() {
	s.broadcast("dismiss_reminder", "{}")
}

// ===== SSE 处理 =====

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "不支持 SSE", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := make(chan string, 16)

	s.mu.Lock()
	s.clients[ch] = struct{}{}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, ch)
		s.mu.Unlock()
		close(ch)
	}()

	// 立即发送当前配置
	cfgData := s.getConfigJSON()
	fmt.Fprintf(w, "event: config\ndata: %s\n\n", cfgData)
	flusher.Flush()

	// 立即发送当前状态
	statusData := s.getStatusJSON()
	fmt.Fprintf(w, "event: status\ndata: %s\n\n", statusData)
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprint(w, msg)
			flusher.Flush()
		case <-ctx.Done():
			return
		}
	}
}

func (s *Server) broadcast(event, data string) {
	msg := fmt.Sprintf("event: %s\ndata: %s\n\n", event, data)

	s.mu.RLock()
	defer s.mu.RUnlock()

	for ch := range s.clients {
		select {
		case ch <- msg:
		default:
			// 客户端缓冲区满，跳过
		}
	}
}

func (s *Server) statusBroadcaster() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		data := s.getStatusJSON()
		s.broadcast("status", data)
	}
}

func (s *Server) getStatusJSON() string {
	state := s.tracker.GetState()
	activeDur := s.tracker.GetActiveDuration()
	remaining := s.tracker.GetRemainingTime()
	threshold := s.cfg.ActiveThreshold()

	stateStr := "active"
	switch state {
	case timer.StateIdle:
		stateStr = "idle"
	case timer.StatePaused:
		stateStr = "paused"
	}

	postponeCount := s.tracker.GetPostponeCount()

	data := map[string]interface{}{
		"state":             stateStr,
		"active_seconds":    activeDur.Seconds(),
		"remaining_seconds": remaining.Seconds(),
		"threshold_seconds": threshold.Seconds(),
		"postpone_count":    postponeCount,
	}
	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}

func (s *Server) getConfigJSON() string {
	s.cfg.Validate()
	clone := s.cfg.Clone()
	data := map[string]interface{}{
		"active_threshold_min":  clone.ActiveThresholdMin,
		"break_duration_min":    clone.BreakDurationMin,
		"idle_threshold_min":    clone.IdleThresholdMin,
		"postpone_interval_min": clone.PostponeIntervalMin,
		"auto_start_enabled":    clone.AutoStartEnabled,
	}
	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}

// ===== API 处理 =====

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, s.getConfigJSON())

	case http.MethodPost:
		var req struct {
			ActiveThresholdMin  int  `json:"active_threshold_min"`
			BreakDurationMin    int  `json:"break_duration_min"`
			IdleThresholdMin    int  `json:"idle_threshold_min"`
			PostponeIntervalMin int  `json:"postpone_interval_min"`
			AutoStartEnabled    bool `json:"auto_start_enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "无效的请求体", http.StatusBadRequest)
			return
		}

		newCfg := &config.Config{
			ActiveThresholdMin:  req.ActiveThresholdMin,
			BreakDurationMin:    req.BreakDurationMin,
			IdleThresholdMin:    req.IdleThresholdMin,
			PostponeIntervalMin: req.PostponeIntervalMin,
			AutoStartEnabled:    req.AutoStartEnabled,
		}
		newCfg.Validate()
		s.cfg.Update(newCfg)

		if err := config.Save(s.cfg); err != nil {
			http.Error(w, "保存配置失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 重置 tracker 使新配置生效
		if s.tracker != nil {
			s.tracker.Reset()
		}

		// 处理开机自启动
		autoStarter := autostart.New()
		if req.AutoStartEnabled {
			if err := autoStarter.Enable(); err != nil {
				log.Printf("启用开机自启动失败: %v", err)
			}
		} else {
			if err := autoStarter.Disable(); err != nil {
				log.Printf("禁用开机自启动失败: %v", err)
			}
		}

		// 广播新配置
		s.broadcast("config", s.getConfigJSON())

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"ok":true}`)

	default:
		http.Error(w, "不支持的方法", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleTogglePause(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "不支持的方法", http.StatusMethodNotAllowed)
		return
	}

	if s.tracker.IsPaused() {
		s.tracker.Resume()
	} else {
		s.tracker.Pause()
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"ok":true}`)
}

func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "不支持的方法", http.StatusMethodNotAllowed)
		return
	}

	s.tracker.Reset()

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"ok":true}`)
}

func (s *Server) handlePostpone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "不支持的方法", http.StatusMethodNotAllowed)
		return
	}

	s.tracker.Postpone()

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"ok":true}`)
}

func (s *Server) handleStartBreak(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "不支持的方法", http.StatusMethodNotAllowed)
		return
	}

	s.tracker.Reset()
	breakSeconds := int(s.cfg.BreakDuration().Seconds())

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"ok":true,"break_duration_seconds":%d}`, breakSeconds)
}
