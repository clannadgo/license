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
	"net/http"
	"os"
	"strings"
	"time"

	"license/internal/database"
	"license/internal/hwid"

	"github.com/gin-gonic/gin"
	"github.com/square/go-jose/v3"
)

// ------------------ 公钥加载 ------------------

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

// ------------------ License Generation ------------------

func generateLicense(pubKeyPath, privateKeyPath, customer, fingerprint string, issuedAt time.Time, exp int64) (string, error) {
	// Load private key (assuming it's in the same directory as public key with .key extension)
	b, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read private key: %v", err)
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return "", errors.New("invalid private key pem")
	}
	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// 尝试使用PKCS8格式解析
		privKeyInterface, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return "", fmt.Errorf("failed to parse private key: %v", err)
		}
		var ok bool
		privKey, ok = privKeyInterface.(*rsa.PrivateKey)
		if !ok {
			return "", fmt.Errorf("not an RSA private key")
		}
	}

	// Create signing key
	signingKey := jose.SigningKey{
		Algorithm: jose.RS256,
		Key:       privKey,
	}

	// Create signer
	signer, err := jose.NewSigner(signingKey, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create signer: %v", err)
	}

	// Create claims
	c := claims{
		Iss:         "license-service",
		Sub:         customer,
		Customer:    customer,
		Fingerprint: fingerprint,
		Iat:         issuedAt.Unix(),
		Exp:         exp,
	}

	// Sign claims
	payload, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %v", err)
	}

	jws, err := signer.Sign(payload)
	if err != nil {
		return "", fmt.Errorf("failed to sign claims: %v", err)
	}

	return jws.CompactSerialize()
}

// ------------------ Activate Handler ------------------

func ActivateHandler(pubKeyPath, privateKeyPath string, db *database.DB) gin.HandlerFunc {
	pub, err := loadPublicKey(pubKeyPath)
	if err != nil {
		panic(err)
	}
	return func(c *gin.Context) {
		var req struct {
			Customer        string `json:"customer"`
			Fingerprint     string `json:"fingerprint"`
			Description     string `json:"description"`
			ValidityDays    int    `json:"validityDays"`
			ValidityHours   int    `json:"validityHours"`
			ValidityMinutes int    `json:"validityMinutes"`
			ValiditySeconds int    `json:"validitySeconds"`
			License         string // 用于内部存储生成的license，不从前端接收
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		// 获取本机激活码并转 hex
		fpCode := hwid.GetFingerprint() // XXXX-XXXX-XXXX-XXXX
		_, err := DecodeActivationCodeToHex(fpCode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode local fingerprint"})
			return
		}

		// 验证至少有一个时间单位被设置
		if req.ValidityDays == 0 && req.ValidityHours == 0 &&
			req.ValidityMinutes == 0 && req.ValiditySeconds == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "at least one time unit must be set"})
			return
		}

		// 校验激活码格式为XXXX-XXXX-XXXX-XXXX
		if strings.Contains(req.Fingerprint, "-") {
			// 验证格式是否为XXXX-XXXX-XXXX-XXXX
			parts := strings.Split(req.Fingerprint, "-")
			if len(parts) != 4 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "激活码格式不正确，应为XXXX-XXXX-XXXX-XXXX"})
				return
			}

			// 验证每个部分是否为4个字符
			for _, part := range parts {
				if len(part) != 4 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "激活码格式不正确，每个部分应为4个字符"})
					return
				}

				// 验证每个字符是否为有效的base32字符（A-Z和2-7）
				for _, char := range part {
					if !((char >= 'A' && char <= 'Z') || (char >= '2' && char <= '7')) {
						c.JSON(http.StatusBadRequest, gin.H{"error": "激活码包含无效字符，只允许A-Z和2-7"})
						return
					}
				}
			}
		} else {
			// 如果没有连字符，验证是否为16个字符的base32编码
			if len(req.Fingerprint) != 16 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "激活码长度不正确，应为16个字符"})
				return
			}

			// 验证每个字符是否为有效的base32字符
			for _, char := range req.Fingerprint {
				if !((char >= 'A' && char <= 'Z') || (char >= '2' && char <= '7')) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "激活码包含无效字符，只允许A-Z和2-7"})
					return
				}
			}
		}

		// 使用前端传入的原始fingerprint，不进行任何转换
		fp := req.Fingerprint

		// 如果fingerprint是带连字符的格式，转换为hex格式用于生成license
		var fpForLicense string
		if strings.Contains(req.Fingerprint, "-") {
			h, err := DecodeActivationCodeToHex(req.Fingerprint)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "failed to decode fingerprint: " + err.Error()})
				return
			}
			fpForLicense = h
		} else {
			fpForLicense = req.Fingerprint
		}

		// 计算过期时间
		now := time.Now().UTC()
		exp := now.Add(
			time.Duration(req.ValidityDays)*24*time.Hour +
				time.Duration(req.ValidityHours)*time.Hour +
				time.Duration(req.ValidityMinutes)*time.Minute +
				time.Duration(req.ValiditySeconds)*time.Second,
		).Unix()

		// 生成新的license
		newLicense, err := generateLicense(pubKeyPath, privateKeyPath, req.Customer, fpForLicense, now, exp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate license: " + err.Error()})
			return
		}

		// 使用新生成的license
		req.License = newLicense

		cl, err := verifyJWS(pub, req.License)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 不需要验证指纹匹配，因为我们使用前端传入的指纹
		// 这允许为任何机器生成许可证，而不仅限于当前机器

		// license已通过数据库存储，不需要写入文件系统

		// 记录激活信息到数据库
		if db != nil {
			// 检查是否已有该指纹的激活记录
			existingActivation, err := db.GetLicenseActivationByFingerprint(fp)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}

			// 如果已有激活记录且仍处于活动状态，返回错误提示指纹已存在
			if existingActivation != nil && existingActivation.IsActive {
				c.JSON(http.StatusBadRequest, gin.H{"error": "当前机器码已经存在，不能重复激活"})
				return
			}

			// 创建新的激活记录
			activation := &database.LicenseActivation{
				Customer:    cl.Customer,
				Fingerprint: fp, // 使用前端传入的指纹
				License:     req.License,
				Description: req.Description,
				IssuedAt:    time.Unix(cl.Iat, 0),
				ExpiresAt:   time.Unix(cl.Exp, 0),
				ActivatedAt: time.Now(),
				IsActive:    true,
			}

			err = db.InsertLicenseActivation(activation)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record activation"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "customer": cl.Customer, "exp": cl.Exp})
	}
}

// ------------------ License Middleware ------------------

func LicenseMiddleware(pubKeyPath, storePath string, db *database.DB) gin.HandlerFunc {
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

		// 获取本机激活码并转 hex
		fpCode := hwid.GetFingerprint() // XXXX-XXXX-XXXX-XXXX
		localHex, err := DecodeActivationCodeToHex(fpCode)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "failed to decode local fingerprint"})
			return
		}

		// 从数据库获取最新的有效license
		var licenseStr string
		if db != nil {
			activation, err := db.GetActiveLicenseActivationByFingerprint(localHex)
			if err != nil {
				c.AbortWithStatusJSON(500, gin.H{"error": "database error"})
				return
			}
			if activation != nil {
				licenseStr = activation.License
			}
		}

		// 如果数据库中没有找到，尝试从文件读取（向后兼容）
		if licenseStr == "" {
			b, err := os.ReadFile(storePath)
			if err != nil {
				c.AbortWithStatusJSON(403, gin.H{"error": "no license, please activate"})
				return
			}
			licenseStr = string(b)
		}

		cl, err := verifyJWS(pub, licenseStr)
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
