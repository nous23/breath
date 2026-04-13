package main

import (
	"fmt"
	"log"

	"breath/internal/browser"
	"breath/internal/config"
	"breath/internal/detector"
	"breath/internal/notify"
	"breath/internal/server"
	"breath/internal/timer"
	"breath/internal/tray"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Printf("加载配置失败，使用默认配置: %v", err)
		cfg = config.Default()
	}

	// 创建空闲检测器
	idleDetector := detector.NewIdleDetector()

	// 创建 HTTP 服务器（先声明，后面设置回调时需要引用）
	var srv *server.Server

	// 创建活跃状态追踪器（先用 nil 回调创建，后面再设置）
	tracker := timer.NewActivityTracker(cfg, idleDetector, nil)

	// 设置活跃阈值回调
	tracker.SetOnAlert(func(activeDuration float64) {
		postponeCount := tracker.GetPostponeCount()

		// 1. 通过 SSE 推送提醒到 Web UI（如果浏览器打开着）
		if srv != nil {
			srv.SendReminder(activeDuration, postponeCount)
		}

		// 2. 首次提醒时自动打开浏览器页面，推迟后的重复提醒不再打开新窗口
		if srv != nil && postponeCount == 0 {
			browserURL := fmt.Sprintf("http://127.0.0.1:%d", srv.GetPort())
			if err := browser.Open(browserURL); err != nil {
				log.Printf("打开浏览器失败: %v", err)
			}
		}

		// 3. 发送系统级强制弹窗（会打断全屏模式）
		// 弹窗会异步等待用户操作，通过回调处理用户选择
		mins := int(activeDuration)
		title := "🫁 Breath - 该休息了"
		message := fmt.Sprintf("您已连续使用电脑 %d 分钟，休息一下吧！", mins)
		notify.SendAlert(title, message, func(action notify.AlertAction) {
			switch action {
			case notify.ActionStartBreak:
				log.Println("用户在系统弹窗选择了「去休息」")
				// 重置 tracker 计时器
				tracker.Reset()
				// 通知 Web UI 关闭提醒弹窗并进入休息倒计时
				if srv != nil {
					srv.SendDismissReminder()
					srv.SendBreak()
				}
			case notify.ActionPostpone:
				log.Println("用户在系统弹窗选择了「稍后提醒」")
				// 推迟提醒
				tracker.Postpone()
				// 通知 Web UI 关闭提醒弹窗
				if srv != nil {
					srv.SendDismissReminder()
				}
			case notify.ActionTimeout:
				log.Println("系统弹窗超时自动关闭")
				// 超时视为推迟
				tracker.Postpone()
				if srv != nil {
					srv.SendDismissReminder()
				}
			case notify.ActionError:
				log.Println("系统弹窗发生错误")
			}
		})
	})

	// 创建 HTTP 服务器
	srv = server.NewServer(cfg, tracker)

	// 启动 HTTP 服务器
	port, err := srv.Start()
	if err != nil {
		log.Fatalf("启动 HTTP 服务器失败: %v", err)
	}

	url := fmt.Sprintf("http://127.0.0.1:%d", port)

	// 启动活跃状态追踪
	tracker.Start()

	// 自动打开浏览器
	if err := browser.Open(url); err != nil {
		log.Printf("打开浏览器失败: %v", err)
	}

	// 创建系统托盘管理器
	trayManager := tray.NewManager(tracker, cfg)
	trayManager.SetOnOpenBrowser(func() {
		if err := browser.Open(url); err != nil {
			log.Printf("打开浏览器失败: %v", err)
		}
	})
	trayManager.SetOnQuit(func() {
		// 退出时保存配置
		if err := config.Save(cfg); err != nil {
			log.Printf("保存配置失败: %v", err)
		}
		tracker.Stop()
	})

	// 运行系统托盘（阻塞主线程）
	trayManager.Run()
}