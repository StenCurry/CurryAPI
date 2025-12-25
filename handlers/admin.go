package handlers

import (
	"Curry2API-go/database"
	"Curry2API-go/middleware"
	"Curry2API-go/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// maskKey 掩码密钥（保留前4后4，中间用*代替）
func maskKey(key string) string {
	keyLen := len(key)
	if keyLen <= 8 {
		return key // 太短不掩码
	}
	return key[:4] + strings.Repeat("*", keyLen-8) + key[keyLen-4:]
}

// AdminAuth 管理员认证中间件（支持会话认证和 Bearer token）
func AdminAuth() gin.HandlerFunc {
	km := middleware.GetKeyManager()

	return func(c *gin.Context) {
		// 方式1: 尝试会话 Cookie 认证
		sessionID, err := c.Cookie("session_id")
		logrus.Debugf("AdminAuth: sessionID=%s, err=%v", sessionID, err)
		
		if err == nil && sessionID != "" {
			// 使用 SessionAuth 的验证逻辑
			session, err := middleware.ValidateSession(sessionID)
			logrus.Debugf("AdminAuth: ValidateSession result: session=%+v, err=%v", session, err)
			
			if err == nil {
				// 任何登录用户都可以访问（不再限制管理员）
				logrus.Debugf("AdminAuth: User role=%s", session.Role)
				c.Set("user_id", session.UserID)
				c.Set("username", session.Username)
				c.Set("role", session.Role)
				c.Set("session_id", session.ID)
				c.Next()
				return
			}
		}

		// 方式2: 尝试 Bearer token 认证
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == km.GetAdminToken() {
				c.Set("user_id", int64(-1))
				c.Set("username", "admin")
				c.Set("role", "admin")
				c.Next()
				return
			}
		}

		// 两种认证方式都失败
		errorResponse := models.NewErrorResponse(
			"需要管理员权限，请先登录或提供有效的管理员令牌",
			"admin_auth_error",
			"unauthorized",
		)
		c.JSON(http.StatusUnauthorized, errorResponse)
		c.Abort()
	}
}

// ListKeysHandler 列出当前用户的密钥
// @Summary 列出当前用户的API密钥
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /admin/keys [get]
func ListKeysHandler(c *gin.Context) {
	// 获取当前用户ID和角色
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse := models.NewErrorResponse(
			"无法获取用户信息",
			"internal_error",
			"user_not_found",
		)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	role, roleExists := c.Get("role")
	if !roleExists {
		role = "user" // 默认为普通用户
	}

	km := middleware.GetKeyManager()
	
	// 如果是管理员角色，显示所有密钥；否则只显示用户自己的密钥
	var keys []*middleware.KeyInfo
	userIDInt := userID.(int64)
	roleStr := role.(string)
	
	logrus.Debugf("ListKeysHandler: userID=%d, role=%s", userIDInt, roleStr)
	
	if roleStr == "admin" {
		keys = km.ListKeys()
		logrus.Debugf("ListKeysHandler: Admin user, returning all %d keys", len(keys))
	} else {
		keys = km.ListKeysByUser(userIDInt)
		logrus.Debugf("ListKeysHandler: Regular user %d, returning %d keys", userIDInt, len(keys))
	}

	c.JSON(http.StatusOK, gin.H{
		"total": len(keys),
		"keys":  keys,
	})
}

// AddKeyRequest 添加密钥请求
type AddKeyRequest struct {
	Key           string    `json:"key" binding:"required"`
	TokenName     string    `json:"token_name,omitempty"`
	QuotaLimit    *float64  `json:"quota_limit,omitempty"`    // Quota limit in USD, nil means unlimited
	ExpiresAt     *string   `json:"expires_at,omitempty"`     // ISO date string, nil means never expires
	AllowedModels []string  `json:"allowed_models,omitempty"` // Allowed models, nil/empty means all models
}

