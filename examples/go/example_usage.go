package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"license/examples"
)

func main() {
	// 创建 Gin 引擎
	router := gin.Default()

	// 全局中间件配置
	// 方式1: 使用默认路径（./public.pem 和 ./license.lic）
	// router.Use(examples.LicenseMiddleware("", ""))

	// 方式2: 自定义文件路径
	router.Use(examples.LicenseMiddleware("./config/public.pem", "./config/license.lic"))

	// 只对特定路由组应用许可证验证
	api := router.Group("/api")
	{
		// 这些路由都需要许可证验证
		api.GET("/data", getData)
		api.POST("/update", updateData)
	}

	// 公开路由（不需要许可证验证）
	router.GET("/health", healthCheck)
	router.GET("/version", getVersion)

	// 启动服务器
	fmt.Println("服务器启动在 http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		fmt.Printf("服务器启动失败: %v\n", err)
	}
}

// getData 获取数据的处理函数
func getData(c *gin.Context) {
	// 从上下文中获取许可证信息
	licenseInfo, exists := c.Get("license_info")
	if exists {
		fmt.Printf("许可证信息: %+v\n", licenseInfo)
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "数据获取成功",
		"data": map[string]interface{}{
			"items": []string{"item1", "item2", "item3"},
		},
	})
}

// updateData 更新数据的处理函数
func updateData(c *gin.Context) {
	c.JSON(200, gin.H{
		"success": true,
		"message": "数据更新成功",
	})
}

// healthCheck 健康检查（公开接口）
func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "healthy",
	})
}

// getVersion 获取版本信息（公开接口）
func getVersion(c *gin.Context) {
	c.JSON(200, gin.H{
		"version": "1.0.0",
		"name":    "License Demo App",
	})
}
