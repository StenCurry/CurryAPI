package handlers

import (
	"Curry2API-go/config"
	"Curry2API-go/database"
	"Curry2API-go/models"
	"Curry2API-go/services"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const sessionDuration = 24 * time.Hour

var emailService *services.EmailService
var turnstileService *services.TurnstileService

// InitEmailService 初始化邮件服务
func InitEmailService(cfg *config.Config) {
	emailService = services.NewEmailService(cfg)
}

// InitTurnstileService 初始化 Turnstile 服务
func InitTurnstileService(secretKey string) {
	turnstileService = services.NewTurnstileService(secretKey)
}

// SendVerificationCodeRequest 发送验证码请求
type SendVerificationCodeRequest struct {
	Email          string `json:"email" binding:"required,email"`
	TurnstileToken string `json:"turnstile_token" binding:"required"`
}

// RegisterRequest 註冊請求
type RegisterRequest struct {
	Username       string `json:"username" binding:"required,min=3,max=32"`
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=6"`
	Code           string `json:"code" binding:"required,len=6"`
	TurnstileToken string `json:"turnstile_token" binding:"required"`
	ReferralCode   string `json:"referral_code,omitempty"` // Optional referral code
}

// LoginRequest 登入請求
type LoginRequest struct {
	UsernameOrEmail string `json:"username_or_email" binding:"required"`
	Password        string `json:"password" binding:"required"`
}

// RegisterHandler 使用者註冊（需要验证码）
func RegisterHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "请求参数无效: "+err.Error())
		return
	}

	// 验证 Turnstile token（必需）
	if turnstileService == nil {
		logrus.Error("Turnstile service not initialized")
		writeError(c, http.StatusInternalServerError, "service_error", "验证服务未初始化")
		return
	}

	success, err := turnstileService.VerifyToken(req.TurnstileToken, c.ClientIP())
	if err != nil || !success {
		logrus.Warnf("Turnstile verification failed for IP %s: %v", c.ClientIP(), err)
		writeError(c, http.StatusBadRequest, "captcha_failed", "人机验证失败，请重试")
		return
	}

	// 验证验证码
	if err := database.VerifyCode(req.Email, req.Code, "register"); err != nil {
		if err == database.ErrCodeNotFound {
			writeError(c, http.StatusBadRequest, "code_not_found", "验证码不存在或已过期")
		} else if err == database.ErrCodeExpired {
			writeError(c, http.StatusBadRequest, "code_expired", "验证码已过期")
		} else if err == database.ErrCodeInvalid {
			writeError(c, http.StatusBadRequest, "code_invalid", "验证码错误")
		} else {
			logrus.Errorf("Failed to verify code: %v", err)
			writeServerError(c)
		}
		return
	}

	if err := ensureUserAvailable(req.Username, req.Email); err != nil {
		if apiErr, ok := err.(*apiError); ok {
			writeError(c, apiErr.status, apiErr.code, apiErr.message)
		} else {
			writeServerError(c)
		}
		return
	}

	user, err := database.CreateUser(req.Username, req.Email, req.Password, "user")
	if err != nil {
		logrus.Errorf("Failed to create user: %v", err)
		writeServerError(c)
		return
	}

	logrus.Infof("User registered: %s (ID: %d)", user.Username, user.ID)

	// Create user balance record with initial balance of $50
	// Requirements: 1.1, 4.1
	userBalance, err := database.CreateUserBalance(user.ID)
	if err != nil {
		logrus.Errorf("Failed to create user balance for user %d: %v", user.ID, err)
		// Note: User is already created, so we log the error but don't fail the registration
		// The balance can be created later by admin if needed
	} else {
		logrus.Infof("User balance created for user %d with initial balance $%.2f and referral code %s",
			user.ID, userBalance.Balance, userBalance.ReferralCode)
	}

	// Process referral bonus if valid referral code provided
	// Requirements: 5.1, 5.2, 5.5
	var referralProcessed bool
	if req.ReferralCode != "" && userBalance != nil {
		referral, err := database.ProcessReferralBonus(req.ReferralCode, user.ID)
		if err != nil {
			if err == database.ErrReferralCodeNotFound {
				logrus.Warnf("Invalid referral code '%s' provided during registration for user %d", req.ReferralCode, user.ID)
				// Continue with registration without referral bonus (Requirement 5.5)
			} else if err == database.ErrSelfReferral {
				logrus.Warnf("Self-referral attempted by user %d with code '%s'", user.ID, req.ReferralCode)
				// Continue with registration without referral bonus
			} else {
				logrus.Errorf("Failed to process referral bonus for user %d: %v", user.ID, err)
				// Continue with registration without referral bonus
			}
		} else {
			referralProcessed = true
			logrus.Infof("Referral bonus processed: referrer_id=%d, referee_id=%d, bonus=$%.2f",
				referral.ReferrerID, referral.RefereeID, referral.BonusAmount)
		}
	}

	response := gin.H{
		"message": "注册成功",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	}

	// Include balance info in response if created successfully
	if userBalance != nil {
		response["balance"] = gin.H{
			"amount":        userBalance.Balance,
			"referral_code": userBalance.ReferralCode,
		}
		if referralProcessed {
			response["referral_bonus_applied"] = true
		}
	}

	c.JSON(http.StatusCreated, response)
}

