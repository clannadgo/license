package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/square/go-jose/v3"
)

const (
	version = "1.0"
	name    = "license"
)

type licenseParam struct {
	privPath    string
	customer    string
	fingerprint string
	days        int
	out         string
	metaStr     string
	issuer      string
}

// loadPrivateKey 支持 PKCS1 和 PKCS8 格式
func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, fmt.Errorf("no pem data in %s", path)
	}

	// try PKCS1
	if pk1, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return pk1, nil
	}
	// try PKCS8
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if pk, ok := key.(*rsa.PrivateKey); ok {
			return pk, nil
		}
	}
	return nil, fmt.Errorf("unsupported private key format")
}

// decodeActivationCodeToHex 支持把 "XXXX-XXXX-XXXX-XXXX" (base32) -> hex string
// 当输入看起来像带 '-' 的激活码时调用
func decodeActivationCodeToHex(code string) (string, error) {
	// remove dashes and spaces, uppercase
	s := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(code, "-", ""), " ", ""))
	if len(s) != 16 {
		return "", fmt.Errorf("activation code must be 16 base32 chars (4-4-4-4)")
	}
	// base32 decode, RFC4648, no padding
	dec := base32.StdEncoding.WithPadding(base32.NoPadding)
	// we expect 10 bytes
	b, err := dec.DecodeString(s)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func main() {
	meta := map[string]interface{}{
		"name":    name,
		"version": version,
	}
	metaStr, err := json.Marshal(meta)
	if err != nil {
		fmt.Fprintf(os.Stderr, "json marshal meta error: %v\n", err)
		os.Exit(8)
	}
	param := &licenseParam{
		privPath:    "./private.pem",
		customer:    "测试用户",
		fingerprint: "DP3E-QBC7-POKI-APX6",
		days:        10,
		out:         "./license.lic",
		metaStr:     string(metaStr),
		issuer:      "lz",
	}
	if param.customer == "" || param.fingerprint == "" {
		fmt.Println("customer and fingerprint are required")
		flag.Usage()
		os.Exit(2)
	}

	// if fingerprint looks like activation code (contains '-'), decode it to hex
	fp := param.fingerprint
	if strings.Contains(param.fingerprint, "-") {
		h, err := decodeActivationCodeToHex(param.fingerprint)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to decode activation code: %v\n", err)
			os.Exit(3)
		}
		fp = h
	}

	priv, err := loadPrivateKey(param.privPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load private key: %v\n", err)
		os.Exit(4)
	}

	//exp := time.Now().UTC().Add(time.Duration(param.days) * 24 * time.Hour).Unix()
	exp := time.Now().UTC().Add(time.Duration(60) * time.Second).Unix() // 测试60秒过期

	// build payload (claims). Use explicit map to avoid ordering issues; jose will sign payload bytes.
	payload := map[string]interface{}{
		"iss":         param.issuer,
		"sub":         "license",
		"customer":    param.customer,
		"fingerprint": fp, // hex string
		"iat":         time.Now().UTC().Unix(),
		"exp":         exp,
	}

	// optional meta
	if param.metaStr != "" {
		var metaObj map[string]interface{}
		if err = json.Unmarshal([]byte(param.metaStr), &metaObj); err != nil {
			fmt.Fprintf(os.Stderr, "invalid meta json: %v\n", err)
			os.Exit(5)
		}
		payload["meta"] = metaObj
	}

	// marshal payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "json marshal payload error: %v\n", err)
		os.Exit(6)
	}

	// create signer with PS256 (RSA-PSS with SHA256)
	signerOpts := (&jose.SignerOptions{}).WithType("JWT")
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.PS256, Key: priv}, signerOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create signer: %v\n", err)
		os.Exit(7)
	}

	jws, err := signer.Sign(payloadBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to sign payload: %v\n", err)
		os.Exit(8)
	}

	compact, err := jws.CompactSerialize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to serialize jws: %v\n", err)
		os.Exit(9)
	}

	if err = os.WriteFile(param.out, []byte(compact), 0600); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write output: %v\n", err)
		os.Exit(10)
	}

	fmt.Printf("license written to %s\n", param.out)
}
