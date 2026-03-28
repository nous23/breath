//go:build darwin

package config

import (
	"os"
	"path/filepath"
)

// getConfigDir 返回 macOS 平台的配置目录路径
func getConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Application Support", "Breath"), nil
}
