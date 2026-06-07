package controllers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"proxypanel/config"
	"proxypanel/models"
	"proxypanel/utils"
)

// UploadSSLCertificate 上传SSL证书
func UploadSSLCertificate(c *gin.Context) {
	id := c.Param("id")
	
	ruleID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	// 获取规则
	rule, err := models.GetProxyRuleByID(uint(ruleID))
	if err != nil {
		utils.NotFound(c, "代理规则不存在")
		return
	}

	// 获取上传的证书文件
	certFile, err := c.FormFile("cert")
	if err != nil {
		utils.BadRequest(c, "证书文件上传失败")
		return
	}

	keyFile, err := c.FormFile("key")
	if err != nil {
		utils.BadRequest(c, "私钥文件上传失败")
		return
	}

	// 创建SSL证书目录
	sslDir := filepath.Join(filepath.Dir(config.AppConfig.Database.Path), "ssl")
	if _, err := os.Stat(sslDir); os.IsNotExist(err) {
		os.MkdirAll(sslDir, 0755)
	}

	// 保存证书文件（文件名含端口，避免同域名不同端口的证书互相覆盖）
	portSuffix := rule.Port
	if portSuffix == 0 {
		portSuffix = 80
	}
	certPath := filepath.Join(sslDir, fmt.Sprintf("%s_%d.crt", utils.SanitizeFilename(rule.Domain), portSuffix))
	keyPath := filepath.Join(sslDir, fmt.Sprintf("%s_%d.key", utils.SanitizeFilename(rule.Domain), portSuffix))

	if err := c.SaveUploadedFile(certFile, certPath); err != nil {
		utils.InternalError(c, "保存证书文件失败: "+err.Error())
		return
	}

	if err := c.SaveUploadedFile(keyFile, keyPath); err != nil {
		os.Remove(certPath)
		utils.InternalError(c, "保存私钥文件失败: "+err.Error())
		return
	}

	// 验证证书和私钥是否匹配
	if err := validateCertificate(certPath, keyPath); err != nil {
		os.Remove(certPath)
		os.Remove(keyPath)
		utils.BadRequest(c, "证书验证失败: "+err.Error())
		return
	}

	// 更新规则
	rule.SSLEnabled = true
	rule.SSLType = "manual"
	rule.SSLCert = certPath
	rule.SSLKey = keyPath
	rule.SSLAutoRenew = false

	if err := models.UpdateProxyRule(rule); err != nil {
		os.Remove(certPath)
		os.Remove(keyPath)
		utils.InternalError(c, "更新规则失败: "+err.Error())
		return
	}

	// 重新生成Nginx配置
	if rule.Enabled {
		if err := utils.GenerateNginxConfig(rule); err != nil {
			utils.ErrorWithData(c, 500, "证书上传成功但生成配置失败", gin.H{
				"rule": rule,
				"nginx_error": err.Error(),
			})
			return
		}
		
		// 重载Nginx
		if err := utils.ReloadNginx(); err != nil {
			utils.ErrorWithData(c, 500, "证书上传成功但重载Nginx失败", gin.H{
				"rule": rule,
				"nginx_error": err.Error(),
			})
			return
		}
	}

	utils.SuccessWithMsg(c, "SSL证书上传成功", rule)
}

