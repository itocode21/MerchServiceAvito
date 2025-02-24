package middleware

import (
	"net/http"
	"strings"

	"github.com/itocode21/MerchServiceAvito/internal/auth"

	"github.com/gin-gonic/gin"
)

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "токен не предоставлен"})
			c.Abort()
			return
		}

		tokenStr := strings.Split(authHeader, " ")
		if len(tokenStr) != 2 || tokenStr[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "неверный формат токена"})
			c.Abort()
			return
		}

		username, err := auth.ValidateJWT(tokenStr[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "недействительный токен"})
			c.Abort()
			return
		}

		c.Set("username", username)
		c.Next()
	}
}
