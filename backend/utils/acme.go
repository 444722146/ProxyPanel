package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"proxypanel/models"
)

const (
	acmeShDir      = "/root/.acme.sh"
	acmeShPath     = "/root/.acme.sh/acme.sh"
	acmeWebRoot    = "/opt/proxypanel/ssl/acme"
	acmeCertDir    = "/opt/proxypanel/ssl/certs"
)

// EnsureAcmeSh 确保 acme.sh 已安装
func EnsureAcmeSh() error {
	if _, err := os.Stat(acmeShPath); err == nil {
		return nil // 已安装
	}

	// 下载并安装 acme.sh
	cmd := exec.Command("bash", "-c",
		"curl https://get.acme.sh | sh -s email=proxypanel@localhost",
	)
	cmd.Env = append(os.Environ(), "HOME=/root")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("安装 acme.sh 失败: %v\n%s", err, string(output))
	}

	return nil
}

// EnsureAcmeDirs 确保 ACME 所需目录存在
func EnsureAcmeDirs() error {
	dirs := []string{acmeWebRoot, acmeCertDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录 %s 失败: %v", dir, err)
		}
		// 确保 acme-challenge 子目录存在
		challengeDir := filepath.Join(dir, ".well-known", "acme-challenge")
		if err := os.MkdirAll(challengeDir, 0755); err != nil {
			return fmt.Errorf("创建挑战目录 %s 失败: %v", challengeDir, err)
		}
	}
	return nil
}

// isIPAddress 检查域名是否为 IP 地址
func isIPAddress(domain string) bool {
	// 简单检查：IPv4 格式
	parts := strings.Split(domain, ".")
	if len(parts) == 4 {
		for _, part := range parts {
			if _, err := fmt.Sscanf(part, "%d", new(int)); err != nil {
				return false
			}
		}
		return true
	}
	// IPv6 简单检查
	if strings.Contains(domain, ":") {
		return true
	}
	return false
}

// RequestFreeCert 申请 SSL 证书
func RequestFreeCert(rule *models.ProxyRule, ca string) error {
	// 确保工具和目录就绪
	if err := EnsureAcmeSh(); err != nil {
		return err
	}
	if err := EnsureAcmeDirs(); err != nil {
		return err
	}

	domain := rule.Domain

	// IP 类型域名只能使用 HTTP-01 文件验证
	forceHTTP := isIPAddress(domain)

	// 构造 acme.sh 命令
	var args []string
	args = append(args, acmeShPath, "--issue")

	if forceHTTP {
		// HTTP-01 文件验证
		args = append(args, "-d", domain, "--webroot", acmeWebRoot)
	} else {
		// 默认也用 HTTP-01，更简单可靠
		args = append(args, "-d", domain, "--webroot", acmeWebRoot)
	}

	// 选择 CA
	switch ca {
	case "zerossl":
		args = append(args, "--server", "zerossl")
	case "letsencrypt":
		args = append(args, "--server", "letsencrypt")
	default:
		args = append(args, "--server", "letsencrypt")
	}

	// 强制重新签发（如果已有证书）
	args = append(args, "--force")

	// 执行申请
	cmd := exec.Command("bash", "-c", strings.Join(args, " "))
	cmd.Env = append(os.Environ(), "HOME=/root")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("申请证书失败: %v\n%s", err, string(output))
	}

	// 安装证书到指定目录
	certDir := filepath.Join(acmeCertDir, SanitizeFilename(domain))
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("创建证书目录失败: %v", err)
	}

	certPath := filepath.Join(certDir, "fullchain.cer")
	keyPath := filepath.Join(certDir, domain+".key")

	installArgs := []string{
		acmeShPath, "--install-cert", "-d", domain,
		"--fullchain-file", certPath,
		"--key-file", keyPath,
		"--reloadcmd", "nginx -s reload",
	}

	installCmd := exec.Command("bash", "-c", strings.Join(installArgs, " "))
	installCmd.Env = append(os.Environ(), "HOME=/root")
	installOutput, err := installCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("安装证书失败: %v\n%s", err, string(installOutput))
	}

	// 更新数据库
	rule.SSLEnabled = true
	rule.SSLType = ca
	rule.SSLCert = certPath
	rule.SSLKey = keyPath
	rule.SSLAutoRenew = true // 默认开启自动续签

	// 解析证书过期时间
	expiresAt := parseCertExpiry(certPath)
	rule.SSLExpiresAt = expiresAt

	if err := models.UpdateProxyRule(rule); err != nil {
		return fmt.Errorf("更新规则失败: %v", err)
	}

	// 重新生成 Nginx 配置
	if rule.Enabled {
		if err := GenerateNginxConfig(rule); err != nil {
			return fmt.Errorf("生成 Nginx 配置失败: %v", err)
		}
		if err := ReloadNginx(); err != nil {
			return fmt.Errorf("重载 Nginx 失败: %v", err)
		}
	}

	return nil
}

