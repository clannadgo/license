package main

/*
#cgo CFLAGS: -Werror
#cgo linux LDFLAGS: -shared -fPIC
#cgo darwin LDFLAGS: -shared -fPIC
#include <stdlib.h>
#include <string.h>

// 导出函数声明
extern char* GenerateFingerprint();
extern int VerifyLicense(char* publicKeyPath, char* licenseContent);
extern char* GetLicenseData(char* publicKeyPath, char* licenseContent);
extern void FreeString(char* str);
*/
import "C"

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
	"unsafe"

	"github.com/square/go-jose/v3"
)

// 定义常量枚举，方便管理配置文件名
const (
	// HiddenConfigFileName 隐藏配置文件名
	HiddenConfigFileName = ".system_config.dat"
)

// 定义错误码
const (
	Success = iota
	ErrorInvalidPublicKey
	ErrorInvalidLicense
	ErrorLicenseExpired
	ErrorFingerprintMismatch
	ErrorInternal
)

// 定义许可证数据结构
type LicenseData struct {
	Issuer      string `json:"issuer"`
	Customer    string `json:"customer"`
	Fingerprint string `json:"fingerprint"`
	IssuedAt    int64  `json:"issuedAt"`
	ExpiresAt   int64  `json:"expiresAt"`
}

// 获取机器ID
func getMachineID() string {
	switch runtime.GOOS {
	case "linux":
		// Linux系统
		if b, err := os.ReadFile("/etc/machine-id"); err == nil {
			return strings.TrimSpace(string(b))
		}
		// 尝试其他可能的位置
		if b, err := os.ReadFile("/var/lib/dbus/machine-id"); err == nil {
			return strings.TrimSpace(string(b))
		}
	case "darwin":
		// macOS系统
		if b, err := os.ReadFile("/etc/hostid"); err == nil {
			return strings.TrimSpace(string(b))
		}
		// 尝试使用系统命令获取
		if output, err := runCommand("ioreg", "-rd1", "-c", "IOPlatformExpertDevice"); err == nil {
			lines := strings.Split(output, "\n")
			for _, line := range lines {
				if strings.Contains(line, "IOPlatformUUID") {
					parts := strings.Split(line, `"`)
					if len(parts) >= 4 {
						return strings.TrimSpace(parts[3])
					}
				}
			}
		}
	case "windows":
		// Windows系统
		if output, err := runCommand("wmic", "csproduct", "get", "UUID"); err == nil {
			lines := strings.Split(output, "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) != "" && !strings.Contains(line, "UUID") {
					return strings.TrimSpace(line)
				}
			}
		}
	}
	return ""
}

// 运行系统命令
func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// 获取主网卡MAC地址
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

// 获取主机名
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
		// 如果文件不存在或读取失败，创建新文件
		return CreateHiddenFileHash()
	}
	// 计算哈希值
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// CreateHiddenFileHash 创建隐藏文件并返回其哈希值
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
func getFingerprint() string {
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
		return ""
	}
	return code
}

