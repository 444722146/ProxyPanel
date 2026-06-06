package controllers

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"proxypanel/models"
	"proxypanel/utils"
)

// RequestFreeCertRequest 申请证书请求
type RequestFreeCertRequest struct {
	CA string `json:"ca"` // letsencrypt / zerossl
}

// RequestFreeCert 申请 SSL 证书
func RequestFreeCert(c *gin.Context) {
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

	// IP 地址不支持证书申请
	if isIPDomain(rule.Domain) {
		utils.BadRequest(c, "IP地址类型不支持证书申请，请使用上传证书或自签名证书")
		return
	}

	// 解析请求
	var req RequestFreeCertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 默认使用 Let's Encrypt
		req.CA = "letsencrypt"
	}

	// 验证 CA 类型
	if req.CA != "letsencrypt" && req.CA != "zerossl" {
		req.CA = "letsencrypt"
	}

	// 确保 acme 所需目录存在
	if err := utils.EnsureAcmeDirs(); err != nil {
		utils.InternalError(c, "初始化ACME目录失败: "+err.Error())
		return
	}

	// 申请证书
	if err := utils.RequestFreeCert(rule, req.CA); err != nil {
		utils.InternalError(c, "申请证书失败: "+err.Error())
		return
	}

	// 设置自动续签 cron
	_ = utils.SetupAutoRenewCron()

	utils.SuccessWithMsg(c, "SSL证书申请成功", gin.H{
		"domain":       rule.Domain,
		"ssl_type":     rule.SSLType,
		"ssl_cert":     rule.SSLCert,
		"ssl_key":      rule.SSLKey,
		"auto_renew":   rule.SSLAutoRenew,
		"expires_at":   rule.SSLExpiresAt,
	})
}

// RenewCertificate 手动续签证书
func RenewCertificate(c *gin.Context) {
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

	if !rule.SSLEnabled || (rule.SSLType != "letsencrypt" && rule.SSLType != "zerossl") {
		utils.BadRequest(c, "该代理规则没有可续签的证书")
		return
	}

	// 续签
	if err := utils.RenewCert(rule); err != nil {
		utils.InternalError(c, "续签证书失败: "+err.Error())
		return
	}

	utils.SuccessWithMsg(c, "证书续签成功", gin.H{
		"domain":     rule.Domain,
		"expires_at": rule.SSLExpiresAt,
	})
}

// GetCertStatus 获取证书状态
func GetCertStatus(c *gin.Context) {
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

	result := gin.H{
		"domain":       rule.Domain,
		"ssl_enabled":  rule.SSLEnabled,
		"ssl_type":     rule.SSLType,
		"auto_renew":   rule.SSLAutoRenew,
		"expires_at":   rule.SSLExpiresAt,
		"is_ip_domain": isIPDomain(rule.Domain),
	}

	// 如果有证书路径，获取详细信息
	if rule.SSLCert != "" {
		certInfo := utils.GetCertInfo(rule.SSLCert)
		result["cert_info"] = certInfo
	}

	utils.Success(c, result)
}

// ToggleAutoRenew 切换自动续签
func ToggleAutoRenew(c *gin.Context) {
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

	if !rule.SSLEnabled || (rule.SSLType != "letsencrypt" && rule.SSLType != "zerossl") {
		utils.BadRequest(c, "该代理规则没有自动续签类型的证书")
		return
	}

	// 切换
	rule.SSLAutoRenew = !rule.SSLAutoRenew
	if err := models.UpdateProxyRule(rule); err != nil {
		utils.InternalError(c, "更新失败: "+err.Error())
		return
	}

	// 如果开启自动续签，确保 cron 存在
	if rule.SSLAutoRenew {
		_ = utils.SetupAutoRenewCron()
	}

	utils.SuccessWithMsg(c, "自动续签已更新", gin.H{
		"auto_renew": rule.SSLAutoRenew,
	})
}

// isIPDomain 判断域名是否为 IP 地址
func isIPDomain(domain string) bool {
	parts := strings.Split(domain, ".")
	if len(parts) == 4 {
		for _, p := range parts {
			for _, ch := range p {
				if ch < '0' || ch > '9' {
					return false
				}
			}
		}
		return true
	}
	return strings.Contains(domain, ":")
}
