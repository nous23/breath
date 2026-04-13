package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Config 应用配置结构体
type Config struct {
	mu sync.RWMutex `json:"-"`

	// 活跃时长阈值（分钟），默认 45，范围 1-120
	ActiveThresholdMin int `json:"active_threshold_min"`
	// 休息时长（分钟），默认 5，范围 1-30
	BreakDurationMin int `json:"break_duration_min"`
	// 空闲判定时长（分钟），默认 5，范围 1-15
	IdleThresholdMin int `json:"idle_threshold_min"`
	// 推迟提醒间隔（分钟），默认 5，范围 1-15
	PostponeIntervalMin int `json:"postpone_interval_min"`
	// 开机自启动
	AutoStartEnabled bool `json:"auto_start_enabled"`
}

// Default 返回默认配置
func Default() *Config {
	return &Config{
		ActiveThresholdMin:  45,
		BreakDurationMin:    5,
		IdleThresholdMin:    5,
		PostponeIntervalMin: 5,
		AutoStartEnabled:    false,
	}
}

// ActiveThreshold 返回活跃时长阈值的 time.Duration
func (c *Config) ActiveThreshold() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Duration(c.ActiveThresholdMin) * time.Minute
}

// BreakDuration 返回休息时长的 time.Duration
func (c *Config) BreakDuration() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Duration(c.BreakDurationMin) * time.Minute
}

// IdleThreshold 返回空闲判定时长的 time.Duration
func (c *Config) IdleThreshold() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Duration(c.IdleThresholdMin) * time.Minute
}

// PostponeInterval 返回推迟提醒间隔的 time.Duration
func (c *Config) PostponeInterval() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Duration(c.PostponeIntervalMin) * time.Minute
}

// Validate 校验配置项范围，超出范围则修正为默认值
func (c *Config) Validate() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ActiveThresholdMin < 1 || c.ActiveThresholdMin > 120 {
		c.ActiveThresholdMin = 45
	}
	if c.BreakDurationMin < 1 || c.BreakDurationMin > 30 {
		c.BreakDurationMin = 5
	}
	if c.IdleThresholdMin < 1 || c.IdleThresholdMin > 15 {
		c.IdleThresholdMin = 5
	}
	if c.PostponeIntervalMin < 1 || c.PostponeIntervalMin > 15 {
		c.PostponeIntervalMin = 5
	}
}

// Update 线程安全地更新配置
func (c *Config) Update(newCfg *Config) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ActiveThresholdMin = newCfg.ActiveThresholdMin
	c.BreakDurationMin = newCfg.BreakDurationMin
	c.IdleThresholdMin = newCfg.IdleThresholdMin
	c.PostponeIntervalMin = newCfg.PostponeIntervalMin
	c.AutoStartEnabled = newCfg.AutoStartEnabled
}

// Clone 返回配置的深拷贝
func (c *Config) Clone() *Config {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return &Config{
		ActiveThresholdMin:  c.ActiveThresholdMin,
		BreakDurationMin:    c.BreakDurationMin,
		IdleThresholdMin:    c.IdleThresholdMin,
		PostponeIntervalMin: c.PostponeIntervalMin,
		AutoStartEnabled:    c.AutoStartEnabled,
	}
}

// Load 从配置文件加载配置
func Load() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("获取配置路径失败: %w", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	cfg := Default()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	cfg.Validate()
	return cfg, nil
}

// Save 将配置保存到配置文件
func Save(cfg *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("获取配置路径失败: %w", err)
	}

	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	cfg.mu.RLock()
	data, err := json.MarshalIndent(cfg, "", "  ")
	cfg.mu.RUnlock()
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// getConfigPath 返回配置文件路径（由平台适配文件实现 getConfigDir）
func getConfigPath() (string, error) {
	dir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}
