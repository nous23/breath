//go:build darwin

package autostart

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

const plistLabel = "com.breath.app"

// plist 模板
const plistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{{.Label}}</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{.ExePath}}</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <false/>
</dict>
</plist>`

type plistData struct {
	Label   string
	ExePath string
}

// darwinAutoStarter macOS 平台开机自启动管理器
type darwinAutoStarter struct{}

func newPlatformAutoStarter() AutoStarter {
	return &darwinAutoStarter{}
}

// getPlistPath 获取 plist 文件路径
func getPlistPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents", plistLabel+".plist"), nil
}

// Enable 通过 LaunchAgent 注册开机自启动
func (d *darwinAutoStarter) Enable() error {
	plistPath, err := getPlistPath()
	if err != nil {
		return fmt.Errorf("获取 plist 路径失败: %w", err)
	}

	// 获取当前可执行文件路径
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %w", err)
	}

	// 确保目录存在
	dir := filepath.Dir(plistPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建 LaunchAgents 目录失败: %w", err)
	}

	// 生成 plist 文件
	f, err := os.Create(plistPath)
	if err != nil {
		return fmt.Errorf("创建 plist 文件失败: %w", err)
	}
	defer f.Close()

	tmpl, err := template.New("plist").Parse(plistTemplate)
	if err != nil {
		return fmt.Errorf("解析模板失败: %w", err)
	}

	data := plistData{
		Label:   plistLabel,
		ExePath: exePath,
	}

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("写入 plist 文件失败: %w", err)
	}

	return nil
}

// Disable 移除 LaunchAgent 注册
func (d *darwinAutoStarter) Disable() error {
	plistPath, err := getPlistPath()
	if err != nil {
		return fmt.Errorf("获取 plist 路径失败: %w", err)
	}

	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除 plist 文件失败: %w", err)
	}
	return nil
}

// IsEnabled 检查是否已注册开机自启动
func (d *darwinAutoStarter) IsEnabled() bool {
	plistPath, err := getPlistPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(plistPath)
	return err == nil
}