// createNewFingerprint 每次调用都创建新的隐藏文件并生成新的指纹
func createNewFingerprint() string {
	mid := getMachineID()
	mac := getPrimaryMAC()
	hn := getHostname()
	// 每次都创建新的隐藏文件，确保生成不同的指纹
	hiddenFileHash := CreateHiddenFileHash()
	// 组合信息，SHA256 生成指纹
	combined := mid + "|" + mac + "|" + hn + "|" + hiddenFileHash
	h := sha256.Sum256([]byte(combined))
	hexFP := hex.EncodeToString(h[:])
	code, err := ToActivationCodeFromHex(hexFP)
	if err != nil {
		return ""
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

// readPublicKey 读取并解析RSA公钥
func loadPublicKey(path string) (*rsa.PublicKey, error) {
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

// 定义许可证claims结构
type claims struct {
	Iss         string `json:"iss"`
	Sub         string `json:"sub"`
	Customer    string `json:"customer"`
	Fingerprint string `json:"fingerprint"`
	Iat         int64  `json:"iat"`
	Exp         int64  `json:"exp"`
}

// 验证许可证
func verifyLicense(publicKeyPath, licenseContent string) (int, *LicenseData, error) {
	// 加载公钥
	pub, err := loadPublicKey(publicKeyPath)
	if err != nil {
		return ErrorInvalidPublicKey, nil, err
	}

	// 解析JWS
	signed, err := jose.ParseSigned(licenseContent)
	if err != nil {
		return ErrorInvalidLicense, nil, fmt.Errorf("failed to parse JWS: %v", err)
	}

	// 验证签名
	out, err := signed.Verify(pub)
	if err != nil {
		return ErrorInvalidLicense, nil, fmt.Errorf("signature verification failed: %v", err)
	}

	// 解析claims
	var c claims
	if err := json.Unmarshal(out, &c); err != nil {
		return ErrorInvalidLicense, nil, fmt.Errorf("解析许可证内容失败: %v", err)
	}

	// 检查过期时间
	if time.Now().UTC().Unix() > c.Exp {
		return ErrorLicenseExpired, nil, errors.New("license expired")
	}

	// 获取本机指纹
	localFingerprint := getFingerprint()
	localHex, err := DecodeActivationCodeToHex(localFingerprint)
	if err != nil {
		return ErrorInternal, nil, err
	}

	// 比较指纹
	if c.Fingerprint != localHex {
		return ErrorFingerprintMismatch, nil, errors.New("fingerprint mismatch")
	}

	// 返回许可证数据
	licenseData := &LicenseData{
		Issuer:      c.Iss,
		Customer:    c.Customer,
		Fingerprint: c.Fingerprint,
		IssuedAt:    c.Iat,
		ExpiresAt:   c.Exp,
	}

	return Success, licenseData, nil
}

// 导出函数：生成机器指纹
//
//export GenerateFingerprint
func GenerateFingerprint() *C.char {
	// 使用createNewFingerprint确保每次生成不同的指纹
	fingerprint := createNewFingerprint()
	return C.CString(fingerprint)
}

// 导出函数：验证许可证
//
//export VerifyLicense
func VerifyLicense(publicKeyPath, licenseContent *C.char) C.int {
	pubKeyPath := C.GoString(publicKeyPath)
	license := C.GoString(licenseContent)

	code, _, err := verifyLicense(pubKeyPath, license)
	if err != nil {
		// 错误已通过code返回
	}
	return C.int(code)
}

// 导出函数：检查许可证是否过期
// 返回值：0 = 未过期，1 = 已过期，-1 = 验证失败（无效的公钥或许可证）
//
//export GetLicenseExpired
func GetLicenseExpired(publicKeyPath, licenseContent *C.char) C.int {
	pubKeyPath := C.GoString(publicKeyPath)
	license := C.GoString(licenseContent)

	code, _, err := verifyLicense(pubKeyPath, license)
	if err != nil {
		// 如果有错误发生，返回-1表示验证失败
		return C.int(-1)
	}
	
	// 根据验证结果返回相应的值
	if code == ErrorLicenseExpired {
		return C.int(1) // 已过期
	} else if code == Success {
		return C.int(0) // 未过期
	} else {
		return C.int(-1) // 其他验证失败（如指纹不匹配等）
	}
}

// 导出函数：获取许可证数据（JSON格式）
//
//export GetLicenseData
func GetLicenseData(publicKeyPath, licenseContent *C.char) *C.char {
	pubKeyPath := C.GoString(publicKeyPath)
	license := C.GoString(licenseContent)

	code, licenseData, err := verifyLicense(pubKeyPath, license)
	if err != nil && code != ErrorFingerprintMismatch {
		return C.CString(fmt.Sprintf(`{"error": "%s"}`, err.Error()))
	}

	if licenseData == nil {
		return C.CString(`{"error": "no license data"}`)
	}

	// 转换为JSON
	jsonData, err := json.Marshal(licenseData)
	if err != nil {
		return C.CString(fmt.Sprintf(`{"error": "%s"}`, err.Error()))
	}

	return C.CString(string(jsonData))
}

// 导出函数：释放字符串内存
//
//export FreeString
func FreeString(str *C.char) {
	C.free(unsafe.Pointer(str))
}

func main() {
	// DLL入口点，不需要实现
}
