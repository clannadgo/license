# License 验证中间件使用说明

本示例提供了一个用于客户端应用的 Gin 中间件，用于验证许可证是否有效且未过期。

## 功能特点

- 自动读取公钥文件和许可证文件
- 验证许可证是否有效
- 检查许可证是否过期
- 支持自定义文件路径

## 使用方法

### 1. 导入中间件

```go
import (
    "github.com/gin-gonic/gin"
    "license/examples"
)
```

### 2. 创建示例应用

```go
package main

import (
	"github.com/gin-gonic/gin"
	"license/examples"
)

func main() {
	// 创建 Gin 引擎
	router := gin.Default()

	// 注册许可证验证中间件
	// 参数1: 公钥文件路径（可选，默认为 ./public.pem）
	// 参数2: 许可证文件路径（可选，默认为 ./license.lic）
	router.Use(examples.LicenseMiddleware("./path/to/public.pem", "./path/to/license.lic"))

	// 定义受保护的路由
	router.GET("/api/protected", func(c *gin.Context) {
		// 从上下文中获取许可证信息（如果需要）
		// licenseInfo, exists := c.Get("license_info")
		// if exists {
		//     // 使用许可证信息
		// }

		c.JSON(200, gin.H{
			"message": "访问成功，许可证有效",
		})
	})

	// 启动服务器
	router.Run(":8080")
}
```

## 许可证文件格式

许可证文件（license.lic）应该是包含以下字段的 JSON 格式或 base64 编码的 JSON：

```json
{
    "activation_code": "激活码",
    "valid_from": "生效时间（ISO格式）",
    "valid_until": "过期时间（ISO格式）",
    "fingerprint": "设备指纹",
    "status": "许可证状态"
}
```

## 注意事项

1. 确保公钥文件和许可证文件可读
2. 根据实际情况调整许可证验证逻辑
3. 在生产环境中，建议对许可证进行签名验证

## 示例

见 `example_usage.go` 文件，提供了一个完整的使用示例。
