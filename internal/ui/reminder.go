package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ShowReminderWindow 显示休息提醒弹窗
func ShowReminderWindow(m *Manager, activeDuration float64) {
	w := m.app.NewWindow("Breath - 该休息了")
	w.SetFixedSize(true)
	w.Resize(fyne.NewSize(420, 320))
	w.CenterOnScreen()

	// 获取推迟次数
	postponeCount := 0
	if m.tracker != nil {
		postponeCount = m.tracker.GetPostponeCount()
	}

	// 标题图标
	titleIcon := widget.NewIcon(theme.WarningIcon())

	// 标题
	titleLabel := widget.NewRichTextFromMarkdown("## 🫁 该休息一下了")
	titleLabel.Wrapping = fyne.TextWrapWord

	// 使用时长提示
	hours := int(activeDuration) / 60
	mins := int(activeDuration) % 60
	var durationText string
	if hours > 0 {
		durationText = fmt.Sprintf("您已连续使用电脑 **%d 小时 %d 分钟**", hours, mins)
	} else {
		durationText = fmt.Sprintf("您已连续使用电脑 **%d 分钟**", mins)
	}
	durationLabel := widget.NewRichTextFromMarkdown(durationText)
	durationLabel.Wrapping = fyne.TextWrapWord

	// 健康提示
	tipLabel := widget.NewLabel("休息一下，远眺窗外，活动活动身体吧！")
	tipLabel.Alignment = fyne.TextAlignCenter

	// 额外健康警示（推迟超过3次）
	var warningLabel *widget.RichText
	if postponeCount >= 3 {
		warningLabel = widget.NewRichTextFromMarkdown(
			fmt.Sprintf("⚠️ 您已经推迟了 **%d** 次，长时间使用屏幕会对眼睛和身体造成伤害，请务必休息！", postponeCount),
		)
		warningLabel.Wrapping = fyne.TextWrapWord
	}

	// 按钮
	startBreakBtn := widget.NewButton("🧘 开始休息", func() {
		w.Close()
		if m.tracker != nil {
			m.tracker.Reset()
		}
		m.ShowCountdown()
	})
	startBreakBtn.Importance = widget.HighImportance

	postponeBtn := widget.NewButton("⏰ 稍后提醒", func() {
		w.Close()
		if m.tracker != nil {
			m.tracker.Postpone()
		}
	})

	buttons := container.NewHBox(
		layout.NewSpacer(),
		startBreakBtn,
		postponeBtn,
		layout.NewSpacer(),
	)

	// 组装界面
	content := container.NewVBox(
		layout.NewSpacer(),
		container.NewHBox(layout.NewSpacer(), titleIcon, titleLabel, layout.NewSpacer()),
		widget.NewSeparator(),
		container.NewHBox(layout.NewSpacer(), durationLabel, layout.NewSpacer()),
		tipLabel,
	)

	if warningLabel != nil {
		content.Add(widget.NewSeparator())
		content.Add(warningLabel)
	}

	content.Add(layout.NewSpacer())
	content.Add(buttons)
	content.Add(layout.NewSpacer())

	w.SetContent(content)

	// 设置窗口置顶（确保在全屏应用之上）
	w.SetOnClosed(func() {
		// 窗口关闭时的清理
	})

	w.Show()
	w.RequestFocus()
}
