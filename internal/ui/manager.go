package ui

import (
	"breath/internal/config"
	"breath/internal/timer"

	"fyne.io/fyne/v2"
)

// Manager UI 管理器，统一管理所有窗口
type Manager struct {
	app        fyne.App
	cfg        *config.Config
	tracker    *timer.ActivityTracker
	mainWindow fyne.Window
}

// NewManager 创建 UI 管理器
func NewManager(app fyne.App, cfg *config.Config) *Manager {
	return &Manager{
		app: app,
		cfg: cfg,
	}
}

// SetTracker 设置活跃状态追踪器
func (m *Manager) SetTracker(tracker *timer.ActivityTracker) {
	m.tracker = tracker
}

// GetApp 获取 Fyne 应用实例
func (m *Manager) GetApp() fyne.App {
	return m.app
}

// GetConfig 获取配置
func (m *Manager) GetConfig() *config.Config {
	return m.cfg
}

// GetTracker 获取追踪器
func (m *Manager) GetTracker() *timer.ActivityTracker {
	return m.tracker
}

// ShowReminder 显示休息提醒弹窗（即使主窗口最小化也会弹出）
func (m *Manager) ShowReminder(activeDuration float64) {
	ShowReminderWindow(m, activeDuration)
}

// ShowCountdown 显示休息倒计时窗口
func (m *Manager) ShowCountdown() {
	ShowCountdownWindow(m)
}

// ShowSettings 显示设置窗口
func (m *Manager) ShowSettings() {
	ShowSettingsWindow(m)
}