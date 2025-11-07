package license

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/square/go-jose/v3"
	"license/internal/hwid"
)

// ------------------ 公钥加载 ------------------

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	b, err := ioutil.ReadFile(path)
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

// ------------------ License Claims ------------------

type claims struct {
	Iss         string `json:"iss"`
	Sub         string `json:"sub"`
	Customer    string `json:"customer"`
	Fingerprint string `json:"fingerprint"`
	Iat         int64  `json:"iat"`
	Exp         int64  `json:"exp"`
	// Meta omitted
}

// ------------------ JWS 验证 ------------------

func verifyJWS(pub *rsa.PublicKey, jwsCompact string) (*claims, error) {
	signed, err := jose.ParseSigned(jwsCompact)
	if err != nil {
		return nil, err
	}
	// verify signature
	out, err := signed.Verify(pub)
	if err != nil {
		return nil, err
	}
	var c claims
	if err := json.Unmarshal(out, &c); err != nil {
		return nil, err
	}
	if time.Now().UTC().Unix() > c.Exp {
		return nil, errors.New("expired")
	}
	return &c, nil
}

// ------------------ HWID helper ------------------

// DecodeActivationCodeToHex 将 "XXXX-XXXX-XXXX-XXXX" -> hex string
func DecodeActivationCodeToHex(code string) (string, error) {
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

// ------------------ Activate Handler ------------------

func ActivateHandler(pubKeyPath, storePath string) gin.HandlerFunc {
	pub, err := loadPublicKey(pubKeyPath)
	if err != nil {
		panic(err)
	}
	return func(c *gin.Context) {
		var req struct {
			License string `json:"license"`
		}
		if err := c.BindJSON(&req); err != nil || req.License == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		// 获取本机激活码并转 hex
		fpCode := hwid.GetFingerprint() // XXXX-XXXX-XXXX-XXXX
		localHex, err := DecodeActivationCodeToHex(fpCode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode local fingerprint"})
			return
		}

		cl, err := verifyJWS(pub, req.License)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// fingerprint check
		if cl.Fingerprint != "" && cl.Fingerprint != localHex {
			c.JSON(http.StatusBadRequest, gin.H{"error": "fingerprint mismatch"})
			return
		}

		// persist license
		if err := os.WriteFile(storePath, []byte(req.License), 0600); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "save failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"ok": true, "customer": cl.Customer, "exp": cl.Exp})
	}
}

// ------------------ License Middleware ------------------

func LicenseMiddleware(pubKeyPath, storePath string) gin.HandlerFunc {
	pub, err := loadPublicKey(pubKeyPath)
	if err != nil {
		panic(err)
	}
	return func(c *gin.Context) {
		// allow activation endpoint and hwid endpoint
		if c.Request.Method == http.MethodPost && c.FullPath() == "/api/license/activate" {
			c.Next()
			return
		}
		if c.FullPath() == "/api/system/fingerprint" {
			c.Next()
			return
		}

		// read persisted license
		b, err := ioutil.ReadFile(storePath)
		if err != nil {
			c.AbortWithStatusJSON(403, gin.H{"error": "no license, please activate"})
			return
		}

		fpCode := hwid.GetFingerprint() // XXXX-XXXX-XXXX-XXXX
		localHex, err := DecodeActivationCodeToHex(fpCode)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "failed to decode local fingerprint"})
			return
		}

		cl, err := verifyJWS(pub, string(b))
		if err != nil {
			c.AbortWithStatusJSON(403, gin.H{"error": "invalid license: " + err.Error()})
			return
		}

		if cl.Fingerprint != "" && cl.Fingerprint != localHex {
			c.AbortWithStatusJSON(403, gin.H{"error": "fingerprint mismatch"})
			return
		}
		fmt.Println("license check ok")
		c.Set("license.customer", cl.Customer)
		c.Next()
	}
}
