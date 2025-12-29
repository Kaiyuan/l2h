package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	ServerA ServerAConfig `json:"server_a,omitempty"`
	ServerB ServerBConfig `json:"server_b,omitempty"`
	Logging LoggingConfig `json:"logging,omitempty"`
}

type ServerAConfig struct {
	Port      int    `json:"port"`
	DBPath    string `json:"db_path"`
	LogFile   string `json:"log_file,omitempty"`
	LogLevel  string `json:"log_level,omitempty"`
}

type ServerBConfig struct {
	Port      int    `json:"port"`
	DBPath    string `json:"db_path"`
	LogFile   string `json:"log_file,omitempty"`
	LogLevel  string `json:"log_level,omitempty"`
}

type LoggingConfig struct {
	Level  string `json:"level,omitempty"`  // DEBUG, INFO, WARN, ERROR, FATAL
	File   string `json:"file,omitempty"`
	Stdout bool   `json:"stdout,omitempty"`
}

// Load 从文件加载配置
func Load(configPath string) (*Config, error) {
	if configPath == "" {
		return Default(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 如果文件不存在，创建默认配置
			cfg := Default()
			if err := Save(configPath, cfg); err != nil {
				return nil, fmt.Errorf("创建默认配置失败: %w", err)
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &cfg, nil
}

// Save 保存配置到文件
func Save(configPath string, cfg *Config) error {
	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// Default 返回默认配置
func Default() *Config {
	return &Config{
		ServerA: ServerAConfig{
			Port:     8080,
			DBPath:   "l2h-s.db",
			LogFile:  "logs/l2h-s.log",
			LogLevel: "INFO",
		},
		ServerB: ServerBConfig{
			Port:     55055,
			DBPath:   "l2h-c.db",
			LogFile:  "logs/l2h-c.log",
			LogLevel: "INFO",
		},
		Logging: LoggingConfig{
			Level:  "INFO",
			Stdout: true,
		},
	}
}

// GetLogLevel 将字符串转换为日志级别
func GetLogLevel(level string) int {
	switch level {
	case "DEBUG":
		return 0
	case "INFO":
		return 1
	case "WARN":
		return 2
	case "ERROR":
		return 3
	case "FATAL":
		return 4
	default:
		return 1 // 默认 INFO
	}
}