// GenerateSelfSignedCertificate 生成自签名证书
func GenerateSelfSignedCertificate(c *gin.Context) {
	id := c.Param("id")
	
	ruleID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	// 获取规则
	rule, err := models.GetProxyRuleByID(uint(ruleID))
	if err != nil {
		utils.NotFound(c, "代理规则不存在")
		return
	}

	// 创建SSL证书目录
	sslDir := filepath.Join(filepath.Dir(config.AppConfig.Database.Path), "ssl")
	if _, err := os.Stat(sslDir); os.IsNotExist(err) {
		os.MkdirAll(sslDir, 0755)
	}

	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		utils.InternalError(c, "生成私钥失败: "+err.Error())
		return
	}

	// 创建证书模板
	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName:   rule.Domain,
			Organization: []string{"ProxyPanel"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0), // 1年有效期
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{rule.Domain},
	}

	// 生成证书
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		utils.InternalError(c, "生成证书失败: "+err.Error())
		return
	}

	// 保存私钥（文件名含端口，避免同域名不同端口的证书互相覆盖）
	portSuffix := rule.Port
	if portSuffix == 0 {
		portSuffix = 80
	}
	keyPath := filepath.Join(sslDir, fmt.Sprintf("%s_%d.key", utils.SanitizeFilename(rule.Domain), portSuffix))
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err := os.WriteFile(keyPath, keyPEM, 0644); err != nil {
		utils.InternalError(c, "保存私钥失败: "+err.Error())
		return
	}

	// 保存证书
	certPath := filepath.Join(sslDir, fmt.Sprintf("%s_%d.crt", utils.SanitizeFilename(rule.Domain), portSuffix))
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})
	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		os.Remove(keyPath)
		utils.InternalError(c, "保存证书失败: "+err.Error())
		return
	}

	// 更新规则
	rule.SSLEnabled = true
	rule.SSLType = "selfsigned"
	rule.SSLCert = certPath
	rule.SSLKey = keyPath
	rule.SSLAutoRenew = false

	if err := models.UpdateProxyRule(rule); err != nil {
		os.Remove(certPath)
		os.Remove(keyPath)
		utils.InternalError(c, "更新规则失败: "+err.Error())
		return
	}

	// 重新生成Nginx配置
	if rule.Enabled {
		if err := utils.GenerateNginxConfig(rule); err != nil {
			utils.ErrorWithData(c, 500, "证书生成成功但生成配置失败", gin.H{
				"rule": rule,
				"nginx_error": err.Error(),
			})
			return
		}
		
		// 重载Nginx
		if err := utils.ReloadNginx(); err != nil {
			utils.ErrorWithData(c, 500, "证书生成成功但重载Nginx失败", gin.H{
				"rule": rule,
				"nginx_error": err.Error(),
			})
			return
		}
	}

	utils.SuccessWithMsg(c, "自签名证书生成成功（仅用于测试，浏览器会提示不安全）", rule)
}

// RemoveSSLCertificate 移除SSL证书
func RemoveSSLCertificate(c *gin.Context) {
	id := c.Param("id")
	
	ruleID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	// 获取规则
	rule, err := models.GetProxyRuleByID(uint(ruleID))
	if err != nil {
		utils.NotFound(c, "代理规则不存在")
		return
	}

	// 删除证书文件
	if rule.SSLCert != "" {
		os.Remove(rule.SSLCert)
	}
	if rule.SSLKey != "" {
		os.Remove(rule.SSLKey)
	}

	// 更新规则
	rule.SSLEnabled = false
	rule.SSLType = ""
	rule.SSLCert = ""
	rule.SSLKey = ""
	rule.SSLAutoRenew = false
	rule.SSLExpiresAt = nil

	if err := models.UpdateProxyRule(rule); err != nil {
		utils.InternalError(c, "更新规则失败: "+err.Error())
		return
	}

	// 重新生成Nginx配置
	if rule.Enabled {
		if err := utils.GenerateNginxConfig(rule); err != nil {
			utils.ErrorWithData(c, 500, "证书移除成功但生成配置失败", gin.H{
				"rule": rule,
				"nginx_error": err.Error(),
			})
			return
		}
		
		// 重载Nginx
		if err := utils.ReloadNginx(); err != nil {
			utils.ErrorWithData(c, 500, "证书移除成功但重载Nginx失败", gin.H{
				"rule": rule,
				"nginx_error": err.Error(),
			})
			return
		}
	}

	utils.SuccessWithMsg(c, "SSL证书已移除", rule)
}

// validateCertificate 验证证书和私钥是否匹配
func validateCertificate(certPath string, keyPath string) error {
	// 读取证书
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("读取证书失败: %v", err)
	}

	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return fmt.Errorf("解析证书失败")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("解析证书失败: %v", err)
	}

	// 读取私钥
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("读取私钥失败: %v", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return fmt.Errorf("解析私钥失败")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		// 尝试PKCS8格式
		key, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
		if err != nil {
			return fmt.Errorf("解析私钥失败: %v", err)
		}
		privateKey = key.(*rsa.PrivateKey)
	}

	// 验证公钥匹配
	certPubKey := cert.PublicKey.(*rsa.PublicKey)
	if certPubKey.N.Cmp(privateKey.N) != 0 || certPubKey.E != privateKey.E {
		return fmt.Errorf("证书和私钥不匹配")
	}

	return nil
}