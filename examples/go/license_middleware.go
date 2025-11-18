package examples

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"license/internal/hwid"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/square/go-jose/v3"
)

// 定义常量枚举，方便管理配置文件名
const (
	// HiddenConfigFileName 隐藏配置文件名
	HiddenConfigFileName = ".system_config.dat"
)

type LicenseClaims struct {
	Iss         string `json:"iss"`
	Sub         string `json:"sub"`
	Customer    string `json:"customer"`
	Fingerprint string `json:"fingerprint"`
	Iat         int64  `json:"iat"`
	Exp         int64  `json:"exp"`
}

// getMachineID 读取 Linux /etc/machine-id
func getMachineID() string {
	if b, err := os.ReadFile("/etc/machine-id"); err == nil {
		return strings.TrimSpace(string(b))
	}
	return ""
}

// getPrimaryMAC 获取首个非回环网卡的 MAC 地址
func getPrimaryMAC() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if len(iface.HardwareAddr) == 0 {
			continue
		}
		return iface.HardwareAddr.String()
	}
	return ""
}

// getHostname 获取主机名
func getHostname() string {
	hn, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hn
}

// getHiddenFileHash 生成或更新隐藏文件，并计算其哈希值
// 每次调用都会写入当前时间戳，确保每次生成的指纹不同
func getHiddenFileHash() string {
	// 读取HiddenConfigFileName文件内容到[]bytes
	data, err := os.ReadFile(HiddenConfigFileName)
	if err != nil {
		panic("system file missed, check license error")
	}
	// 计算哈希值
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func CreateHiddenFileHash() string {
	// 使用一个不明显的文件名
	hiddenFilePath := HiddenConfigFileName

	// 获取当前时间戳
	timestamp := time.Now().UnixNano()

	// 生成随机数据
	randomData := make([]byte, 32)
	_, err := rand.Read(randomData)
	if err != nil {
		return ""
	}

	// 将时间戳和随机数据组合写入文件
	data := append([]byte{byte(timestamp >> 56), byte(timestamp >> 48), byte(timestamp >> 40), byte(timestamp >> 32),
		byte(timestamp >> 24), byte(timestamp >> 16), byte(timestamp >> 8), byte(timestamp)}, randomData...)

	// 写入文件（覆盖原有内容）
	err = os.WriteFile(hiddenFilePath, data, 0600)
	if err != nil {
		return ""
	}

	// 计算哈希值
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// GetFingerprint 生成包含隐藏文件哈希的指纹
func GetFingerprint() string {
	mid := getMachineID()
	mac := getPrimaryMAC()
	hn := getHostname()
	hiddenFileHash := getHiddenFileHash()
	// 组合信息，SHA256 生成指纹
	// 将隐藏文件哈希融入到指纹中，这样每次生成的指纹就会不同
	combined := mid + "|" + mac + "|" + hn + "|" + hiddenFileHash
	h := sha256.Sum256([]byte(combined))
	hexFP := hex.EncodeToString(h[:])
	code, err := ToActivationCodeFromHex(hexFP)
	if err != nil {
		panic(err)
	}
	return code
}

func CreateNewFingerprint() string {
	mid := getMachineID()
	mac := getPrimaryMAC()
	hn := getHostname()
	hiddenFileHash := CreateHiddenFileHash()
	// 组合信息，SHA256 生成指纹
	// 将隐藏文件哈希融入到指纹中，这样每次生成的指纹就会不同
	combined := mid + "|" + mac + "|" + hn + "|" + hiddenFileHash
	h := sha256.Sum256([]byte(combined))
	hexFP := hex.EncodeToString(h[:])
	code, err := ToActivationCodeFromHex(hexFP)
	if err != nil {
		panic(err)
	}
	return code
}

// ToActivationCodeFromHex 将指纹（hex 或 raw bytes）转成 25-char base32 激活码 "XXXXX-XXXXX-XXXXX-XXXXX-XXXXX"
func ToActivationCodeFromHex(hexStr string) (string, error) {
	b, err := hex.DecodeString(strings.TrimSpace(hexStr))
	if err != nil {
		return "", err
	}
	return ToActivationCodeFromBytes(b)
}

// ToActivationCodeFromBytes 从原始字节数组生成激活码
// 规则：取 bytes 的前 16 个字节（不足则用整个数组并右补 0），用 RFC4648 base32 大写编码 -> 得到 25 字符 -> 分段 5-5-5-5-5
func ToActivationCodeFromBytes(b []byte) (string, error) {
	const targetBytes = 16 // 128 bits -> 26 base32 chars, we'll use first 25
	buf := make([]byte, targetBytes)
	if len(b) >= targetBytes {
		copy(buf, b[:targetBytes])
	} else {
		copy(buf, b)
		// 若不足 16 字节，后面已是 0 补齐（可接受）
	}

	// 使用标准 base32 (RFC4648) 大写，无 padding
	enc := base32.StdEncoding.WithPadding(base32.NoPadding)
	s := enc.EncodeToString(buf) // 26 chars for 16 bytes, use first 25
	if len(s) < 25 {
		return "", errors.New("unexpected encoded length")
	}
	s = s[:25] // 截取前25个字符

	// 确保严格使用标准base32字符 (ABCDEFGHIJKLMNOPQRSTUVWXYZ234567)
	// 标准base32不包含0,1,O,I等容易混淆的字符
	s = strings.Map(func(r rune) rune {
		// 保留有效的base32字符
		if (r >= 'A' && r <= 'Z' && r != 'I' && r != 'L' && r != 'O') ||
			(r >= '2' && r <= '7') {
			return r
		}
		// 替换无效字符为有效的base32字符
		if r == '0' || r == 'O' || r == 'o' {
			return '2'
		} else if r == '1' || r == 'I' || r == 'L' || r == 'i' || r == 'l' {
			return 'A'
		}
		return 'A'
	}, s)

	// 格式化为 5-5-5-5-5
	parts := []string{s[0:5], s[5:10], s[10:15], s[15:20], s[20:25]}
	return strings.Join(parts, "-"), nil
}

// DecodeActivationCodeToHex 将 "XXXXX-XXXXX-XXXXX-XXXXX-XXXXX" -> hex string
func DecodeActivationCodeToHex(code string) (string, error) {
	s := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(code, "-", ""), " ", ""))
	if len(s) != 25 {
		return "", fmt.Errorf("activation code must be 25 base32 chars")
	}
	b, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(s)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
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

	// 预先加载公钥
	publicKey, err := readPublicKey(publicKeyPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load public key: %v", err))
	}

	return func(c *gin.Context) {
		// 跳过获取激活码的接口
		if c.FullPath() == "/license/code" {
			c.Next()
			return
		}

		// 验证许可证
		claims, err := verifyAndDecodeLicense(publicKey, licenseFilePath)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "许可证验证失败: " + err.Error()})
			c.Abort()
			return
		}

		// 获取当前机器的指纹
		currentFingerprint := hwid.GetFingerprint()
		currentHex, err := DecodeActivationCodeToHex(currentFingerprint)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode local fingerprint"})
			c.Abort()
			return
		}

		// 验证指纹匹配
		if claims.Fingerprint != "" && claims.Fingerprint != currentHex {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":               "许可证与当前设备不匹配",
				"license_fingerprint": claims.Fingerprint,
				"current_fingerprint": currentHex,
			})
			c.Abort()
			return
		}

		// 检查许可证是否过期
		if time.Now().UTC().Unix() > claims.Exp {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "许可证已过期", "valid_until": time.Unix(claims.Exp, 0)})
			c.Abort()
			return
		}

		// 许可证有效，将许可证信息存储到上下文中供后续使用
		c.Set("license.customer", claims.Customer)
		c.Set("license.expiresAt", claims.Exp)
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
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取公钥文件失败: %v", err)
	}

	// 解码PEM格式的公钥
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	// 尝试解析PKIX格式
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		// 尝试解析PKCS1格式
		rpub, err2 := x509.ParsePKCS1PublicKey(block.Bytes)
		if err2 == nil {
			return rpub, nil
		}
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
func verifyAndDecodeLicense(publicKey *rsa.PublicKey, licensePath string) (*LicenseClaims, error) {
	// 读取许可证文件（直接读取为JWS字符串）
	licenseContent, err := os.ReadFile(licensePath)
	if err != nil {
		return nil, fmt.Errorf("许可证文件读取失败: %v", err)
	}

	// 解析并验证JWS
	return verifyJWS(publicKey, string(licenseContent))
}

// verifyJWS 验证JWS签名并解析声明
func verifyJWS(pub *rsa.PublicKey, jwsCompact string) (*LicenseClaims, error) {
	signed, err := jose.ParseSigned(jwsCompact)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWS: %v", err)
	}

	// 验证签名
	out, err := signed.Verify(pub)
	if err != nil {
		return nil, fmt.Errorf("signature verification failed: %v", err)
	}

	// 解析声明
	var claims LicenseClaims
	if err := json.Unmarshal(out, &claims); err != nil {
		return nil, fmt.Errorf("解析许可证内容失败: %v", err)
	}

	return &claims, nil
}
