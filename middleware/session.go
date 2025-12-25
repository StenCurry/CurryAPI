package middleware

import (
	"Curry2API-go/database"
	"Curry2API-go/models"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SessionAuth 會話驗證
func SessionAuth() gin.HandlerFunc {
	km := GetKeyManager()

	return func(c *gin.Context) {
		// 清除任何可能的旧用户信息
		c.Set("user_id", nil)
		c.Set("username", nil)
		c.Set("role", nil)
		c.Set("session_id", nil)
		
		if ok := validateSessionCookie(c); ok {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == km.GetAdminToken() {
				logrus.WithFields(logrus.Fields{
					"client_ip": c.ClientIP(),
					"token_prefix": token[:4] + "...",
				}).Info("Admin token authentication successful")
				c.Set("user_id", int64(-1))
				c.Set("username", "admin")
				c.Set("role", "admin")
				c.Next()
				return
			}
		}

		logrus.WithFields(logrus.Fields{
			"client_ip": c.ClientIP(),
			"path": c.Request.URL.Path,
			"has_auth_header": c.GetHeader("Authorization") != "",
		}).Info("Authentication failed - no valid session or token")
		
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			"未登录，请先登录",
			"invalid_session",
			"invalid_session",
		))
		c.Abort()
	}
}

// ValidateSession 驗證會話ID並返回會話信息（公開函數供其他包使用）
func ValidateSession(sessionID string) (*database.Session, error) {
	if sessionID == "" {
		return nil, errors.New("session ID is empty")
	}

	session, err := database.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func validateSessionCookie(c *gin.Context) bool {
	sessionID, err := c.Cookie("session_id")
	
	// 详细日志：记录会话验证尝试
	logrus.WithFields(logrus.Fields{
		"has_cookie":  err == nil && sessionID != "",
		"session_id":  func() string {
			if sessionID != "" && len(sessionID) > 8 {
				return sessionID[:8] + "..."
			}
			return "none"
		}(),
		"client_ip":   c.ClientIP(),
		"user_agent":  c.GetHeader("User-Agent"),
		"path":        c.Request.URL.Path,
	}).Info("Session validation attempt")
	
	if err != nil || sessionID == "" {
		logrus.Info("No session cookie found - clearing any stale cookies")
		// 强制清除可能存在的无效cookie
		domain := os.Getenv("COOKIE_DOMAIN")
		c.SetCookie("session_id", "", -1, "/", domain, false, true)
		return false
	}

	session, err := ValidateSession(sessionID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"session_id": sessionID[:8] + "...",
			"error":      err.Error(),
			"client_ip":  c.ClientIP(),
		}).Warn("Session validation failed - clearing cookie")
		
		// 清除无效的session cookie
		domain := os.Getenv("COOKIE_DOMAIN")
		c.SetCookie("session_id", "", -1, "/", domain, false, true)
		return false
	}

	// 验证IP地址是否匹配（可选，通过环境变量控制）
	checkIP := os.Getenv("SESSION_CHECK_IP")
	if checkIP == "true" {
		currentIP := c.ClientIP()
		if session.IPAddress != currentIP {
			logrus.WithFields(logrus.Fields{
				"session_id":      session.ID[:8] + "...",
				"session_ip":      session.IPAddress,
				"current_ip":      currentIP,
				"user_id":         session.UserID,
				"username":        session.Username,
			}).Warn("Session IP mismatch - possible session hijacking")
			
			// IP不匹配，删除会话
			_ = database.DeleteSession(sessionID)
			
			// 清除客户端cookie
			domain := os.Getenv("COOKIE_DOMAIN")
			c.SetCookie("session_id", "", -1, "/", domain, false, true)
			return false
		}
	}

	// 成功验证会话
	logrus.WithFields(logrus.Fields{
		"user_id":    session.UserID,
		"username":   session.Username,
		"role":       session.Role,
		"session_id": session.ID[:8] + "...",
		"client_ip":  c.ClientIP(),
	}).Debug("Session validated successfully")

	// 验证用户是否仍然活跃
	user, err := database.GetUserByID(session.UserID)
	if err != nil || !user.IsActive {
		logrus.WithFields(logrus.Fields{
			"user_id":    session.UserID,
			"username":   session.Username,
			"session_id": session.ID[:8] + "...",
			"error":      err,
		}).Warn("Session user is inactive or not found")
		
		// 删除无效会话
		_ = database.DeleteSession(sessionID)
		
		// 清除客户端cookie
		domain := os.Getenv("COOKIE_DOMAIN")
		c.SetCookie("session_id", "", -1, "/", domain, false, true)
		return false
	}

	c.Set("user_id", session.UserID)
	c.Set("username", session.Username)
	c.Set("role", session.Role)
	c.Set("session_id", session.ID)
	return true
}

// AdminOnly 僅允許管理員訪問
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, models.NewErrorResponse(
				"仅限管理员访问",
				"admin_only",
				"admin_only",
			))
			c.Abort()
			return
		}
		c.Next()
	}
}
