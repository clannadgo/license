package hwid

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"net"
	"os"
	"strings"
	"time"
)

// 定义常量枚举，方便管理配置文件名
const (
	// HiddenConfigFileName 隐藏配置文件名
	HiddenConfigFileName = ".system_config.dat"
)

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

// getOrCreateHiddenFileHash 生成或更新隐藏文件，并计算其哈希值
// 每次调用都会写入当前时间戳，确保每次生成的指纹不同
func getOrCreateHiddenFileHash() string {
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
	hiddenFileHash := getOrCreateHiddenFileHash()

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
	// 格式化为 5-5-5-5-5
	parts := []string{s[0:5], s[5:10], s[10:15], s[15:20], s[20:25]}
	return strings.Join(parts, "-"), nil
}
