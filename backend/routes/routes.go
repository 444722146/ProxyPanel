package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"proxypanel/controllers"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	router := gin.New()
	
	// 使用中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	
	// CORS配置
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	router.Use(cors.New(corsConfig))

	// API路由组
	api := router.Group("/api")
	{
		// 代理规则路由
		proxyGroup := api.Group("/proxy")
		{
			proxyGroup.GET("", controllers.GetProxyRules)           // 获取所有代理规则
			proxyGroup.GET("/:id", controllers.GetProxyRule)        // 获取单个代理规则
			proxyGroup.POST("", controllers.CreateProxyRule)        // 创建代理规则
			proxyGroup.PUT("/:id", controllers.UpdateProxyRule)     // 更新代理规则
			proxyGroup.DELETE("/:id", controllers.DeleteProxyRule)  // 删除代理规则
			proxyGroup.POST("/:id/toggle", controllers.ToggleProxyRule) // 切换状态
		}

		// Nginx管理路由
		nginxGroup := api.Group("/nginx")
		{
			nginxGroup.POST("/sync", controllers.SyncProxyConfigs) // 同步配置
			nginxGroup.POST("/test", controllers.TestNginx)        // 测试配置
		}

		// 日志路由
		logGroup := api.Group("/log")
		{
			logGroup.GET("/access", controllers.GetAccessLog)      // 获取指定域名的访问日志
			logGroup.GET("/error", controllers.GetErrorLog)        // 获取指定域名的错误日志
			logGroup.GET("/access/general", controllers.GetGeneralAccessLog) // 获取通用访问日志
			logGroup.GET("/error/general", controllers.GetGeneralErrorLog)   // 获取通用错误日志
			logGroup.DELETE("/clear", controllers.ClearLog)        // 清空日志
			logGroup.GET("/search", controllers.SearchLog)         // 搜索日志
		}

		// SSL证书路由
		sslGroup := api.Group("/ssl")
		{
			sslGroup.POST("/:id/upload", controllers.UploadSSLCertificate)      // 上传证书
			sslGroup.POST("/:id/generate", controllers.GenerateSelfSignedCertificate) // 生成自签名证书
			sslGroup.DELETE("/:id", controllers.RemoveSSLCertificate)            // 移除证书
			sslGroup.POST("/:id/request-free", controllers.RequestFreeCert)      // 申请证书
			sslGroup.POST("/:id/renew", controllers.RenewCertificate)            // 续签证书
			sslGroup.GET("/:id/status", controllers.GetCertStatus)               // 获取证书状态
			sslGroup.POST("/:id/toggle-auto-renew", controllers.ToggleAutoRenew) // 切换自动续签
		}
	}

	// 静态文件服务（前端）
	router.Static("/assets", "./frontend/dist/assets")
	router.StaticFile("/", "./frontend/dist/index.html")
	router.StaticFile("/favicon.ico", "./frontend/dist/favicon.ico")
	
	// 前端路由处理（所有未匹配的路由返回index.html）
	router.NoRoute(func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
	})

	return router
}