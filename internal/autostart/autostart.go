package autostart

// AutoStarter 开机自启动接口
type AutoStarter interface {
	// Enable 注册开机自启动
	Enable() error
	// Disable 取消开机自启动
	Disable() error
	// IsEnabled 检查是否已启用开机自启动
	IsEnabled() bool
}

// New 创建当前平台对应的开机自启动管理器
func New() AutoStarter {
	return newPlatformAutoStarter()
}