// LoginHandler 使用者登入
func LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "请求参数无效")
		return
	}

	identifier := strings.TrimSpace(req.UsernameOrEmail)
	if identifier == "" {
		writeError(c, http.StatusBadRequest, "invalid_request", "用户名或邮箱不能为空")
		return
	}

	var (
		user *database.User
		err  error
	)

	if strings.Contains(identifier, "@") {
		user, err = database.GetUserByEmail(identifier)
	} else {
		user, err = database.GetUserByUsername(identifier)
	}

	if err != nil {
		if err == database.ErrUserNotFound {
			writeError(c, http.StatusUnauthorized, "invalid_credentials", "用户名或密码错误")
			return
		}
		logrus.Errorf("Failed to query user: %v", err)
		writeServerError(c)
		return
	}

	if !database.ValidatePassword(user, req.Password) {
		writeError(c, http.StatusUnauthorized, "invalid_credentials", "用户名或密码错误")
		return
	}

	// 检查账号状态
	if !user.IsActive {
		writeError(c, http.StatusForbidden, "account_disabled", "您的账号存在问题，请联系管理员")
		return
	}

	// 清理用户的旧会话（保留最新的3个）
	if err := database.DeleteUserOldSessions(user.ID, 2); err != nil {
		logrus.Warnf("Failed to clean old sessions for user %d: %v", user.ID, err)
	}

	session, err := database.CreateSession(
		user.ID,
		user.Username,
		user.Role,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		sessionDuration,
	)
	if err != nil {
		logrus.Errorf("Failed to create session: %v", err)
		writeServerError(c)
		return
	}

	go func(id int64) {
		if err := database.UpdateLastLogin(id); err != nil {
			logrus.Warnf("Failed to update last login for user %d: %v", id, err)
		}
	}(user.ID)

	logrus.Infof("User logged in: %s (Session: %s)", user.Username, session.ID)

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
	}).Info("Session cookie set")

	c.JSON(http.StatusOK, gin.H{
		"message":    "登录成功",
		"session_id": session.ID,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

// LogoutHandler 登出
func LogoutHandler(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		writeError(c, http.StatusUnauthorized, "no_session", "未登录")
		return
	}

	if err := database.DeleteSession(sessionID); err != nil {
		logrus.Warnf("Failed to delete session %s: %v", sessionID, err)
	}

	// 清除cookie，使用与设置时相同的domain
	domain := os.Getenv("COOKIE_DOMAIN")
	c.SetCookie("session_id", "", -1, "/", domain, false, true)
	logrus.Infof("User logged out (Session: %s)", sessionID)

	c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
}

// GetCurrentUserHandler 取得目前用戶資訊
func GetCurrentUserHandler(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		logrus.WithFields(logrus.Fields{
			"client_ip": c.ClientIP(),
			"path":      c.Request.URL.Path,
		}).Warn("GetCurrentUser: No user_id in context")
		writeError(c, http.StatusUnauthorized, "unauthorized", "未登录")
		return
	}

	id, ok := userID.(int64)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"user_id_raw": userID,
			"user_id_type": fmt.Sprintf("%T", userID),
			"client_ip": c.ClientIP(),
		}).Error("GetCurrentUser: Invalid user_id type in context")
		writeServerError(c)
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id": id,
		"client_ip": c.ClientIP(),
	}).Info("GetCurrentUser: Looking up user by ID")

	user, err := database.GetUserByID(id)
	if err != nil {
		if err == database.ErrUserNotFound {
			logrus.WithFields(logrus.Fields{
				"user_id": id,
				"client_ip": c.ClientIP(),
			}).Warn("GetCurrentUser: User not found in database")
			writeError(c, http.StatusNotFound, "user_not_found", "用户不存在")
			return
		}
		logrus.WithFields(logrus.Fields{
			"user_id": id,
			"error": err.Error(),
			"client_ip": c.ClientIP(),
		}).Error("GetCurrentUser: Failed to get user profile")
		writeServerError(c)
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id": user.ID,
		"username": user.Username,
		"role": user.Role,
		"client_ip": c.ClientIP(),
	}).Info("GetCurrentUser: Successfully retrieved user")

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": user.CreatedAt,
			"last_login": user.LastLogin,
		},
	})
}

