package tray

import (
	"fmt"
	"runtime"
	"time"

	"breath/assets"
	"breath/internal/config"
	"breath/internal/timer"

	"github.com/getlantern/systray"
)

// Manager 系统托盘管理器
type Manager struct {
	tracker *timer.ActivityTracker
	cfg     *config.Config

	onOpenBrowser func() // 打开浏览器的回调
	onQuit        func() // 退出应用的回调

	mStatus *systray.MenuItem
	mPause  *systray.MenuItem
}

// NewManager 创建系统托盘管理器
func NewManager(tracker *timer.ActivityTracker, cfg *config.Config) *Manager {
	return &Manager{
		tracker: tracker,
		cfg:     cfg,
	}
}

// SetOnOpenBrowser 设置打开浏览器的回调
func (m *Manager) SetOnOpenBrowser(fn func()) {
	m.onOpenBrowser = fn
}

// SetOnQuit 设置退出应用的回调
func (m *Manager) SetOnQuit(fn func()) {
	m.onQuit = fn
}

// Run 启动系统托盘（阻塞，应在主 goroutine 中调用）
func (m *Manager) Run() {
	systray.Run(m.onReady, m.onExit)
}

// Quit 退出系统托盘
func (m *Manager) Quit() {
	systray.Quit()
}

// getIcon 根据当前平台返回合适格式的图标数据
// Windows 需要 ICO 格式，macOS/Linux 使用 PNG 格式
func getIcon(pngData, icoData []byte) []byte {
	if runtime.GOOS == "windows" {
		return icoData
	}
	return pngData
}

func (m *Manager) onReady() {
	systray.SetIcon(getIcon(assets.AppIconPNG, assets.AppIconICO))
	systray.SetTitle("Breath")
	systray.SetTooltip("Breath - 呼吸提醒")

	// 状态项
	m.mStatus = systray.AddMenuItem("状态: 计算中...", "当前状态")
	m.mStatus.Disable()

	systray.AddSeparator()

	// 打开界面
	mOpen := systray.AddMenuItem("🌐 打开界面", "在浏览器中打开")

	// 暂停/恢复
	m.mPause = systray.AddMenuItem("⏸ 暂停", "暂停/恢复追踪")

	// 重置
	mReset := systray.AddMenuItem("🔄 重置", "重置计时器")

	systray.AddSeparator()

	// 退出
	mQuit := systray.AddMenuItem("🚪 退出", "退出应用")

	// 启动状态更新
	go m.statusUpdater()

	// 事件循环
	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				if m.onOpenBrowser != nil {
					m.onOpenBrowser()
				}
			case <-m.mPause.ClickedCh:
				if m.tracker.IsPaused() {
					m.tracker.Resume()
					m.mPause.SetTitle("⏸ 暂停")
					systray.SetIcon(getIcon(assets.AppIconPNG, assets.AppIconICO))
				} else {
					m.tracker.Pause()
					m.mPause.SetTitle("▶️ 恢复")
					systray.SetIcon(getIcon(assets.PausedIconPNG, assets.PausedIconICO))
				}
			case <-mReset.ClickedCh:
				m.tracker.Reset()
			case <-mQuit.ClickedCh:
				if m.onQuit != nil {
					m.onQuit()
				}
				systray.Quit()
				return
			}
		}
	}()
}

func (m *Manager) onExit() {
	// 清理资源
}

func (m *Manager) statusUpdater() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if m.mStatus == nil {
			continue
		}

		state := m.tracker.GetState()
		switch state {
		case timer.StatePaused:
			m.mStatus.SetTitle("状态: ⏸ 已暂停")
		case timer.StateIdle:
			m.mStatus.SetTitle("状态: 💤 空闲中")
		case timer.StateActive:
			active := m.tracker.GetActiveDuration()
			remaining := m.tracker.GetRemainingTime()
			m.mStatus.SetTitle(fmt.Sprintf("已活跃 %s | 剩余 %s",
				formatShortDuration(active),
				formatShortDuration(remaining)))
		}

		// 更新暂停按钮状态
		if m.mPause != nil {
			if state == timer.StatePaused {
				m.mPause.SetTitle("▶️ 恢复")
			} else {
				m.mPause.SetTitle("⏸ 暂停")
			}
		}
	}
}

// formatShortDuration 将 time.Duration 格式化为简短文字
func formatShortDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	totalSec := int(d.Seconds())
	min := totalSec / 60
	sec := totalSec % 60
	if min >= 60 {
		hours := min / 60
		mins := min % 60
		return fmt.Sprintf("%dh%dm", hours, mins)
	}
	return fmt.Sprintf("%dm%ds", min, sec)
}