package timer

import (
	"sync"
	"time"

	"breath/internal/config"
	"breath/internal/detector"
)

// State 表示用户的活跃状态
type State int

const (
	// StateActive 用户活跃状态
	StateActive State = iota
	// StateIdle 用户空闲状态
	StateIdle
	// StatePaused 追踪器暂停状态（由用户手动暂停）
	StatePaused
)

// OnThresholdReached 活跃时长达到阈值时的回调函数类型
// activeDuration 参数为已活跃的分钟数
type OnThresholdReached func(activeDuration float64)

// ActivityTracker 活跃状态追踪器
type ActivityTracker struct {
	mu sync.RWMutex

	cfg      *config.Config
	detector detector.IdleDetector
	onAlert  OnThresholdReached

	state         State
	activeStart   time.Time     // 本轮活跃开始时间
	activeDur     time.Duration // 累计活跃时长
	idleStart     time.Time     // 进入空闲的时间
	postponeCount int           // 推迟提醒次数
	alerted       bool          // 本轮是否已触发过提醒

	ticker     *time.Ticker
	stopCh     chan struct{}
	postponeCh chan struct{} // 用于取消所有正在运行的 Postpone 定时器
}

// NewActivityTracker 创建活跃状态追踪器
func NewActivityTracker(cfg *config.Config, det detector.IdleDetector, onAlert OnThresholdReached) *ActivityTracker {
	return &ActivityTracker{
		cfg:        cfg,
		detector:   det,
		onAlert:    onAlert,
		state:      StateActive,
		stopCh:     make(chan struct{}),
		postponeCh: make(chan struct{}),
	}
}

// SetOnAlert 设置活跃阈值回调函数
func (t *ActivityTracker) SetOnAlert(fn OnThresholdReached) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onAlert = fn
}

// Start 启动活跃状态追踪
func (t *ActivityTracker) Start() {
	t.mu.Lock()
	t.activeStart = time.Now()
	t.activeDur = 0
	t.state = StateActive
	t.alerted = false
	t.postponeCount = 0
	t.ticker = time.NewTicker(1 * time.Second)
	t.mu.Unlock()

	go t.loop()
}

// Stop 停止活跃状态追踪
func (t *ActivityTracker) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.ticker != nil {
		t.ticker.Stop()
	}
	close(t.stopCh)
}

// Pause 暂停追踪
func (t *ActivityTracker) Pause() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state != StatePaused {
		// 保存当前累计的活跃时长
		if t.state == StateActive {
			t.activeDur += time.Since(t.activeStart)
		}
		t.state = StatePaused
	}
}

// Resume 恢复追踪
func (t *ActivityTracker) Resume() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state == StatePaused {
		t.state = StateActive
		t.activeStart = time.Now()
	}
}

// Reset 重置活跃计时器
func (t *ActivityTracker) Reset() {
	t.mu.Lock()

	t.activeDur = 0
	t.activeStart = time.Now()
	t.alerted = false
	t.postponeCount = 0
	if t.state != StatePaused {
		t.state = StateActive
	}

	// 关闭旧的 postponeCh，取消所有正在运行的 Postpone 定时器
	close(t.postponeCh)
	// 创建新的 postponeCh 供后续使用
	t.postponeCh = make(chan struct{})

	t.mu.Unlock()
}

// Postpone 推迟提醒，在推迟间隔后重新触发
func (t *ActivityTracker) Postpone() {
	t.mu.Lock()
	t.postponeCount++
	// 注意：保持 alerted = true，不设置为 false
	// 这样 tick() 不会因为活跃时长仍超过阈值而立即重复触发提醒
	// 推迟后的重新提醒完全由下面的定时器 goroutine 负责
	// 捕获当前的 postponeCh，用于检测 Reset 是否发生
	currentPostponeCh := t.postponeCh
	t.mu.Unlock()

	// 在推迟间隔后重新触发提醒
	postponeInterval := t.cfg.PostponeInterval()
	go func() {
		timer := time.NewTimer(postponeInterval)
		defer timer.Stop()

		select {
		case <-timer.C:
			t.mu.RLock()
			state := t.state
			onAlert := t.onAlert
			activeMins := t.GetActiveDuration().Minutes()
			t.mu.RUnlock()

			if state == StateActive && onAlert != nil {
				t.mu.Lock()
				t.alerted = true
				t.mu.Unlock()
				onAlert(activeMins)
			}
		case <-currentPostponeCh:
			// Reset 被调用，取消本次推迟提醒
			return
		case <-t.stopCh:
			return
		}
	}()
}

// GetState 获取当前状态
func (t *ActivityTracker) GetState() State {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.state
}

// IsPaused 是否处于暂停状态
func (t *ActivityTracker) IsPaused() bool {
	return t.GetState() == StatePaused
}

// GetActiveDuration 获取当前累计活跃时长
func (t *ActivityTracker) GetActiveDuration() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()

	dur := t.activeDur
	if t.state == StateActive {
		dur += time.Since(t.activeStart)
	}
	return dur
}

// GetRemainingTime 获取距离下次提醒的剩余时间
func (t *ActivityTracker) GetRemainingTime() time.Duration {
	active := t.GetActiveDuration()
	threshold := t.cfg.ActiveThreshold()
	remaining := threshold - active
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetPostponeCount 获取推迟次数
func (t *ActivityTracker) GetPostponeCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.postponeCount
}

// loop 主循环：每秒检查一次空闲时间并更新状态
func (t *ActivityTracker) loop() {
	for {
		select {
		case <-t.ticker.C:
			t.tick()
		case <-t.stopCh:
			return
		}
	}
}

// tick 单次检查逻辑
func (t *ActivityTracker) tick() {
	t.mu.Lock()

	// 暂停状态不做任何处理
	if t.state == StatePaused {
		t.mu.Unlock()
		return
	}

	idleDuration := t.detector.GetIdleDuration()
	idleThreshold := t.cfg.IdleThreshold()
	breakDuration := t.cfg.BreakDuration()

	switch t.state {
	case StateActive:
		if idleDuration >= idleThreshold {
			// 用户空闲超过阈值，切换到空闲状态
			t.activeDur += time.Since(t.activeStart)
			t.idleStart = time.Now().Add(-idleDuration) // 估算真实的空闲开始时间
			t.state = StateIdle
			t.mu.Unlock()
			return
		}

		// 检查是否达到活跃阈值
		currentActive := t.activeDur + time.Since(t.activeStart)
		activeThreshold := t.cfg.ActiveThreshold()

		if currentActive >= activeThreshold && !t.alerted {
			t.alerted = true
			onAlert := t.onAlert
			activeMins := currentActive.Minutes()
			t.mu.Unlock()

			if onAlert != nil {
				onAlert(activeMins)
			}
			return
		}

	case StateIdle:
		if idleDuration < idleThreshold {
			// 用户恢复输入
			totalIdleDuration := time.Since(t.idleStart)

			if totalIdleDuration >= breakDuration {
				// 空闲时间已超过休息时长，重置计时器
				t.activeDur = 0
				t.alerted = false
				t.postponeCount = 0
			}
			// 否则继续累计之前的活跃时长

			t.activeStart = time.Now()
			t.state = StateActive
			t.mu.Unlock()
			return
		}
	}

	t.mu.Unlock()
}
