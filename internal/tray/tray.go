package tray

import (
	"fmt"
	"time"

	"breath/internal/config"
	"breath/internal/timer"
	"breath/internal/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// Manager 系统托盘管理器
type Manager struct {
	tracker   *timer.ActivityTracker
	uiManager *ui.Manager
	cfg       *config.Config
	menu      *fyne.Menu
	app       fyne.App

	statusItem  *fyne.MenuItem
	pauseItem   *fyne.MenuItem
	updateTimer *time.Ticker
	stopCh      chan struct{}
}

// NewManager 创建系统托盘管理器
func NewManager(tracker *timer.ActivityTracker, uiManager *ui.Manager, cfg *config.Config) *Manager {
	return &Manager{
		tracker:   tracker,
		uiManager: uiManager,
		cfg:       cfg,
		stopCh:    make(chan struct{}),
	}
}

// Setup 初始化系统托盘
func (m *Manager) Setup(a fyne.App) {
	m.app = a

	// 检查是否支持系统托盘（桌面环境）
	if desk, ok := a.(desktop.App); ok {
		m.createMenu()
		desk.SetSystemTrayMenu(m.menu)

		// 启动状态更新定时器
		m.startStatusUpdater()
	}
}

// UpdateStatus 更新托盘菜单中的状态显示
func (m *Manager) UpdateStatus() {
	if m.statusItem == nil {
		return
	}

	state := m.tracker.GetState()
	switch state {
	case timer.StatePaused:
		m.statusItem.Label = "状态: ⏸ 已暂停"
	case timer.StateIdle:
		m.statusItem.Label = "状态: 💤 空闲中"
	case timer.StateActive:
		active := m.tracker.GetActiveDuration()
		remaining := m.tracker.GetRemainingTime()
		m.statusItem.Label = fmt.Sprintf("已活跃 %s | 剩余 %s",
			formatShortDuration(active),
			formatShortDuration(remaining))
	}
}

// createMenu 创建托盘菜单
func (m *Manager) createMenu() {
	// 状态项（不可点击，仅显示信息）
	m.statusItem = fyne.NewMenuItem("状态: 计算中...", nil)
	m.statusItem.Disabled = true

	// 暂停/恢复项
	m.pauseItem = fyne.NewMenuItem("⏸ 暂停", func() {
		if m.tracker.IsPaused() {
			m.tracker.Resume()
			m.pauseItem.Label = "⏸ 暂停"
		} else {
			m.tracker.Pause()
			m.pauseItem.Label = "▶️ 恢复"
		}
		m.UpdateStatus()
	})

	// 设置项
	settingsItem := fyne.NewMenuItem("⚙️ 设置", func() {
		m.uiManager.ShowSettings()
	})

	// 退出项
	quitItem := fyne.NewMenuItem("🚪 退出", func() {
		m.cleanup()
		m.app.Quit()
	})

	m.menu = fyne.NewMenu("Breath",
		m.statusItem,
		fyne.NewMenuItemSeparator(),
		m.pauseItem,
		settingsItem,
		fyne.NewMenuItemSeparator(),
		quitItem,
	)
}

// startStatusUpdater 启动定时更新状态的协程
func (m *Manager) startStatusUpdater() {
	m.updateTimer = time.NewTicker(1 * time.Second)

	go func() {
		for {
			select {
			case <-m.updateTimer.C:
				m.UpdateStatus()
			case <-m.stopCh:
				return
			}
		}
	}()
}

// cleanup 清理资源
func (m *Manager) cleanup() {
	if m.updateTimer != nil {
		m.updateTimer.Stop()
	}
	close(m.stopCh)

	// 保存配置
	if err := config.Save(m.cfg); err != nil {
		fmt.Printf("保存配置失败: %v\n", err)
	}
}

// formatShortDuration 将 time.Duration 格式化为简短文字
func formatShortDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	totalMin := int(d.Minutes())
	if totalMin >= 60 {
		hours := totalMin / 60
		mins := totalMin % 60
		return fmt.Sprintf("%dh%dm", hours, mins)
	}
	totalSec := int(d.Seconds())
	min := totalSec / 60
	sec := totalSec % 60
	return fmt.Sprintf("%dm%ds", min, sec)
}
