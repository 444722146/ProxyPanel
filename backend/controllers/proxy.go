package controllers

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"proxypanel/models"
	"proxypanel/utils"
)

// GetProxyRules 获取所有代理规则
func GetProxyRules(c *gin.Context) {
	rules, err := models.GetAllProxyRules()
	if err != nil {
		utils.InternalError(c, "获取代理规则失败: "+err.Error())
		return
	}
	utils.Success(c, rules)
}

// GetProxyRule 获取单个代理规则
func GetProxyRule(c *gin.Context) {
	id := c.Param("id")
	
	ruleID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	rule, err := models.GetProxyRuleByID(uint(ruleID))
	if err != nil {
		utils.NotFound(c, "代理规则不存在")
		return
	}
	
	utils.Success(c, rule)
}

// CreateProxyRule 创建代理规则
func CreateProxyRule(c *gin.Context) {
	var rule models.ProxyRule
	
	if err := c.ShouldBindJSON(&rule); err != nil {
		utils.BadRequest(c, "参数解析失败: "+err.Error())
		return
	}

	// 验证必填字段
	if rule.Name == "" || rule.Domain == "" || rule.TargetURL == "" {
		utils.BadRequest(c, "名称、域名和目标地址为必填项")
		return
	}

	// 检查域名是否已存在
	existing, err := models.GetProxyRuleByDomain(rule.Domain)
	if err != nil {
		utils.InternalError(c, "检查域名失败: "+err.Error())
		return
	}
	if existing != nil {
		utils.BadRequest(c, "该域名已被使用")
		return
	}

	// 创建规则
	if err := models.CreateProxyRule(&rule); err != nil {
		utils.InternalError(c, "创建代理规则失败: "+err.Error())
		return
	}

	// 如果规则已启用，生成Nginx配置
	if rule.Enabled {
		if err := utils.GenerateNginxConfig(&rule); err != nil {
			utils.ErrorWithData(c, 500, "创建成功但生成Nginx配置失败", gin.H{
				"rule": rule,
				"nginx_error": err.Error(),
			})
			return
		}
		
		// 重载Nginx
		if err := utils.ReloadNginx(); err != nil {
			utils.ErrorWithData(c, 500, "创建成功但重载Nginx失败", gin.H{
				"rule": rule,
				"nginx_error": err.Error(),
			})
			return
		}
	}

	utils.SuccessWithMsg(c, "代理规则创建成功", rule)
}

// UpdateProxyRule 更新代理规则
func UpdateProxyRule(c *gin.Context) {
	id := c.Param("id")
	
	ruleID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	// 获取现有规则
	existingRule, err := models.GetProxyRuleByID(uint(ruleID))
	if err != nil {
		utils.NotFound(c, "代理规则不存在")
		return
	}

	// 解析更新数据
	var updateData models.ProxyRule
	if err := c.ShouldBindJSON(&updateData); err != nil {
		utils.BadRequest(c, "参数解析失败: "+err.Error())
		return
	}

	// 检查域名是否与其他规则冲突
	if updateData.Domain != existingRule.Domain {
		conflictRule, err := models.GetProxyRuleByDomain(updateData.Domain)
		if err != nil {
			utils.InternalError(c, "检查域名失败: "+err.Error())
			return
		}
		if conflictRule != nil && conflictRule.ID != existingRule.ID {
			utils.BadRequest(c, "该域名已被其他规则使用")
			return
		}
	}

	// 删除旧的Nginx配置
	utils.DeleteNginxConfig(existingRule.Domain)

	// 更新规则
	existingRule.Name = updateData.Name
	existingRule.Domain = updateData.Domain
	existingRule.Port = updateData.Port
	existingRule.TargetURL = updateData.TargetURL
	existingRule.FakeIP = updateData.FakeIP
	existingRule.Token = updateData.Token
	existingRule.Enabled = updateData.Enabled
	existingRule.Whitelist = updateData.Whitelist
	existingRule.SSLEnabled = updateData.SSLEnabled
	existingRule.SSLCert = updateData.SSLCert
	existingRule.SSLKey = updateData.SSLKey

	if err := models.UpdateProxyRule(existingRule); err != nil {
		utils.InternalError(c, "更新代理规则失败: "+err.Error())
		return
	}

	// 如果规则已启用，生成新的Nginx配置
	if existingRule.Enabled {
		if err := utils.GenerateNginxConfig(existingRule); err != nil {
			utils.ErrorWithData(c, 500, "更新成功但生成Nginx配置失败", gin.H{
				"rule": existingRule,
				"nginx_error": err.Error(),
			})
			return
		}
		
		// 重载Nginx
		if err := utils.ReloadNginx(); err != nil {
			utils.ErrorWithData(c, 500, "更新成功但重载Nginx失败", gin.H{
				"rule": existingRule,
				"nginx_error": err.Error(),
			})
			return
		}
	}

	utils.SuccessWithMsg(c, "代理规则更新成功", existingRule)
}

