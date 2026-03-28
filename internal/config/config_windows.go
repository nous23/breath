//go:build windows

package config

import (
	"os"
	"path/filepath"
)

// getConfigDir 返回 Windows 平台的配置目录路径
func getConfigDir() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		appData = filepath.Join(home, "AppData", "Roaming")
	}
	return filepath.Join(appData, "Breath"), nil
}
