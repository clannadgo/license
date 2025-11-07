package examples

import (
	"github.com/gin-gonic/gin"
	"license/internal/hwid"
	"license/internal/license"
)

func main() {
	pub := "./public.pem"    // 放到镜像或挂载
	store := "./license.lic" // volume 持久化

	r := gin.Default()

	r.GET("/api/system/fingerprint", func(c *gin.Context) {
		fp := hwid.GetFingerprint()
		c.JSON(200, gin.H{"fingerprint": fp})
	})

	r.POST("/api/license/activate", license.ActivateHandler(pub, store))

	// global license middleware
	r.Use(license.LicenseMiddleware(pub, store))

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": "protected resource"})
	})

	r.Run(":8080")
}
