//go:build darwin

package notify

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/gen2brain/beeep"
)

// SendAlert 发送强制弹窗提醒（macOS: osascript display dialog，会打断全屏模式）
// 通过回调 onAction 异步通知调用方用户的选择
func SendAlert(title, message string, onAction func(AlertAction)) {
	go func() {
		action := sendDarwinAlert(title, message)
		if onAction != nil {
			onAction(action)
		}
	}()
}

// sendDarwinAlert 使用 osascript 弹出模态对话框，同步等待用户操作并返回选择
func sendDarwinAlert(title, message string) AlertAction {
	osa, err := exec.LookPath("osascript")
	if err != nil {
		// 降级为普通通知
		_ = beeep.Alert(title, message, "")
		return ActionError
	}

	// 使用 display dialog 创建模态对话框
	// - with title: 对话框标题
	// - with icon caution: 显示警告图标
	// - giving up after 300: 5分钟后自动关闭（防止永久阻塞）
	// osascript 返回格式: "button returned:去休息, gave up:false"
	script := fmt.Sprintf(
		`display dialog %q with title %q buttons {"稍后提醒", "去休息"} default button "去休息" with icon caution giving up after 300`,
		message, title,
	)

	cmd := exec.Command(osa, "-e", script)
	output, err := cmd.Output()
	if err != nil {
		// 用户点击了关闭按钮（取消），或者发生错误
		return ActionPostpone
	}

	result := string(output)
	// 解析 osascript 返回值
	// 正常返回: "button returned:去休息, gave up:false\n"
	// 超时返回: "button returned:, gave up:true\n"
	if strings.Contains(result, "gave up:true") {
		return ActionTimeout
	}
	if strings.Contains(result, "去休息") {
		return ActionStartBreak
	}
	return ActionPostpone
}
