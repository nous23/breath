//go:build windows

package detector

import (
	"syscall"
	"time"
	"unsafe"
)

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	getLastInputInfo = user32.NewProc("GetLastInputInfo")
	getTickCount     = kernel32.NewProc("GetTickCount")
)

// LASTINPUTINFO Win32 结构体
type lastInputInfo struct {
	cbSize uint32
	dwTime uint32
}

// windowsDetector Windows 平台空闲检测器
type windowsDetector struct{}

func newPlatformDetector() IdleDetector {
	return &windowsDetector{}
}

// GetIdleDuration 通过 GetLastInputInfo Win32 API 获取用户空闲时长
func (d *windowsDetector) GetIdleDuration() time.Duration {
	var lii lastInputInfo
	lii.cbSize = uint32(unsafe.Sizeof(lii))

	ret, _, _ := getLastInputInfo.Call(uintptr(unsafe.Pointer(&lii)))
	if ret == 0 {
		return 0
	}

	tickCount, _, _ := getTickCount.Call()
	idleMs := uint32(tickCount) - lii.dwTime
	return time.Duration(idleMs) * time.Millisecond
}
