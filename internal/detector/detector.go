package detector

import "time"

// IdleDetector 空闲检测接口
type IdleDetector interface {
	// GetIdleDuration 返回用户自上次输入以来的空闲时长
	GetIdleDuration() time.Duration
}

// NewIdleDetector 创建当前平台对应的空闲检测器实例
// 具体实现由平台适配文件提供（detector_darwin.go / detector_windows.go）
func NewIdleDetector() IdleDetector {
	return newPlatformDetector()
}
