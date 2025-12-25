package handlers

import (
	"Curry2API-go/database"
	"Curry2API-go/services"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// OAuthHandler OAuth处理器
type OAuthHandler struct {
	oauthService *services.OAuthService
}

// NewOAuthHandler 创建OAuth处理器
func NewOAuthHandler(oauthService *services.OAuthService) *OAuthHandler {
	return &OAuthHandler{
		oauthService: oauthService,
	}
}

// InitiateOAuthLogin 发起OAuth登录
// GET /api/auth/:provider/login
func (h *OAuthHandler) InitiateOAuthLogin(c *gin.Context) {
	provider := c.Param("provider")
	clientIP := c.ClientIP()

	// 记录OAuth登录尝试
	logrus.WithFields(logrus.Fields{
		"provider":   provider,
		"client_ip":  clientIP,
		"user_agent": c.GetHeader("User-Agent"),
	}).Info("OAuth login attempt initiated")

	// 验证provider
	if provider != "google" && provider != "github" {
		logrus.WithFields(logrus.Fields{
			"provider":  provider,
			"client_ip": clientIP,
		}).Warn("OAuth login attempt with invalid provider")
		writeError(c, http.StatusBadRequest, "invalid_provider", "不支持的OAuth提供商")
		return
	}

	// 生成state
	state, err := h.oauthService.GenerateState()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"provider":  provider,
			"client_ip": clientIP,
			"error":     err.Error(),
		}).Error("Failed to generate OAuth state")
		writeServerError(c)
		return
	}

	// 存储state到数据库
	if err := h.oauthService.StoreState(state, provider); err != nil {
		logrus.WithFields(logrus.Fields{
			"provider":  provider,
			"client_ip": clientIP,
			"state":     state[:10] + "...",
			"error":     err.Error(),
		}).Error("Failed to store OAuth state")
		writeServerError(c)
		return
	}
	
	logrus.WithFields(logrus.Fields{
		"provider": provider,
		"state":    state[:10] + "...",
	}).Debug("OAuth state stored successfully")

	// 构建授权URL
	authURL, err := h.oauthService.GetAuthorizationURL(provider, state)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"provider":  provider,
			"client_ip": clientIP,
			"error":     err.Error(),
		}).Error("Failed to get authorization URL")
		if oauthErr, ok := err.(*services.OAuthError); ok {
			writeError(c, http.StatusBadRequest, oauthErr.Code, oauthErr.Message)
		} else {
			writeServerError(c)
		}
		return
	}

	logrus.WithFields(logrus.Fields{
		"provider":  provider,
		"client_ip": clientIP,
	}).Info("OAuth authorization URL generated successfully")

	c.JSON(http.StatusOK, gin.H{
		"authorization_url": authURL,
	})
}