// ListUsersHandler 列出所有使用者 (僅管理員)
func ListUsersHandler(c *gin.Context) {
	users, err := database.ListUsers()
	if err != nil {
		logrus.Errorf("Failed to list users: %v", err)
		writeServerError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": len(users),
	})
}

type apiError struct {
	status  int
	code    string
	message string
}

func (e *apiError) Error() string {
	return e.message
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, models.NewErrorResponse(message, code, code))
}

func writeServerError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
		"服务器内部错误",
		"internal_error",
		"internal_error",
	))
}

func ensureUserAvailable(username, email string) error {
	if user, err := database.GetUserByUsername(username); err == nil && user != nil {
		return &apiError{status: http.StatusConflict, code: "username_exists", message: "用户名已存在"}
	} else if err != nil && err != database.ErrUserNotFound {
		return err
	}

	if user, err := database.GetUserByEmail(email); err == nil && user != nil {
		return &apiError{status: http.StatusConflict, code: "email_exists", message: "邮箱已被注册"}
	} else if err != nil && err != database.ErrUserNotFound {
		return err
	}

	return nil
}

// SendVerificationCodeHandler 发送验证码
func SendVerificationCodeHandler(c *gin.Context) {
	var req SendVerificationCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "请求参数无效: "+err.Error())
		return
	}

	// 验证 Turnstile token（必需）
	if turnstileService == nil {
		logrus.Error("Turnstile service not initialized")
		writeError(c, http.StatusInternalServerError, "service_error", "验证服务未初始化")
		return
	}

	success, err := turnstileService.VerifyToken(req.TurnstileToken, c.ClientIP())
	if err != nil || !success {
		logrus.Warnf("Turnstile verification failed for IP %s: %v", c.ClientIP(), err)
		writeError(c, http.StatusBadRequest, "captcha_failed", "人机验证失败，请重试")
		return
	}

	// 检查邮箱是否已注册
	if user, err := database.GetUserByEmail(req.Email); err == nil && user != nil {
		writeError(c, http.StatusConflict, "email_exists", "该邮箱已被注册")
		return
	} else if err != nil && err != database.ErrUserNotFound {
		logrus.Errorf("Failed to check email: %v", err)
		writeServerError(c)
		return
	}

	// 检查发送频率限制（60秒内只能发送一次）
	lastSentTime, err := database.GetRecentCodeSentTime(req.Email, "register")
	if err != nil {
		logrus.Errorf("Failed to check last sent time: %v", err)
		writeServerError(c)
		return
	}

	if !lastSentTime.IsZero() && time.Since(lastSentTime) < 60*time.Second {
		remainingSeconds := int(60 - time.Since(lastSentTime).Seconds())
		writeError(c, http.StatusTooManyRequests, "too_frequent",
			fmt.Sprintf("发送过于频繁，请在 %d 秒后重试", remainingSeconds))
		return
	}

	// 使旧验证码失效
	if err := database.InvalidateOldCodes(req.Email, "register"); err != nil {
		logrus.Warnf("Failed to invalidate old codes: %v", err)
	}

	// 创建新验证码
	verificationCode, err := database.CreateVerificationCode(req.Email, "register", c.ClientIP())
	if err != nil {
		logrus.Errorf("Failed to create verification code: %v", err)
		writeServerError(c)
		return
	}

	// 发送验证码邮件
	if err := emailService.SendVerificationCode(req.Email, verificationCode.Code); err != nil {
		logrus.Errorf("Failed to send verification email: %v", err)
		writeError(c, http.StatusInternalServerError, "email_send_failed", "验证码发送失败，请稍后重试")
		return
	}

	logrus.Infof("Verification code sent to %s", req.Email)
	
	// DEBUG模式下在控制台输出验证码（方便测试）
	if os.Getenv("DEBUG") == "true" {
		logrus.Warnf("🔑 DEBUG: Verification code for %s is: %s (expires in 10 minutes)", req.Email, verificationCode.Code)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "验证码已发送",
		"email":      req.Email,
		"expires_in": int(database.VerificationExpiry.Seconds()),
	})
}
