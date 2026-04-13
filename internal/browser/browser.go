package browser

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Open 在默认浏览器中打开指定 URL
func Open(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	return cmd.Start()
}
