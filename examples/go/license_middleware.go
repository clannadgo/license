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
	"strings"
	"time"

	"license/internal/hwid"

	"github.com/gin-gonic/gin"
)

// LicenseInfo 许可证信息结构体
type LicenseInfo struct {
	ActivationCode string    `json:"activation_code"`
	ValidFrom      time.Time `json:"valid_from"`
	ValidUntil     time.Time `json:"valid_until"`
	Fingerprint    string    `json:"fingerprint"`
	Status         string    `json:"status"`
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

		// 获取当前机器的指纹
		currentFingerprint := hwid.GetFingerprint()

		// 比较指纹是否匹配（忽略连字符进行比较）
		licenseFingerprintClean := strings.ReplaceAll(licenseInfo.Fingerprint, "-", "")
		currentFingerprintClean := strings.ReplaceAll(currentFingerprint, "-", "")

		if licenseFingerprintClean != currentFingerprintClean {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":               "许可证与当前设备不匹配",
				"license_fingerprint": licenseInfo.Fingerprint,
				"current_fingerprint": currentFingerprint,
			})
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

		// 检查许可证状态
		if licenseInfo.Status != "active" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":  "许可证状态无效",
				"status": licenseInfo.Status,
			})
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
	// 首先尝试解码base64编码的许可证数据
	decoded, err := base64.StdEncoding.DecodeString(string(licenseData))
	if err != nil {
		// 如果base64解码失败，可能不是base64编码，尝试直接作为JSON解析
		decoded = licenseData
	}

	// 解析JSON格式的许可证信息
	var licenseInfo LicenseInfo
	if err := json.Unmarshal(decoded, &licenseInfo); err != nil {
		return nil, fmt.Errorf("解析许可证内容失败: %v", err)
	}

	// 验证激活码格式（5-5-5-5-5格式）
	if err := validateActivationCode(licenseInfo.ActivationCode); err != nil {
		return nil, fmt.Errorf("激活码格式无效: %v", err)
	}

	// 验证指纹格式
	if err := validateActivationCode(licenseInfo.Fingerprint); err != nil {
		return nil, fmt.Errorf("指纹格式无效: %v", err)
	}

	// 这里可以添加签名验证逻辑，但需要根据实际的许可证签名机制实现
	// 目前主要依赖指纹验证和时间验证

	return &licenseInfo, nil
}

// validateActivationCode 验证激活码或指纹格式是否正确
func validateActivationCode(code string) error {
	// 去除连字符
	code = strings.ReplaceAll(code, "-", "")

	// 检查长度是否为25个字符
	if len(code) != 25 {
		return fmt.Errorf("长度应为25个字符")
	}

	// 检查字符是否符合base32标准（A-Z和2-7）
	for _, char := range code {
		if !((char >= 'A' && char <= 'Z') || (char >= '2' && char <= '7')) {
			return fmt.Errorf("包含无效字符，只允许A-Z和2-7")
		}
	}

	return nil
}
