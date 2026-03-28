package ui

import (
	"fmt"
	"log"
	"strconv"

	"breath/internal/autostart"
	"breath/internal/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// ShowSettingsWindow 显示设置窗口
func ShowSettingsWindow(m *Manager) {
	w := m.app.NewWindow("Breath - 设置")
	w.Resize(fyne.NewSize(450, 400))
	w.CenterOnScreen()

	cfg := m.cfg.Clone()

	// 活跃时长阈值
	activeLabel := widget.NewLabel(fmt.Sprintf("活跃时长阈值: %d 分钟", cfg.ActiveThresholdMin))
	activeSlider := widget.NewSlider(15, 120)
	activeSlider.Step = 5
	activeSlider.Value = float64(cfg.ActiveThresholdMin)
	activeSlider.OnChanged = func(v float64) {
		cfg.ActiveThresholdMin = int(v)
		activeLabel.SetText(fmt.Sprintf("活跃时长阈值: %d 分钟", int(v)))
	}

	// 休息时长
	breakLabel := widget.NewLabel(fmt.Sprintf("休息时长: %d 分钟", cfg.BreakDurationMin))
	breakSlider := widget.NewSlider(1, 30)
	breakSlider.Step = 1
	breakSlider.Value = float64(cfg.BreakDurationMin)
	breakSlider.OnChanged = func(v float64) {
		cfg.BreakDurationMin = int(v)
		breakLabel.SetText(fmt.Sprintf("休息时长: %d 分钟", int(v)))
	}

	// 空闲判定时长
	idleLabel := widget.NewLabel(fmt.Sprintf("空闲判定时长: %d 分钟", cfg.IdleThresholdMin))
	idleSlider := widget.NewSlider(1, 15)
	idleSlider.Step = 1
	idleSlider.Value = float64(cfg.IdleThresholdMin)
	idleSlider.OnChanged = func(v float64) {
		cfg.IdleThresholdMin = int(v)
		idleLabel.SetText(fmt.Sprintf("空闲判定时长: %d 分钟", int(v)))
	}

	// 推迟提醒间隔
	postponeLabel := widget.NewLabel(fmt.Sprintf("推迟提醒间隔: %d 分钟", cfg.PostponeIntervalMin))
	postponeSlider := widget.NewSlider(1, 15)
	postponeSlider.Step = 1
	postponeSlider.Value = float64(cfg.PostponeIntervalMin)
	postponeSlider.OnChanged = func(v float64) {
		cfg.PostponeIntervalMin = int(v)
		postponeLabel.SetText(fmt.Sprintf("推迟提醒间隔: %d 分钟", int(v)))
	}

	// 开机自启动
	autoStartCheck := widget.NewCheck("开机自启动", func(checked bool) {
		cfg.AutoStartEnabled = checked
	})
	autoStartCheck.Checked = cfg.AutoStartEnabled

	// 保存按钮
	saveBtn := widget.NewButton("保存", func() {
		cfg.Validate()
		m.cfg.Update(cfg)

		if err := config.Save(m.cfg); err != nil {
			// 显示错误对话框
			errWin := m.app.NewWindow("错误")
			errWin.SetContent(widget.NewLabel("保存配置失败: " + err.Error()))
			errWin.Resize(fyne.NewSize(300, 100))
			errWin.Show()
			return
		}

		// 处理开机自启动设置变更
		autoStarter := autostart.New()
		if cfg.AutoStartEnabled {
			if err := autoStarter.Enable(); err != nil {
				log.Printf("启用开机自启动失败: %v", err)
			}
		} else {
			if err := autoStarter.Disable(); err != nil {
				log.Printf("禁用开机自启动失败: %v", err)
			}
		}
		w.Close()
	})
	saveBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("取消", func() {
		w.Close()
	})

	// 版本信息
	versionLabel := widget.NewLabel("Breath v1.0.0")
	versionLabel.Alignment = fyne.TextAlignCenter

	content := container.NewVBox(
		widget.NewRichTextFromMarkdown("### ⚙️ 设置"),
		widget.NewSeparator(),

		activeLabel,
		activeSlider,
		widget.NewLabel(""),

		breakLabel,
		breakSlider,
		widget.NewLabel(""),

		idleLabel,
		idleSlider,
		widget.NewLabel(""),

		postponeLabel,
		postponeSlider,
		widget.NewLabel(""),

		autoStartCheck,
		widget.NewSeparator(),

		container.NewHBox(layout.NewSpacer(), saveBtn, cancelBtn, layout.NewSpacer()),
		layout.NewSpacer(),
		versionLabel,
	)

	w.SetContent(container.NewScroll(content))
	w.Show()
}

// 辅助函数：将字符串转为整数（带默认值）
func atoi(s string, defaultVal int) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return v
}
