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
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
	"unsafe"

	"github.com/square/go-jose/v3"
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
		if b, err := ioutil.ReadFile("/etc/machine-id"); err == nil {
			return strings.TrimSpace(string(b))
		}
		// 尝试其他可能的位置
		if b, err := ioutil.ReadFile("/var/lib/dbus/machine-id"); err == nil {
			return strings.TrimSpace(string(b))
		}
	case "darwin":
		// macOS系统
		if b, err := ioutil.ReadFile("/etc/hostid"); err == nil {
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

// 生成机器指纹
func getFingerprint() string {
	mid := getMachineID()
	mac := getPrimaryMAC()
	hn := getHostname()

	// 组合信息，SHA256 生成指纹
	combined := mid + "|" + mac + "|" + hn
	h := sha256.Sum256([]byte(combined))
	hexFP := hex.EncodeToString(h[:])
	code, err := toActivationCodeFromHex(hexFP)
	if err != nil {
		return ""
	}
	return code
}

// 将十六进制字符串转换为激活码格式
func toActivationCodeFromHex(hexStr string) (string, error) {
	b, err := hex.DecodeString(strings.TrimSpace(hexStr))
	if err != nil {
		return "", err
	}
	return toActivationCodeFromBytes(b)
}

// 将字节数组转换为激活码格式
func toActivationCodeFromBytes(b []byte) (string, error) {
	const targetBytes = 10 // 80 bits -> 16 base32 chars
	buf := make([]byte, targetBytes)
	if len(b) >= targetBytes {
		copy(buf, b[:targetBytes])
	} else {
		copy(buf, b)
		// 若不足 10 字节，后面已是 0 补齐（可接受）
	}

	// 使用标准 base32 (RFC4648) 大写，无 padding
	enc := base32.StdEncoding.WithPadding(base32.NoPadding)
	s := enc.EncodeToString(buf) // 16 chars for 10 bytes
	if len(s) != 16 {
		return "", errors.New("unexpected encoded length")
	}
	// 格式化为 4-4-4-4
	parts := []string{s[0:4], s[4:8], s[8:12], s[12:16]}
	return strings.Join(parts, "-"), nil
}

// 将激活码解码为十六进制字符串
func decodeActivationCodeToHex(code string) (string, error) {
	s := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(code, "-", ""), " ", ""))
	if len(s) != 16 {
		return "", fmt.Errorf("activation code must be 16 base32 chars")
	}
	b, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(s)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// 加载公钥
func loadPublicKey(path string) (*rsa.PublicKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errors.New("invalid pem")
	}
	pubIface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		// try rsa pub1
		rpub, err2 := x509.ParsePKCS1PublicKey(block.Bytes)
		if err2 == nil {
			return rpub, nil
		}
		return nil, err
	}
	if pub, ok := pubIface.(*rsa.PublicKey); ok {
		return pub, nil
	}
	return nil, errors.New("not rsa pub")
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
		return ErrorInvalidLicense, nil, err
	}

	// 验证签名
	out, err := signed.Verify(pub)
	if err != nil {
		return ErrorInvalidLicense, nil, err
	}

	// 解析claims
	var c claims
	if err := json.Unmarshal(out, &c); err != nil {
		return ErrorInvalidLicense, nil, err
	}

	// 检查过期时间
	if time.Now().UTC().Unix() > c.Exp {
		return ErrorLicenseExpired, nil, errors.New("license expired")
	}

	// 获取本机指纹
	localFingerprint := getFingerprint()
	localHex, err := decodeActivationCodeToHex(localFingerprint)
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
	fingerprint := getFingerprint()
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
