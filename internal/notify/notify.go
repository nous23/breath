package notify

import (
	"github.com/gen2brain/beeep"
)

// AlertAction 表示用户在弹窗中的选择
type AlertAction int

const (
	ActionStartBreak AlertAction = iota // 用户选择"去休息"
	ActionPostpone                      // 用户选择"稍后提醒"
	ActionTimeout                       // 弹窗超时自动关闭
	ActionError                         // 发生错误
)

// SendReminder 发送系统通知提醒用户休息（通知中心横幅，不会打断全屏）
func SendReminder(title, message string) error {
	return beeep.Notify(title, message, "")
}