// RenewCert 续签 SSL 证书
func RenewCert(rule *models.ProxyRule) error {
	if rule.SSLType == "" || rule.SSLType == "manual" || rule.SSLType == "selfsigned" {
		return fmt.Errorf("该证书类型不支持续签")
	}

	domain := rule.Domain

	args := []string{acmeShPath, "--renew", "-d", domain, "--force"}

	// 选择 CA
	switch rule.SSLType {
	case "zerossl":
		args = append(args, "--server", "zerossl")
	case "letsencrypt":
		args = append(args, "--server", "letsencrypt")
	}

	cmd := exec.Command("bash", "-c", strings.Join(args, " "))
	cmd.Env = append(os.Environ(), "HOME=/root")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("续签证书失败: %v\n%s", err, string(output))
	}

	// 更新过期时间
	certDir := filepath.Join(acmeCertDir, SanitizeFilename(domain))
	certPath := filepath.Join(certDir, "fullchain.cer")
	expiresAt := parseCertExpiry(certPath)
	rule.SSLExpiresAt = expiresAt

	if err := models.UpdateProxyRule(rule); err != nil {
		return fmt.Errorf("更新规则失败: %v", err)
	}

	// 重载 Nginx
	if rule.Enabled {
		if err := ReloadNginx(); err != nil {
			return fmt.Errorf("重载 Nginx 失败: %v", err)
		}
	}

	return nil
}

// SetupAutoRenewCron 设置 acme.sh 自动续签 cron
func SetupAutoRenewCron() error {
	// acme.sh 安装时会自动添加 cron，这里确保它存在
	cmd := exec.Command("bash", "-c",
		fmt.Sprintf("(crontab -l 2>/dev/null | grep -v '%s'; echo '0 0 * * * %s --cron --home %s') | crontab -",
			acmeShPath, acmeShPath, acmeShDir),
	)
	cmd.Env = append(os.Environ(), "HOME=/root")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("设置自动续签 cron 失败: %v\n%s", err, string(output))
	}
	return nil
}

// parseCertExpiry 解析证书过期时间
func parseCertExpiry(certPath string) *time.Time {
	// 使用 openssl 命令解析证书过期时间
	cmd := exec.Command("openssl", "x509", "-enddate", "-noout", "-in", certPath)
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	// 输出格式：notAfter=Mar 10 00:00:00 2026 GMT
	line := strings.TrimSpace(string(output))
	line = strings.TrimPrefix(line, "notAfter=")

	t, err := time.Parse("Jan 2 15:04:05 2006 MST", line)
	if err != nil {
		return nil
	}

	return &t
}

// GetCertInfo 获取证书详细信息
func GetCertInfo(certPath string) map[string]interface{} {
	info := make(map[string]interface{})

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		info["exists"] = false
		return info
	}
	info["exists"] = true

	// 过期时间
	cmd := exec.Command("openssl", "x509", "-enddate", "-noout", "-in", certPath)
	output, err := cmd.Output()
	if err == nil {
		line := strings.TrimPrefix(strings.TrimSpace(string(output)), "notAfter=")
		if t, err := time.Parse("Jan 2 15:04:05 2006 MST", line); err == nil {
			info["expires_at"] = t.Format("2006-01-02 15:04:05")
			daysLeft := int(time.Until(t).Hours() / 24)
			info["days_left"] = daysLeft
			info["expired"] = daysLeft < 0
		}
	}

	// 域名/主题
	cmd = exec.Command("openssl", "x509", "-subject", "-noout", "-in", certPath)
	output, err = cmd.Output()
	if err == nil {
		info["subject"] = strings.TrimSpace(string(output))
	}

	// 颁发者
	cmd = exec.Command("openssl", "x509", "-issuer", "-noout", "-in", certPath)
	output, err = cmd.Output()
	if err == nil {
		info["issuer"] = strings.TrimSpace(string(output))
	}

	return info
}
