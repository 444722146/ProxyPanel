package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"proxypanel/config"
	"proxypanel/routes"
	"proxypanel/utils"
)

func main() {
	// 初始化配置
	config.InitConfig()
	
	// 设置运行模式
	os.Setenv("GIN_MODE", config.AppConfig.Server.Mode)
	
	// 初始化数据库
	db, err := utils.InitDB()
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer utils.CloseDB(db)
	
	// 同步所有代理配置到Nginx
	// 只在首次启动时同步
	log.Println("同步代理配置到Nginx...")
	if err := utils.SyncAllProxyConfigs(); err != nil {
		log.Printf("同步配置失败（可能Nginx未安装或未运行）: %v", err)
	} else {
		log.Println("配置同步完成")
	}
	
	// 设置路由
	router := routes.SetupRouter()
	
	// 启动服务器
	port := config.AppConfig.Server.Port
	address := fmt.Sprintf(":%s", port)
	
	log.Printf("ProxyPanel 服务启动在 http://localhost:%s", port)
	log.Printf("访问管理面板: http://localhost:%s/", port)
	
	// 优雅关闭
	go func() {
		if err := router.Run(address); err != nil {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()
	
	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("正在关闭服务...")
	log.Println("服务已关闭")
}