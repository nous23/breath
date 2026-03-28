package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// ShowCountdownWindow 显示休息倒计时窗口
func ShowCountdownWindow(m *Manager) {
	w := m.app.NewWindow("Breath - 休息中")
	w.SetFixedSize(true)
	w.Resize(fyne.NewSize(380, 280))
	w.CenterOnScreen()

	breakDuration := m.cfg.BreakDuration()
	remaining := breakDuration

	// 标题
	titleLabel := widget.NewRichTextFromMarkdown("## 🌿 休息时间")
	titleLabel.Wrapping = fyne.TextWrapWord

	// 倒计时显示
	countdownLabel := widget.NewLabel(formatDuration(remaining))
	countdownLabel.Alignment = fyne.TextAlignCenter
	countdownLabel.TextStyle = fyne.TextStyle{Bold: true}

	// 提示语
	tipLabel := widget.NewLabel("站起来走走，让眼睛看看远方 🌈")
	tipLabel.Alignment = fyne.TextAlignCenter

	// 进度条
	progress := widget.NewProgressBar()
	progress.Max = float64(breakDuration.Seconds())
	progress.SetValue(0)

	// 跳过休息按钮
	skipBtn := widget.NewButton("⏭ 跳过休息", func() {
		w.Close()
	})

	content := container.NewVBox(
		layout.NewSpacer(),
		container.NewHBox(layout.NewSpacer(), titleLabel, layout.NewSpacer()),
		widget.NewSeparator(),
		countdownLabel,
		progress,
		tipLabel,
		layout.NewSpacer(),
		container.NewHBox(layout.NewSpacer(), skipBtn, layout.NewSpacer()),
		layout.NewSpacer(),
	)

	w.SetContent(content)

	// 启动倒计时协程
	stopCh := make(chan struct{})
	w.SetOnClosed(func() {
		close(stopCh)
		// 无论是倒计时结束还是手动跳过，都重置计时器
		if m.tracker != nil {
			m.tracker.Reset()
		}
	})

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		elapsed := time.Duration(0)

		for {
			select {
			case <-ticker.C:
				elapsed += time.Second
				remaining = breakDuration - elapsed

				if remaining <= 0 {
					// 倒计时结束，关闭窗口
					w.Close()
					return
				}

				// 更新 UI（必须在主线程）
				countdownLabel.SetText(formatDuration(remaining))
				progress.SetValue(float64(elapsed.Seconds()))

			case <-stopCh:
				return
			}
		}
	}()

	w.Show()
	w.RequestFocus()
}

// formatDuration 将 time.Duration 格式化为 "MM:SS" 格式
func formatDuration(d time.Duration) string {
	totalSec := int(d.Seconds())
	if totalSec < 0 {
		totalSec = 0
	}
	min := totalSec / 60
	sec := totalSec % 60
	return fmt.Sprintf("%02d:%02d", min, sec)
}
