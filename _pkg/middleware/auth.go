package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// RequireAPIToken защищает мутирующие эндпоинты.
//
// Токен читается из API_TOKEN и передаётся клиентом в заголовке
// "Authorization: Bearer <token>" либо в "X-API-Token".
//
// Если API_TOKEN не задан, запрос отклоняется: незаданный секрет — это
// открытая на запись база, а не повод пропустить проверку.
func RequireAPIToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		expected := os.Getenv("API_TOKEN")
		if expected == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "server is not configured for write access",
			})
			return
		}

		provided := c.GetHeader("X-API-Token")
		if provided == "" {
			const prefix = "Bearer "
			if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, prefix) {
				provided = strings.TrimPrefix(auth, prefix)
			}
		}

		// subtle.ConstantTimeCompare защищает от подбора по времени ответа
		if subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or missing API token",
			})
			return
		}

		c.Next()
	}
}
