package main

import (
	"log"
	"net/http"
	"time"

	"license/internal/database"
	"license/internal/hwid"
	"license/internal/license"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化数据库
	db, err := database.NewDB("license.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// 清理过期的许可证
	if err := db.CleanupExpiredLicenses(); err != nil {
		log.Printf("Warning: Failed to cleanup expired licenses: %v", err)
	}

	// 创建Gin路由器
	r := gin.Default()

	// 配置CORS中间件
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 添加中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 许可证中间件配置
	pubKeyPath := "public.pem"
	storePath := "license.lic"

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "License service is running",
		})
	})

	// API路由组
	api := r.Group("/api")
	{
		// 获取系统指纹
		api.GET("/system/fingerprint", func(c *gin.Context) {
			fingerprint := hwid.GetFingerprint()
			c.JSON(http.StatusOK, gin.H{
				"fingerprint": fingerprint,
			})
		})

		// 许可证激活端点
		api.POST("/license/activate", license.ActivateHandler(pubKeyPath, storePath, db))

		// 获取所有许可证激活记录
		api.GET("/license/activations", func(c *gin.Context) {
			activations, err := db.GetAllLicenseActivations()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get activations"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"activations": activations,
			})
		})

		// 获取已过期的许可证
		api.GET("/license/expired", func(c *gin.Context) {
			expired, err := db.GetExpiredLicenses()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get expired licenses"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"expired": expired,
			})
		})
	}

	// 应用许可证中间件到所有路由（除了健康检查和API路由组）
	r.Use(license.LicenseMiddleware(pubKeyPath, storePath))

	// 启动服务器
	log.Println("Starting license service on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
