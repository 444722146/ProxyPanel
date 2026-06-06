package config

import (
	"os"
	"path/filepath"
)

// Config 应用配置结构体
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Nginx    NginxConfig
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string
	Mode string
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Path string
}

// NginxConfig Nginx配置
type NginxConfig struct {
	ConfigDir    string
	LogDir       string
	TemplatePath string
	ReloadCmd    string
}

// AppConfig 全局配置实例
var AppConfig Config

// InitConfig 初始化配置
func InitConfig() {
	// 获取程序运行目录
	execPath, _ := os.Executable()
	workDir := filepath.Dir(execPath)
	
	// 设置默认配置
	AppConfig = Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "5000"),
			Mode: getEnv("GIN_MODE", "release"),
		},
		Database: DatabaseConfig{
			Path: filepath.Join(workDir, "data", "proxypanel.db"),
		},
		Nginx: NginxConfig{
			ConfigDir:    "/etc/nginx/sites-enabled",
			LogDir:       "/var/log/nginx",
			TemplatePath: filepath.Join(workDir, "templates", "nginx.conf.tmpl"),
			ReloadCmd:    "nginx -s reload",
		},
	}
	
	// 确保数据目录存在
	dataDir := filepath.Dir(AppConfig.Database.Path)
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		os.MkdirAll(dataDir, 0755)
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}