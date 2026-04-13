//go:build windows

package notify

import (
	"syscall"
	"unsafe"
)

var (
	user32          = syscall.NewLazyDLL("user32.dll")
	procMessageBoxW = user32.NewProc("MessageBoxW")
)

const (
	// MessageBox 按钮类型
	mbYesNo = 0x00000004 // "是"和"否"两个按钮

	// MessageBox 图标
	mbIconWarning = 0x00000030 // 警告图标

	// MessageBox 模态标志
	mbSystemModal   = 0x00001000 // 系统模态，强制置顶
	mbTopmost       = 0x00040000 // 始终在最前面
	mbSetForeground = 0x00010000 // 将窗口设为前台

	// MessageBox 返回值
	idYes = 6 // 用户点击了"是"
	idNo  = 7 // 用户点击了"否"
)

// SendAlert 发送强制弹窗提醒（Windows: MessageBoxW + MB_SYSTEMMODAL，会打断全屏模式）
// 通过回调 onAction 异步通知调用方用户的选择
func SendAlert(title, message string, onAction func(AlertAction)) {
	go func() {
		action := messageBoxW(title, message)
		if onAction != nil {
			onAction(action)
		}
	}()
}

// messageBoxW 调用 Win32 API MessageBoxW 弹出系统级模态对话框
// 使用"是/否"按钮，"是"=去休息，"否"=稍后提醒
func messageBoxW(title, message string) AlertAction {
	// 在消息末尾追加操作提示
	fullMessage := message + "\n\n点击「是」开始休息，点击「否」稍后提醒"

	titlePtr, _ := syscall.UTF16PtrFromString(title)
	messagePtr, _ := syscall.UTF16PtrFromString(fullMessage)

	flags := mbYesNo | mbIconWarning | mbSystemModal | mbTopmost | mbSetForeground

	ret, _, _ := procMessageBoxW.Call(
		0, // hWnd = NULL（桌面窗口）
		uintptr(unsafe.Pointer(messagePtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(flags),
	)

	switch int(ret) {
	case idYes:
		return ActionStartBreak
	case idNo:
		return ActionPostpone
	default:
		return ActionPostpone
	}
}