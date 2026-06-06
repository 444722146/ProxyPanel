package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ProxyRule 代理规则模型
type ProxyRule struct {
	ID         uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Name       string         `json:"name" gorm:"type:varchar(255);not null;uniqueIndex:idx_name"`
	Domain     string         `json:"domain" gorm:"type:varchar(255);not null;uniqueIndex:idx_domain_port"`
	Port       int            `json:"port" gorm:"default:0;uniqueIndex:idx_domain_port"` // 监听端口，0表示使用默认端口
	TargetURL  string         `json:"target_url" gorm:"type:varchar(500);not null"`
	FakeIP     string         `json:"fake_ip" gorm:"type:varchar(50);default:'127.0.0.1'"`
	Token      string         `json:"token" gorm:"type:varchar(255)"`
	Enabled    bool           `json:"enabled" gorm:"default:true"`
	Whitelist  string         `json:"whitelist" gorm:"type:text"` // IP白名单，多个IP用逗号分隔
	SSLEnabled  bool           `json:"ssl_enabled" gorm:"default:false"`
	SSLType     string         `json:"ssl_type" gorm:"type:varchar(20);default:'manual'"` // manual/selfsigned/letsencrypt/zerossl
	SSLCert     string         `json:"ssl_cert,omitempty" gorm:"type:text"`
	SSLKey      string         `json:"ssl_key,omitempty" gorm:"type:text"`
	SSLAutoRenew bool          `json:"ssl_auto_renew" gorm:"default:false"`
	SSLExpiresAt *time.Time    `json:"ssl_expires_at,omitempty" gorm:"type:datetime"`
	AccessURL   string         `json:"access_url" gorm:"-"` // 计算字段，不存数据库
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// AfterFind GORM钩子，查询后自动计算访问地址
func (r *ProxyRule) AfterFind(tx *gorm.DB) error {
	if r.Domain != "" {
		scheme := "http"
		if r.SSLEnabled {
			scheme = "https"
		}
		port := r.Port
		if port == 0 {
			if r.SSLEnabled {
				port = 443
			} else {
				port = 80
			}
		}
		if (scheme == "http" && port != 80) || (scheme == "https" && port != 443) {
			r.AccessURL = fmt.Sprintf("%s://%s:%d", scheme, r.Domain, port)
		} else {
			r.AccessURL = fmt.Sprintf("%s://%s", scheme, r.Domain)
		}
	}
	return nil
}

// TableName 指定表名
func (ProxyRule) TableName() string {
	return "proxy_rules"
}

// CreateProxyRule 创建代理规则
func CreateProxyRule(rule *ProxyRule) error {
	return db.Create(rule).Error
}

// GetProxyRuleByID 根据ID获取代理规则
func GetProxyRuleByID(id uint) (*ProxyRule, error) {
	var rule ProxyRule
	err := db.First(&rule, id).Error
	return &rule, err
}

// GetProxyRuleByDomain 根据域名获取代理规则
func GetProxyRuleByDomain(domain string) (*ProxyRule, error) {
	var rule ProxyRule
	err := db.Where("domain = ?", domain).First(&rule).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// GetProxyRuleByDomainAndPort 根据域名和端口获取代理规则（组合唯一校验）
func GetProxyRuleByDomainAndPort(domain string, port int) (*ProxyRule, error) {
	var rule ProxyRule
	err := db.Where("domain = ? AND port = ?", domain, port).First(&rule).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// GetAllProxyRules 获取所有代理规则
func GetAllProxyRules() ([]ProxyRule, error) {
	var rules []ProxyRule
	err := db.Order("created_at DESC").Find(&rules).Error
	return rules, err
}

// UpdateProxyRule 更新代理规则
func UpdateProxyRule(rule *ProxyRule) error {
	return db.Save(rule).Error
}

// DeleteProxyRule 删除代理规则（硬删除，避免软删除导致唯一索引冲突）
func DeleteProxyRule(id uint) error {
	return db.Unscoped().Delete(&ProxyRule{}, id).Error
}

// GetEnabledProxyRules 获取所有启用的代理规则
func GetEnabledProxyRules() ([]ProxyRule, error) {
	var rules []ProxyRule
	err := db.Where("enabled = ?", true).Find(&rules).Error
	return rules, err
}

// GetUsedPorts 获取所有已使用的端口（大于0的）
func GetUsedPorts() ([]int, error) {
	var ports []int
	err := db.Model(&ProxyRule{}).Where("port > 0").Pluck("port", &ports).Error
	return ports, err
}

// db 全局数据库实例，在database.go中初始化
var db *gorm.DB

// SetDB 设置数据库实例
func SetDB(database *gorm.DB) {
	db = database
}