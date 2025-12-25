package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// OptionalAuth 可选认证中间件
// 如果提供了 Authorization 头，则验证它
// 如果没有提供，则使用默认的 API key（仅用于开发/测试）
func OptionalAuth(defaultKey string) gin.HandlerFunc {
	km := GetKeyManager()

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// 如果没有 Authorization 头，使用默认 key
		if authHeader == "" {
			logrus.WithFields(logrus.Fields{
				"path":        c.Request.URL.Path,
				"default_key": defaultKey[:10] + "...",
			}).Debug("No Authorization header, using default key")

			// 验证默认 key 是否有效
			if defaultKey != "" && km.IsValidKey(defaultKey) {
				km.IncrementUsage(defaultKey)
				c.Set("api_key", defaultKey)
				c.Next()
				return
			}

			// 如果默认 key 无效，记录警告但继续
			logrus.Warn("Default key is invalid or not set, proceeding without authentication")
			c.Next()
			return
		}

		// 如果有 Authorization 头，正常验证
		if !strings.HasPrefix(authHeader, "Bearer ") {
			logrus.Debug("Invalid authorization format")
			c.Next()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		if km.IsValidKey(token) {
			km.IncrementUsage(token)
			c.Set("api_key", token)
			logrus.Debug("Authorization successful with provided token")
		} else {
			logrus.Debug("Provided token is invalid, proceeding anyway")
		}

		c.Next()
	}
}
