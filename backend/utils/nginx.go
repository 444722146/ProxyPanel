package utils

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"proxypanel/config"
	"proxypanel/models"
)

// NginxConfigData Nginx配置模板数据
type NginxConfigData struct {
	Domain     string
	Port       int
	TargetURL  string
	TargetHost string
	Token      string
	FakeIP     string
	Whitelist  []string
	SSLEnabled bool
	SSLCert    string
	SSLKey     string
}

// GenerateNginxConfig 生成Nginx配置文件
func GenerateNginxConfig(rule *models.ProxyRule) error {
	// 准备模板数据
	data := NginxConfigData{
		Domain:     rule.Domain,
		Port:       rule.Port,
		TargetURL:  rule.TargetURL,
		Token:      rule.Token,
		FakeIP:     rule.FakeIP,
		SSLEnabled: rule.SSLEnabled,
		SSLCert:    rule.SSLCert,
		SSLKey:     rule.SSLKey,
	}

	// 从目标地址中提取 Host（去掉协议）
	if parsedURL, err := url.Parse(rule.TargetURL); err == nil && parsedURL.Host != "" {
		data.TargetHost = parsedURL.Host
	} else {
		// 如果解析失败，兜底处理：去掉协议前缀
		data.TargetHost = strings.TrimPrefix(strings.TrimPrefix(rule.TargetURL, "http://"), "https://")
		if idx := strings.Index(data.TargetHost, "/"); idx != -1 {
			data.TargetHost = data.TargetHost[:idx]
		}
	}

	// 解析IP白名单
	if rule.Whitelist != "" {
		data.Whitelist = strings.Split(rule.Whitelist, ",")
		for i := range data.Whitelist {
			data.Whitelist[i] = strings.TrimSpace(data.Whitelist[i])
		}
	}

	// 读取模板文件
	tmplPath := config.AppConfig.Nginx.TemplatePath
	tmplContent, err := os.ReadFile(tmplPath)
	if err != nil {
		return fmt.Errorf("读取模板文件失败: %v", err)
	}

	// 解析模板
	tmpl, err := template.New("nginx").Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("解析模板失败: %v", err)
	}

	// 执行模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("执行模板失败: %v", err)
	}

	// 写入配置文件
	configPath := filepath.Join(config.AppConfig.Nginx.ConfigDir, fmt.Sprintf("proxy_%s.conf", SanitizeFilename(rule.Domain)))
	if err := os.WriteFile(configPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// DeleteNginxConfig 删除Nginx配置文件
func DeleteNginxConfig(domain string) error {
	configPath := filepath.Join(config.AppConfig.Nginx.ConfigDir, fmt.Sprintf("proxy_%s.conf", SanitizeFilename(domain)))
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil // 文件不存在，无需删除
	}
	return os.Remove(configPath)
}

// ReloadNginx 重载Nginx配置
func ReloadNginx() error {
	// 先测试配置是否正确
	testCmd := exec.Command("nginx", "-t")
	if output, err := testCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("Nginx配置测试失败: %v\n%s", err, string(output))
	}

	// 重载配置
	reloadCmd := exec.Command("nginx", "-s", "reload")
	if output, err := reloadCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("Nginx重载失败: %v\n%s", err, string(output))
	}

	return nil
}

// TestNginxConfig 测试Nginx配置
func TestNginxConfig() (bool, string) {
	cmd := exec.Command("nginx", "-t")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, string(output)
	}
	return true, string(output)
}

// SyncAllProxyConfigs 同步所有代理规则到Nginx配置
func SyncAllProxyConfigs() error {
	// 获取所有启用的代理规则
	rules, err := models.GetEnabledProxyRules()
	if err != nil {
		return fmt.Errorf("获取代理规则失败: %v", err)
	}

	// 清空现有配置文件（保留备份）
	configDir := config.AppConfig.Nginx.ConfigDir
	files, err := filepath.Glob(filepath.Join(configDir, "proxy_*.conf"))
	if err != nil {
		return fmt.Errorf("扫描配置文件失败: %v", err)
	}

	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return fmt.Errorf("删除旧配置失败: %v", err)
		}
	}

	// 生成新的配置文件
	for _, rule := range rules {
		if err := GenerateNginxConfig(&rule); err != nil {
			return fmt.Errorf("生成配置失败 [域名: %s]: %v", rule.Domain, err)
		}
	}

	// 重载Nginx
	return ReloadNginx()
}

// sanitizeFilename 清理文件名中的特殊字符
func SanitizeFilename(name string) string {
	// 替换特殊字符
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(name)
}