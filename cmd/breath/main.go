package main

import (
	"log"

	"breath/assets"
	"breath/internal/config"
	"breath/internal/detector"
	"breath/internal/timer"
	"breath/internal/tray"
	"breath/internal/ui"

	"fyne.io/fyne/v2/app"
)

func main() {
	// 创建 Fyne 应用
	a := app.NewWithID("com.breath.app")
	a.SetIcon(assets.AppIcon)

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Printf("加载配置失败，使用默认配置: %v", err)
		cfg = config.Default()
	}

	// 创建空闲检测器
	idleDetector := detector.NewIdleDetector()

	// 创建 UI 管理器
	uiManager := ui.NewManager(a, cfg)

	// 创建活跃状态追踪器
	tracker := timer.NewActivityTracker(cfg, idleDetector, func(activeDuration float64) {
		// 活跃时长达到阈值时的回调
		uiManager.ShowReminder(activeDuration)
	})

	// 创建系统托盘
	trayManager := tray.NewManager(tracker, uiManager, cfg)

	// 设置 UI 管理器的依赖
	uiManager.SetTracker(tracker)
	uiManager.SetTrayManager(trayManager)

	// 启动活跃状态追踪
	tracker.Start()

	// 初始化系统托盘
	trayManager.Setup(a)

	// 运行应用（Fyne 主循环，此处会阻塞直到应用退出）
	a.Run()

	// 应用退出时保存配置
	if err := config.Save(cfg); err != nil {
		log.Printf("保存配置失败: %v", err)
	}
	tracker.Stop()
}