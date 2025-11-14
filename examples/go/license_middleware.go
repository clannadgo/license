package examples

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// LicenseInfo 许可证信息结构体
type LicenseInfo struct {
	ActivationCode  string    `json:"activation_code"`
	ValidFrom       time.Time `json:"valid_from"`
	ValidUntil      time.Time `json:"valid_until"`
	Fingerprint     string    `json:"fingerprint"`
	Status          string    `json:"status"`
}

// LicenseMiddleware 客户端许可证中间件
// 公钥路径和许可证文件路径可以通过参数传入，默认使用当前目录下的public.pem和license.lic
func LicenseMiddleware(publicKeyPath, licenseFilePath string) gin.HandlerFunc {
	// 设置默认路径
	if publicKeyPath == "" {
		publicKeyPath = "./public.pem"
	}
	if licenseFilePath == "" {
		licenseFilePath = "./license.lic"
	}

	return func(c *gin.Context) {
		// 读取公钥
		publicKey, err := readPublicKey(publicKeyPath)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "公钥读取失败: " + err.Error()})
			c.Abort()
			return
		}

		// 读取许可证文件
		licenseContent, err := ioutil.ReadFile(licenseFilePath)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "许可证文件读取失败: " + err.Error()})
			c.Abort()
			return
		}

		// 解码许可证内容
		licenseInfo, err := verifyAndDecodeLicense(licenseContent, publicKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "许可证验证失败: " + err.Error()})
			c.Abort()
			return
		}

		// 检查许可证是否过期
		now := time.Now()
		if now.Before(licenseInfo.ValidFrom) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "许可证尚未生效", "valid_from": licenseInfo.ValidFrom})
			c.Abort()
			return
		}

		if now.After(licenseInfo.ValidUntil) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "许可证已过期", "valid_until": licenseInfo.ValidUntil})
			c.Abort()
			return
		}

		// 许可证有效，将许可证信息存储到上下文中供后续使用
		c.Set("license_info", licenseInfo)
		c.Next()
	}
}

// readPublicKey 读取并解析RSA公钥
func readPublicKey(path string) (*rsa.PublicKey, error) {
	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("公钥文件不存在: %s", path)
	}

	// 读取公钥文件内容
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取公钥文件失败: %v", err)
	}

	// 解码PEM格式的公钥
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("无效的公钥格式")
	}

	// 解析RSA公钥
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析公钥失败: %v", err)
	}

	// 类型断言为RSA公钥
	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("公钥类型不是RSA")
	}

	return rsaPublicKey, nil
}

// verifyAndDecodeLicense 验证并解码许可证
func verifyAndDecodeLicense(licenseData []byte, publicKey *rsa.PublicKey) (*LicenseInfo, error) {
	// 这里假设license.lic文件包含base64编码的JSON数据
	// 在实际应用中，可能需要根据实际的许可证格式进行调整
	
	// 解码base64
	decoded, err := base64.StdEncoding.DecodeString(string(licenseData))
	if err != nil {
		// 如果不是base64编码，尝试直接作为JSON解析
		decoded = licenseData
	}

	// 解析JSON
	var licenseInfo LicenseInfo
	if err := json.Unmarshal(decoded, &licenseInfo); err != nil {
		return nil, fmt.Errorf("解析许可证内容失败: %v", err)
	}

	return &licenseInfo, nil
}