// DeleteProxyRule 删除代理规则
func DeleteProxyRule(c *gin.Context) {
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

	// 删除Nginx配置
	if err := utils.DeleteNginxConfig(rule.Domain); err != nil {
		utils.Error(c, 500, "删除Nginx配置失败: "+err.Error())
		return
	}

	// 删除数据库记录
	if err := models.DeleteProxyRule(uint(ruleID)); err != nil {
		utils.InternalError(c, "删除代理规则失败: "+err.Error())
		return
	}

	// 重载Nginx
	if err := utils.ReloadNginx(); err != nil {
		utils.ErrorWithData(c, 500, "删除成功但重载Nginx失败", gin.H{
			"nginx_error": err.Error(),
		})
		return
	}

	utils.SuccessWithMsg(c, "代理规则删除成功", nil)
}

// ToggleProxyRule 切换代理规则状态
func ToggleProxyRule(c *gin.Context) {
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

	// 切换状态
	rule.Enabled = !rule.Enabled

	if err := models.UpdateProxyRule(rule); err != nil {
		utils.InternalError(c, "更新状态失败: "+err.Error())
		return
	}

	// 根据新状态生成或删除Nginx配置
	if rule.Enabled {
		if err := utils.GenerateNginxConfig(rule); err != nil {
			utils.ErrorWithData(c, 500, "状态更新成功但生成Nginx配置失败", gin.H{
				"rule": rule,
				"nginx_error": err.Error(),
			})
			return
		}
	} else {
		if err := utils.DeleteNginxConfig(rule.Domain); err != nil {
			utils.ErrorWithData(c, 500, "状态更新成功但删除Nginx配置失败", gin.H{
				"rule": rule,
				"nginx_error": err.Error(),
			})
			return
		}
	}

	// 重载Nginx
	if err := utils.ReloadNginx(); err != nil {
		utils.ErrorWithData(c, 500, "状态更新成功但重载Nginx失败", gin.H{
			"rule": rule,
			"nginx_error": err.Error(),
		})
		return
	}

	utils.SuccessWithMsg(c, "状态切换成功", rule)
}

// SyncProxyConfigs 同步所有代理配置到Nginx
func SyncProxyConfigs(c *gin.Context) {
	if err := utils.SyncAllProxyConfigs(); err != nil {
		utils.InternalError(c, "同步配置失败: "+err.Error())
		return
	}
	utils.SuccessWithMsg(c, "配置同步成功", nil)
}

// TestNginx 测试Nginx配置
func TestNginx(c *gin.Context) {
	ok, output := utils.TestNginxConfig()
	utils.Success(c, gin.H{
		"success": ok,
		"output":  output,
	})
}

// GetServerInfo 获取服务器信息（公网IP、下一个可用端口）
func GetServerInfo(c *gin.Context) {
	// 获取公网IP
	publicIP := ""
	resp, err := http.Get("https://api.ip.sb/ip")
	if err == nil {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		publicIP = strings.TrimSpace(string(body))
	}
	if publicIP == "" {
		publicIP = c.ClientIP()
	}

	// 获取下一个可用端口（从9000开始）
	nextPort := 9000
	ports, err := models.GetUsedPorts()
	if err == nil {
		portMap := make(map[int]bool)
		for _, p := range ports {
			portMap[p] = true
		}
		for p := 9000; p <= 65535; p++ {
			if !portMap[p] {
				nextPort = p
				break
			}
		}
	}

	utils.Success(c, gin.H{
		"public_ip": publicIP,
		"next_port": nextPort,
	})
}