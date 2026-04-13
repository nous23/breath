//go:build !darwin && !windows

package notify

import (
	"github.com/gen2brain/beeep"
)

// SendAlert 发送强制弹窗提醒（其他平台: 降级为 beeep.Alert）
// 通过回调 onAction 异步通知调用方用户的选择
func SendAlert(title, message string, onAction func(AlertAction)) {
	go func() {
		err := beeep.Alert(title, message, "")
		if onAction != nil {
			if err != nil {
				onAction(ActionError)
			} else {
				// 其他平台无法获取用户选择，默认当作超时处理
				onAction(ActionTimeout)
			}
		}
	}()
}