// AddKeyHandler 添加新密钥
// @Summary 添加新API密钥
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body AddKeyRequest true "密钥信息"
// @Success 201 {object} map[string]interface{}
// @Router /admin/keys [post]
func AddKeyHandler(c *gin.Context) {
	var req AddKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse := models.NewErrorResponse(
			"无效的请求格式",
			"validation_error",
			"invalid_request",
		)
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse := models.NewErrorResponse(
			"无法获取用户信息",
			"internal_error",
			"user_not_found",
		)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	// Parse expiration date if provided
	var expiresAt *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		parsed, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			errorResponse := models.NewErrorResponse(
				"无效的过期时间格式，请使用 ISO 8601 格式",
				"validation_error",
				"invalid_expires_at",
			)
			c.JSON(http.StatusBadRequest, errorResponse)
			return
		}
		expiresAt = &parsed
	}

	// Build options for the new key
	opts := &database.APIKeyOptions{
		QuotaLimit:    req.QuotaLimit,
		ExpiresAt:     expiresAt,
		AllowedModels: req.AllowedModels,
	}

	userIDInt := userID.(int64)
	var userIDPtr *int64
	if userIDInt > 0 {
		userIDPtr = &userIDInt
	}

	// Use the new function that supports all options
	if err := database.AddAPIKeyWithOptions(req.Key, userIDPtr, req.TokenName, opts); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			errorResponse := models.NewErrorResponse(
				"密钥已存在",
				"validation_error",
				"duplicate_key",
			)
			c.JSON(http.StatusBadRequest, errorResponse)
			return
		}
		errorResponse := models.NewErrorResponse(
			err.Error(),
			"internal_error",
			"add_key_failed",
		)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	// Reload keys in memory
	km := middleware.GetKeyManager()
	km.ReloadKeys()

	c.JSON(http.StatusCreated, gin.H{
		"message":        "密钥添加成功",
		"key":            maskKey(req.Key),
		"token_name":     req.TokenName,
		"quota_limit":    req.QuotaLimit,
		"expires_at":     req.ExpiresAt,
		"allowed_models": req.AllowedModels,
	})
}

// ToggleKeyStatusHandler 切换密钥的启用/禁用状态
// @Summary 切换API密钥状态
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param key path string true "要切换状态的密钥"
// @Success 200 {object} map[string]interface{}
// @Router /admin/keys/{key}/toggle [put]
func ToggleKeyStatusHandler(c *gin.Context) {
	key := c.Param("key")

	km := middleware.GetKeyManager()
	if err := km.ToggleKeyStatus(key); err != nil {
		if keyErr, ok := err.(*middleware.KeyError); ok {
			statusCode := http.StatusBadRequest
			if keyErr.Code == "key_not_found" {
				statusCode = http.StatusNotFound
			}
			errorResponse := models.NewErrorResponse(
				keyErr.Message,
				"validation_error",
				keyErr.Code,
			)
			c.JSON(statusCode, errorResponse)
			return
		}
		errorResponse := models.NewErrorResponse(
			err.Error(),
			"internal_error",
			"toggle_key_failed",
		)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "密钥状态切换成功",
		"key":     maskKey(key),
	})
}

// RemoveKeyHandler 删除密钥
// @Summary 删除API密钥
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param key path string true "要删除的密钥"
// @Success 200 {object} map[string]interface{}
// @Router /admin/keys/{key} [delete]
func RemoveKeyHandler(c *gin.Context) {
	key := c.Param("key")

	km := middleware.GetKeyManager()
	if err := km.RemoveKey(key); err != nil {
		if keyErr, ok := err.(*middleware.KeyError); ok {
			statusCode := http.StatusBadRequest
			if keyErr.Code == "key_not_found" {
				statusCode = http.StatusNotFound
			}
			errorResponse := models.NewErrorResponse(
				keyErr.Message,
				"validation_error",
				keyErr.Code,
			)
			c.JSON(statusCode, errorResponse)
			return
		}
		errorResponse := models.NewErrorResponse(
			err.Error(),
			"internal_error",
			"remove_key_failed",
		)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "密钥删除成功",
		"key":     maskKey(key),
	})
}

// UpdateKeyNameRequest 更新密钥名称请求
type UpdateKeyNameRequest struct {
	Name string `json:"name" binding:"required"`
}

// UpdateKeyNameHandler 更新密钥名称
// @Summary 更新API密钥名称
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param key path string true "要更新的密钥"
// @Param request body UpdateKeyNameRequest true "新名称"
// @Success 200 {object} map[string]interface{}
// @Router /admin/keys/{key}/name [put]
func UpdateKeyNameHandler(c *gin.Context) {
	key := c.Param("key")
	
	var req UpdateKeyNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse := models.NewErrorResponse(
			"无效的请求格式",
			"validation_error",
			"invalid_request",
		)
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Validate name length and characters
	if len(req.Name) > 255 {
		errorResponse := models.NewErrorResponse(
			"名称长度不能超过255个字符",
			"validation_error",
			"name_too_long",
		)
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	km := middleware.GetKeyManager()
	if err := km.UpdateKeyName(key, req.Name); err != nil {
		if keyErr, ok := err.(*middleware.KeyError); ok {
			statusCode := http.StatusBadRequest
			if keyErr.Code == "key_not_found" {
				statusCode = http.StatusNotFound
			}
			errorResponse := models.NewErrorResponse(
				keyErr.Message,
				"validation_error",
				keyErr.Code,
			)
			c.JSON(statusCode, errorResponse)
			return
		}
		errorResponse := models.NewErrorResponse(
			err.Error(),
			"internal_error",
			"update_key_name_failed",
		)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "密钥名称更新成功",
		"key":     maskKey(key),
		"name":    req.Name,
	})
}