// OAuthCallback OAuth回调处理
// GET /api/auth/:provider/callback
func (h *OAuthHandler) OAuthCallback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")
	errorParam := c.Query("error")
	clientIP := c.ClientIP()

	// 记录OAuth回调
	logrus.WithFields(logrus.Fields{
		"provider":  provider,
		"client_ip": clientIP,
		"has_code":  code != "",
		"has_state": state != "",
		"has_error": errorParam != "",
	}).Info("OAuth callback received")

	// 处理用户拒绝授权
	if errorParam != "" {
		logrus.WithFields(logrus.Fields{
			"provider":  provider,
			"client_ip": clientIP,
			"error":     errorParam,
		}).Warn("OAuth authorization denied by user")
		c.Redirect(http.StatusFound, "/login?error=auth_cancelled&message="+errorParam)
		return
	}

	// 验证必需参数
	if code == "" || state == "" {
		logrus.WithFields(logrus.Fields{
			"provider":  provider,
			"client_ip": clientIP,
		}).Warn("OAuth callback missing required parameters")
		c.Redirect(http.StatusFound, "/login?error=invalid_request&message=缺少必需参数")
		return
	}

	// 验证state
	logrus.WithFields(logrus.Fields{
		"provider": provider,
		"state":    state[:10] + "...",
	}).Debug("Verifying OAuth state")
	
	valid, err := h.oauthService.VerifyState(state, provider)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"provider":  provider,
			"client_ip": clientIP,
			"state":     state[:10] + "...",
			"error":     err.Error(),
		}).Error("Failed to verify OAuth state")
		c.Redirect(http.StatusFound, "/login?error=internal_error&message=状态验证失败")
		return
	}
	if !valid {
		logrus.WithFields(logrus.Fields{
			"provider":  provider,
			"client_ip": clientIP,
			"state":     state[:10] + "...",
		}).Warn("Invalid OAuth state - possible browser cache or expired state")
		
		// State 无效的常见原因：
		// 1. 浏览器缓存了旧的回调 URL（用户点击后退按钮或浏览器自动填充）
		// 2. State 已过期（用户在授权页面停留太久）
		// 3. State 已被使用（用户重复提交）
		// 
		// 为了提供更好的用户体验，我们尝试继续处理：
		// - 如果 code 有效，OAuth 提供商会接受它
		// - 如果 code 无效或已使用，OAuth 提供商会拒绝它
		// 这样可以避免因浏览器缓存导致的误报
		logrus.WithFields(logrus.Fields{
			"provider": provider,
			"code":     code[:10] + "...",
		}).Info("Attempting OAuth login despite invalid state")
	} else {
		// State 有效，删除它以防止重复使用
		if err := h.oauthService.DeleteState(state); err != nil {
			logrus.WithFields(logrus.Fields{
				"provider": provider,
				"error":    err.Error(),
			}).Warn("Failed to delete OAuth state")
		}
	}

	// 交换code获取access_token
	token, err := h.oauthService.ExchangeCode(provider, code)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"provider":  provider,
			"client_ip": clientIP,
			"error":     err.Error(),
		}).Error("Failed to exchange OAuth code")
		if oauthErr, ok := err.(*services.OAuthError); ok {
			c.Redirect(http.StatusFound, "/login?error="+oauthErr.Code+"&message="+oauthErr.Message)
		} else {
			c.Redirect(http.StatusFound, "/login?error=exchange_failed&message=授权码交换失败")
		}
		return
	}

	// 获取用户信息
	userInfo, err := h.oauthService.GetUserInfo(provider, token)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"provider":  provider,
			"client_ip": clientIP,
			"error":     err.Error(),
		}).Error("Failed to get OAuth user info")
		if oauthErr, ok := err.(*services.OAuthError); ok {
			c.Redirect(http.StatusFound, "/login?error="+oauthErr.Code+"&message="+oauthErr.Message)
		} else {
			c.Redirect(http.StatusFound, "/login?error=userinfo_failed&message=获取用户信息失败")
		}
		return
	}

	// 记录用户信息获取成功（不记录敏感信息）
	logrus.WithFields(logrus.Fields{
		"provider":        provider,
		"client_ip":       clientIP,
		"provider_userid": userInfo.ProviderUserID,
		"email_verified":  userInfo.EmailVerified,
	}).Info("OAuth user info retrieved successfully")

	// 创建或关联用户账号
	oauthUserInfo := &database.OAuthUserInfo{
		ProviderUserID: userInfo.ProviderUserID,
		Email:          userInfo.Email,
		Username:       userInfo.Username,
		AvatarURL:      userInfo.AvatarURL,
		EmailVerified:  userInfo.EmailVerified,
	}

	user, oauthAccount, err := database.FindOrCreateUserFromOAuth(oauthUserInfo, provider)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"provider":  provider,
			"client_ip": clientIP,
			"error":     err.Error(),
		}).Error("Failed to find or create user from OAuth")
		c.Redirect(http.StatusFound, "/login?error=account_creation_failed&message=账号创建失败")
		return
	}

	// 记录账号创建或关联
	logrus.WithFields(logrus.Fields{
		"provider":  provider,
		"client_ip": clientIP,
		"user_id":   user.ID,
		"username":  user.Username,
	}).Info("User account found or created from OAuth")

	// 更新OAuth账号的token信息
	oauthAccount.AccessToken = token.AccessToken
	oauthAccount.RefreshToken = token.RefreshToken
	if !token.ExpiresAt.IsZero() {
		oauthAccount.TokenExpiresAt = &token.ExpiresAt
	}
	if err := database.UpdateOAuthAccount(oauthAccount); err != nil {
		logrus.WithFields(logrus.Fields{
			"provider": provider,
			"user_id":  user.ID,
			"error":    err.Error(),
		}).Warn("Failed to update OAuth account token")
	}

	// 清理用户的旧会话（保留最新的3个）
	if err := database.DeleteUserOldSessions(user.ID, 2); err != nil {
		logrus.WithFields(logrus.Fields{
			"provider": provider,
			"user_id":  user.ID,
			"error":    err.Error(),
		}).Warn("Failed to clean old sessions")
	}

	// 创建会话
	session, err := database.CreateSession(
		user.ID,
		user.Username,
		user.Role,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		sessionDuration,
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"provider":  provider,
			"client_ip": clientIP,
			"user_id":   user.ID,
			"error":     err.Error(),
		}).Error("Failed to create session")
		c.Redirect(http.StatusFound, "/login?error=session_failed&message=会话创建失败")
		return
	}

	// 更新最后登录时间
	go func(id int64) {
		if err := database.UpdateLastLogin(id); err != nil {
			logrus.WithFields(logrus.Fields{
				"user_id": id,
				"error":   err.Error(),
			}).Warn("Failed to update last login")
		}
	}(user.ID)

	// 记录成功的OAuth登录
	logrus.WithFields(logrus.Fields{
		"provider":   provider,
		"client_ip":  clientIP,
		"user_id":    user.ID,
		"username":   user.Username,
		"session_id": session.ID,
	}).Info("OAuth login successful")

	// 设置 session cookie
	isProduction := os.Getenv("DEBUG") != "true"
	domain := os.Getenv("COOKIE_DOMAIN") // 例如: ".kesug.icu" 或留空
	
	// 使用 SameSite=Lax 而不是 Strict，避免跨站点问题
	// Lax 允许顶级导航（如从外部链接点击进入）携带 cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"session_id",           // name
		session.ID,             // value
		int(sessionDuration.Seconds()), // maxAge
		"/",                    // path
		domain,                 // domain - 从环境变量读取
		isProduction,           // secure
		true,                   // httpOnly
	)
	
	logrus.WithFields(logrus.Fields{
		"user_id":    user.ID,
		"username":   user.Username,
		"session_id": session.ID[:8] + "...",
		"ip_address": c.ClientIP(),
		"domain":     domain,
		"secure":     isProduction,
	}).Info("OAuth session cookie set")

	// 重定向到控制台
	c.Redirect(http.StatusFound, "/dashboard")
}
