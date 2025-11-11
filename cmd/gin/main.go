package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"license/internal/config"
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

	// 启动协程，定期检查过期的许可证
	db.StartExpiredLicenseChecker()

	// 创建Gin路由器
	r := gin.Default()

	// 配置CORS中间件
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:5174"},
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
	pubKeyPath := config.Conf.PublicKeyPath
	storePath := config.Conf.LicenseStorePath
	privateKeyPath := config.Conf.PrivateKeyPath
	// 判断文件是否存在
	if _, err := os.Stat(pubKeyPath); os.IsNotExist(err) {
		log.Fatalf("Public key file not found: %s", pubKeyPath)
	}
	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		log.Fatalf("Private key file not found: %s", storePath)
	}
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
		api.POST("/license/activate", license.ActivateHandler(pubKeyPath, privateKeyPath, db))

		// 获取所有许可证激活记录（支持分页和客户名称搜索）
		api.GET("/license/activations", func(c *gin.Context) {
			// 获取分页参数
			page := 1
			pageSize := 10

			if pageStr := c.Query("page"); pageStr != "" {
				if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
					page = p
				}
			}

			if sizeStr := c.Query("size"); sizeStr != "" {
				if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
					pageSize = s
				}
			}

			// 获取客户名称搜索参数
			customerName := c.Query("customer")

			// 使用分页和搜索查询
			activations, total, err := db.GetLicenseActivationsWithPaginationAndSearch(page, pageSize, customerName)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get activations"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"activations": activations,
				"total":       total,
				"page":        page,
				"pageSize":    pageSize,
				"customer":    customerName,
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

		// 删除许可证激活记录
		api.DELETE("/license/activations/:id", func(c *gin.Context) {
			idStr := c.Param("id")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid license id"})
				return
			}

			err = db.DeleteLicenseActivation(id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "License deleted successfully",
			})
		})

		// 停用许可证激活记录
		api.PUT("/license/activations/:id/deactivate", func(c *gin.Context) {
			idStr := c.Param("id")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid license id"})
				return
			}

			err = db.DeactivateLicense(id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "License deactivated successfully",
			})
		})

		// 更新许可证激活记录（只允许更新客户名称和描述）
		api.PUT("/license/activations/:id", func(c *gin.Context) {
			idStr := c.Param("id")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid license id"})
				return
			}

			// 解析请求体
			var request struct {
				Customer    string `json:"customer" binding:"required"`
				Description string `json:"description"`
			}

			if err := c.ShouldBindJSON(&request); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// 验证客户名称不为空
			if strings.TrimSpace(request.Customer) == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "customer name cannot be empty"})
				return
			}

			// 更新许可证记录
			err = db.UpdateLicenseActivation(id, request.Customer, request.Description)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "License updated successfully",
			})
		})

		// 下载许可证文件
		api.GET("/license/activations/:id/download", func(c *gin.Context) {
			idStr := c.Param("id")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid license id"})
				return
			}

			// 获取许可证激活记录
			activation, err := db.GetLicenseActivationByID(int64(id))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get license activation"})
				return
			}

			if activation == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "license activation not found"})
				return
			}

			// 返回许可证内容
			c.JSON(http.StatusOK, gin.H{
				"success":        true,
				"licenseContent": activation.License,
			})
		})
	}

	// 应用许可证中间件到所有路由（除了健康检查和API路由组）
	r.Use(license.LicenseMiddleware(pubKeyPath, storePath, db))

	// 启动服务器
	log.Println("Starting license service on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