// ============================================
// Admin Balance Management Handlers
// ============================================

// AdjustBalanceRequest represents the request body for adjusting user balance
type AdjustBalanceRequest struct {
	UserID int64   `json:"user_id" binding:"required"`
	Amount float64 `json:"amount" binding:"required"`
	Reason string  `json:"reason" binding:"required"`
}

// AdjustUserBalanceHandler adjusts a user's balance (add or deduct)
// POST /admin/balance/adjust
// Requirements: 8.1, 8.2, 8.3
func AdjustUserBalanceHandler(c *gin.Context) {
	// Get admin user ID
	adminIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			"Admin not authenticated",
			"authentication_error",
			"missing_admin_id",
		))
		return
	}

	adminID, ok := adminIDInterface.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Invalid admin ID format",
			"internal_error",
			"invalid_admin_id_type",
		))
		return
	}

	// Check if user is admin
	role, roleExists := c.Get("role")
	if !roleExists || role.(string) != "admin" {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			"Admin privileges required",
			"authorization_error",
			"admin_required",
		))
		return
	}

	// Parse request body
	var req AdjustBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid request format",
			"validation_error",
			"invalid_request",
		))
		return
	}

	// Validate amount is not zero
	if req.Amount == 0 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Amount cannot be zero",
			"validation_error",
			"invalid_amount",
		))
		return
	}

	// Validate reason is not empty
	if strings.TrimSpace(req.Reason) == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Reason is required",
			"validation_error",
			"missing_reason",
		))
		return
	}

	// Build description with admin info
	description := "Admin adjustment: " + req.Reason

	// Add balance (positive or negative amount)
	transaction, err := database.AddBalance(
		req.UserID,
		req.Amount,
		description,
		&adminID,
		nil,
		database.TransactionTypeAdminAdjust,
	)
	if err != nil {
		if err == database.ErrBalanceNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				"User balance not found",
				"not_found_error",
				"balance_not_found",
			))
			return
		}
		logrus.WithError(err).WithFields(logrus.Fields{
			"user_id":  req.UserID,
			"admin_id": adminID,
			"amount":   req.Amount,
		}).Error("Failed to adjust user balance")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to adjust balance",
			"internal_error",
			"database_error",
		))
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id":       req.UserID,
		"admin_id":      adminID,
		"amount":        req.Amount,
		"reason":        req.Reason,
		"balance_after": transaction.BalanceAfter,
	}).Info("Admin adjusted user balance")

	c.JSON(http.StatusOK, gin.H{
		"message":       "Balance adjusted successfully",
		"user_id":       req.UserID,
		"amount":        req.Amount,
		"balance_after": transaction.BalanceAfter,
		"transaction_id": transaction.ID,
	})
}


// GetAllUserBalancesHandler retrieves all user balances with pagination
// GET /admin/balance/users
// Query params: limit (default 20, max 100), offset (default 0)
// Requirements: 8.3
func GetAllUserBalancesHandler(c *gin.Context) {
	// Check if user is admin
	role, roleExists := c.Get("role")
	if !roleExists || role.(string) != "admin" {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			"Admin privileges required",
			"authorization_error",
			"admin_required",
		))
		return
	}

	// Parse pagination parameters
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 100 {
				limit = 100
			}
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get all user balances with user info
	balances, total, err := database.GetAllUserBalancesWithInfo(limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to get all user balances")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve user balances",
			"internal_error",
			"database_error",
		))
		return
	}

	// Format balances for response
	formattedBalances := make([]gin.H, 0, len(balances))
	for _, balance := range balances {
		formattedBalances = append(formattedBalances, gin.H{
			"user_id":         balance.UserID,
			"username":        balance.Username,
			"email":           balance.Email,
			"balance":         balance.Balance,
			"status":          balance.Status,
			"referral_code":   balance.ReferralCode,
			"total_consumed":  balance.TotalConsumed,
			"total_recharged": balance.TotalRecharged,
			"created_at":      balance.CreatedAt,
			"updated_at":      balance.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"users":  formattedBalances,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}
