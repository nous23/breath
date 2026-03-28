package ui

import (
	"fmt"
	"time"

	"breath/internal/timer"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ShowMainWindow 创建并显示主窗口
func ShowMainWindow(m *Manager) {
	w := m.app.NewWindow("Breath - 呼吸提醒")
	w.Resize(fyne.NewSize(420, 480))
	w.CenterOnScreen()
	w.SetMaster() // 设为主窗口，关闭时退出应用

	// ===== 标题区域 =====
	titleLabel := widget.NewRichTextFromMarkdown("# 🫁 Breath")
	subtitleLabel := widget.NewLabel("健康呼吸，定时提醒你休息")
	subtitleLabel.Alignment = fyne.TextAlignCenter

	// ===== 状态区域 =====
	stateLabel := widget.NewLabel("状态: 计算中...")
	stateLabel.Alignment = fyne.TextAlignCenter
	stateLabel.TextStyle = fyne.TextStyle{Bold: true}

	// ===== 倒计时显示 =====
	countdownLabel := widget.NewLabel("--:--")
	countdownLabel.Alignment = fyne.TextAlignCenter
	countdownLabel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}

	// 活跃时长进度条
	progress := widget.NewProgressBar()
	progress.Min = 0
	progress.Max = 1.0

	progressHint := widget.NewLabel("")
	progressHint.Alignment = fyne.TextAlignCenter

	// ===== 按钮区域 =====
	pauseBtn := widget.NewButtonWithIcon("暂停", theme.MediaPauseIcon(), nil)
	pauseBtn.Importance = widget.MediumImportance

	resetBtn := widget.NewButtonWithIcon("重置", theme.MediaReplayIcon(), nil)

	settingsBtn := widget.NewButtonWithIcon("设置", theme.SettingsIcon(), func() {
		m.ShowSettings()
	})

	// 暂停/恢复按钮逻辑
	pauseBtn.OnTapped = func() {
		if m.tracker == nil {
			return
		}
		if m.tracker.IsPaused() {
			m.tracker.Resume()
			pauseBtn.SetText("暂停")
			pauseBtn.SetIcon(theme.MediaPauseIcon())
		} else {
			m.tracker.Pause()
			pauseBtn.SetText("恢复")
			pauseBtn.SetIcon(theme.MediaPlayIcon())
		}
	}

	// 重置按钮逻辑
	resetBtn.OnTapped = func() {
		if m.tracker != nil {
			m.tracker.Reset()
		}
	}

	buttonRow := container.NewHBox(
		layout.NewSpacer(),
		pauseBtn,
		resetBtn,
		settingsBtn,
		layout.NewSpacer(),
	)

	// ===== 组装界面 =====
	content := container.NewVBox(
		layout.NewSpacer(),
		container.NewHBox(layout.NewSpacer(), titleLabel, layout.NewSpacer()),
		container.NewHBox(layout.NewSpacer(), subtitleLabel, layout.NewSpacer()),
		widget.NewSeparator(),
		layout.NewSpacer(),
		stateLabel,
		countdownLabel,
		progress,
		progressHint,
		layout.NewSpacer(),
		widget.NewSeparator(),
		buttonRow,
		layout.NewSpacer(),
	)

	w.SetContent(container.NewPadded(content))

	// ===== 定时刷新 UI =====
	stopCh := make(chan struct{})
	w.SetOnClosed(func() {
		close(stopCh)
	})

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if m.tracker == nil {
					continue
				}

				state := m.tracker.GetState()
				activeDur := m.tracker.GetActiveDuration()
				remaining := m.tracker.GetRemainingTime()
				threshold := m.cfg.ActiveThreshold()

				// 更新状态文字
				switch state {
				case timer.StatePaused:
					stateLabel.SetText("⏸ 已暂停")
				case timer.StateIdle:
					stateLabel.SetText("💤 空闲中")
				case timer.StateActive:
					stateLabel.SetText("🟢 活跃中")
				}

				// 更新倒计时
				if state == timer.StatePaused {
					countdownLabel.SetText("--:--")
				} else {
					remainMin := int(remaining.Minutes())
					remainSec := int(remaining.Seconds()) % 60
					countdownLabel.SetText(fmt.Sprintf("距离下次休息: %02d:%02d", remainMin, remainSec))
				}

				// 更新进度条
				if threshold > 0 {
					prog := float64(activeDur) / float64(threshold)
					if prog > 1.0 {
						prog = 1.0
					}
					progress.SetValue(prog)
				}

				// 更新进度提示
				activeMin := int(activeDur.Minutes())
				activeSec := int(activeDur.Seconds()) % 60
				thresholdMin := int(threshold.Minutes())
				progressHint.SetText(fmt.Sprintf("已活跃 %02d:%02d / 阈值 %d 分钟", activeMin, activeSec, thresholdMin))

				// 更新暂停按钮状态
				if state == timer.StatePaused {
					pauseBtn.SetText("恢复")
					pauseBtn.SetIcon(theme.MediaPlayIcon())
				} else {
					pauseBtn.SetText("暂停")
					pauseBtn.SetIcon(theme.MediaPauseIcon())
				}

			case <-stopCh:
				return
			}
		}
	}()

	// 保存主窗口引用
	m.mainWindow = w

	w.Show()
}
