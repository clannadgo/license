package hwid

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

// getMachineID 读取 Linux /etc/machine-id
func getMachineID() string {
	if b, err := ioutil.ReadFile("/etc/machine-id"); err == nil {
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

// GetFingerprint 仅绑定物理机
func GetFingerprint() string {
	mid := getMachineID()
	mac := getPrimaryMAC()
	hn := getHostname()

	// 组合信息，SHA256 生成指纹
	combined := mid + "|" + mac + "|" + hn
	h := sha256.Sum256([]byte(combined))
	hexFP := hex.EncodeToString(h[:])
	code, err := ToActivationCodeFromHex(hexFP)
	if err != nil {
		panic(err)
	}
	return code
}

// ToActivationCode 将指纹（hex 或 raw bytes）转成 16-char base32 激活码 "XXXX-XXXX-XXXX-XXXX"
// 输入可为 hex string（如 sha256 hex）或任意字节 slice（请使用 ToActivationCodeFromBytes）
func ToActivationCodeFromHex(hexStr string) (string, error) {
	b, err := hex.DecodeString(strings.TrimSpace(hexStr))
	if err != nil {
		return "", err
	}
	return ToActivationCodeFromBytes(b)
}

// ToActivationCodeFromBytes 从原始字节数组生成激活码
// 规则：取 bytes 的前 10 个字节（不足则用整个数组并右补 0），用 RFC4648 base32 大写编码 -> 得到 16 字符 -> 分段 4-4-4-4
func ToActivationCodeFromBytes(b []byte) (string, error) {
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
