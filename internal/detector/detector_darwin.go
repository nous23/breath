//go:build darwin

package detector

/*
#cgo LDFLAGS: -framework CoreGraphics
#include <CoreGraphics/CGEventSource.h>

double getIdleSeconds() {
    return CGEventSourceSecondsSinceLastEventType(
        kCGEventSourceStateCombinedSessionState,
        kCGAnyInputEventType
    );
}
*/
import "C"
import "time"

// darwinDetector macOS 平台空闲检测器
type darwinDetector struct{}

func newPlatformDetector() IdleDetector {
	return &darwinDetector{}
}

// GetIdleDuration 通过 CGEventSource 获取用户空闲时长
func (d *darwinDetector) GetIdleDuration() time.Duration {
	seconds := float64(C.getIdleSeconds())
	return time.Duration(seconds * float64(time.Second))
}
