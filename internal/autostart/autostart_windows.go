//go:build windows

package autostart

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows/registry"
)

const (
	registryKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`
	registryAppName = "Breath"
)

// windowsAutoStarter Windows 平台开机自启动管理器
type windowsAutoStarter struct{}

func newPlatformAutoStarter() AutoStarter {
	return &windowsAutoStarter{}
}

// Enable 通过注册表注册开机自启动
func (w *windowsAutoStarter) Enable() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %w", err)
	}

	key, _, err := registry.CreateKey(registry.CURRENT_USER, registryKeyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("打开注册表键失败: %w", err)
	}
	defer key.Close()

	if err := key.SetStringValue(registryAppName, exePath); err != nil {
		return fmt.Errorf("设置注册表值失败: %w", err)
	}

	return nil
}

// Disable 从注册表移除开机自启动
func (w *windowsAutoStarter) Disable() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, registryKeyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("打开注册表键失败: %w", err)
	}
	defer key.Close()

	if err := key.DeleteValue(registryAppName); err != nil {
		return fmt.Errorf("删除注册表值失败: %w", err)
	}

	return nil
}

// IsEnabled 检查是否已注册开机自启动
func (w *windowsAutoStarter) IsEnabled() bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, registryKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()

	_, _, err = key.GetStringValue(registryAppName)
	return err == nil
}
