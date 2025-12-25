package middleware

import (
	"Curry2API-go/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AuthRequired 认证中间件（支持多密钥和运行时密钥管理）
func AuthRequired() gin.HandlerFunc {
	// 获取密钥管理器实例
	km := GetKeyManager()

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// 调试日志：记录所有请求头
		if authHeader == "" {
			// 记录所有头信息以便调试
			headers := make(map[string]string)
			for key, values := range c.Request.Header {
				if len(values) > 0 {
					headers[key] = values[0]
				}
			}
			logrus.WithFields(logrus.Fields{
				"headers": headers,
				"path":    c.Request.URL.Path,
			}).Debug("Missing Authorization header - all request headers")
			
			errorResponse := models.NewErrorResponse(
				"Missing authorization header",
				"authentication_error",
				"missing_auth",
			)
			c.JSON(http.StatusUnauthorized, errorResponse)
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			errorResponse := models.NewErrorResponse(
				"Invalid authorization format. Expected 'Bearer <token>'",
				"authentication_error",
				"invalid_auth_format",
			)
			c.JSON(http.StatusUnauthorized, errorResponse)
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		// 使用密钥管理器验证密钥
		if !km.IsValidKey(token) {
			errorResponse := models.NewErrorResponse(
				"Invalid API key",
				"authentication_error",
				"invalid_api_key",
			)
			c.JSON(http.StatusUnauthorized, errorResponse)
			c.Abort()
			return
		}

		// Check balance status after token validation
		// Requirements: 3.2
		if err := km.CheckBalanceStatus(token); err != nil {
			if err == ErrBalanceExhausted {
				errorResponse := models.NewErrorResponse(
					"Insufficient balance - your account balance is exhausted",
					"payment_required",
					"balance_exhausted",
				)
				c.JSON(http.StatusPaymentRequired, errorResponse)
				c.Abort()
				return
			}
		}

		// Check token quota
		// Requirements: 12.4
		if err := km.CheckTokenQuota(token); err != nil {
			if err == ErrTokenQuotaExceeded {
				errorResponse := models.NewErrorResponse(
					"Token quota exceeded - this token has reached its spending limit",
					"payment_required",
					"token_quota_exceeded",
				)
				c.JSON(http.StatusPaymentRequired, errorResponse)
				c.Abort()
				return
			}
		}

		// Check token expiration
		// Requirements: 13.3
		if err := km.CheckTokenExpiration(token); err != nil {
			if err == ErrTokenExpired {
				errorResponse := models.NewErrorResponse(
					"Token expired - this token has passed its expiration date",
					"authentication_error",
					"token_expired",
				)
				c.JSON(http.StatusUnauthorized, errorResponse)
				c.Abort()
				return
			}
		}

		// 认证通过，记录使用次数
		km.IncrementUsage(token)

		// 将使用的密钥存入上下文（用于日志和管理）
		c.Set("api_key", token)
		
		// 获取密钥关联的用户信息并存入上下文（用于使用跟踪）
		km.mu.RLock()
		if keyInfo, exists := km.keys[token]; exists {
			if keyInfo.UserID != nil {
				c.Set("user_id", *keyInfo.UserID)
			}
			if keyInfo.Username != "" {
				c.Set("username", keyInfo.Username)
			}
			if keyInfo.TokenName != "" {
				c.Set("token_name", keyInfo.TokenName)
			}
		}
		km.mu.RUnlock()

		c.Next()
	}
}