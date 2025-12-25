package handlers

import (
	"Curry2API-go/middleware"
	"Curry2API-go/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ListCursorSessionsHandler 列出所有 Cursor sessions
// @Summary 列出所有 Cursor 账号 sessions
// @Tags Cursor Session Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /admin/cursor/sessions [get]
func ListCursorSessionsHandler(c *gin.Context) {
	csm := middleware.GetCursorSessionManager()
	sessions := csm.ListSessions()
	stats := csm.GetStats()

	logrus.WithFields(logrus.Fields{
		"session_count": len(sessions),
		"stats":         stats,
	}).Debug("Listing Cursor sessions")

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"stats":    stats,
	})
}

// AddCursorSessionRequest 添加 Cursor session 请求
type AddCursorSessionRequest struct {
	Email        string            `json:"email" binding:"required"`
	SessionToken string            `json:"session_token" binding:"required"`
	ExpiresAt    string            `json:"expires_at,omitempty"`
	ExtraCookies map[string]string `json:"extra_cookies,omitempty"`
}

// AddCursorSessionHandler 添加新的 Cursor session
// @Summary 添加 Cursor 账号 session
// @Tags Cursor Session Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body AddCursorSessionRequest true "Session 信息"
// @Success 201 {object} map[string]interface{}
// @Router /admin/cursor/sessions [post]
func AddCursorSessionHandler(c *gin.Context) {
	var req AddCursorSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse := models.NewErrorResponse(
			"无效的请求格式",
			"validation_error",
			"invalid_request",
		)
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	if req.ExpiresAt != "" {
		parsed, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			errorResponse := models.NewErrorResponse(
				"expires_at 必须为 RFC3339 时间格式",
				"validation_error",
				"invalid_expires_at",
			)
			c.JSON(http.StatusBadRequest, errorResponse)
			return
		}
		expiresAt = parsed
	}

	csm := middleware.GetCursorSessionManager()
	if err := csm.AddSession(req.Email, req.SessionToken, expiresAt, req.ExtraCookies); err != nil {
		errorResponse := models.NewErrorResponse(
			err.Error(),
			"validation_error",
			"add_session_failed",
		)
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Cursor session 添加成功",
		"email":   req.Email,
	})
}

// RemoveCursorSessionHandler 删除 Cursor session
// @Summary 删除 Cursor 账号 session
// @Tags Cursor Session Admin
// @Security BearerAuth
// @Produce json
// @Param email path string true "账号邮箱"
// @Success 200 {object} map[string]interface{}
// @Router /admin/cursor/sessions/{email} [delete]
func RemoveCursorSessionHandler(c *gin.Context) {
	email := c.Param("email")

	csm := middleware.GetCursorSessionManager()
	if err := csm.RemoveSession(email); err != nil {
		statusCode := http.StatusNotFound
		errorResponse := models.NewErrorResponse(
			err.Error(),
			"not_found",
			"session_not_found",
		)
		c.JSON(statusCode, errorResponse)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cursor session 删除成功",
		"email":   email,
	})
}

// ValidateCursorSessionRequest 验证 session 请求
type ValidateCursorSessionRequest struct {
	Email string `json:"email" binding:"required"`
}

// ValidateCursorSessionHandler 手动验证 Cursor session
// @Summary 验证 Cursor session 有效性
// @Tags Cursor Session Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ValidateCursorSessionRequest true "验证请求"
// @Success 200 {object} map[string]interface{}
// @Router /admin/cursor/sessions/validate [post]
func ValidateCursorSessionHandler(c *gin.Context) {
	var req ValidateCursorSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse := models.NewErrorResponse(
			"无效的请求格式",
			"validation_error",
			"invalid_request",
		)
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	csm := middleware.GetCursorSessionManager()
	sessions := csm.ListSessions()

	var targetSession *middleware.CursorSessionInfo
	for _, session := range sessions {
		if session.Email == req.Email {
			targetSession = session
			break
		}
	}

	if targetSession == nil {
		errorResponse := models.NewErrorResponse(
			"Session 不存在",
			"not_found",
			"session_not_found",
		)
		c.JSON(http.StatusNotFound, errorResponse)
		return
	}

	// 执行验证（需要获取原始 session，而非掩码版本）
	// 注意：这里简化处理，实际应该有更好的方式
	isValid := csm.ValidateSession(c.Request.Context(), targetSession)

	c.JSON(http.StatusOK, gin.H{
		"email":    req.Email,
		"is_valid": isValid,
		"message": func() string {
			if isValid {
				return "Session 有效"
			}
			return "Session 无效或已过期"
		}(),
	})
}

// GetCursorSessionStatsHandler 获取 Cursor session 统计信息
// @Summary 获取统计信息
// @Tags Cursor Session Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /admin/cursor/sessions/stats [get]
func GetCursorSessionStatsHandler(c *gin.Context) {
	csm := middleware.GetCursorSessionManager()
	stats := csm.GetStats()

	c.JSON(http.StatusOK, stats)
}

// ReloadCursorSessionsHandler 重新加载 Cursor sessions
// @Summary 从数据库重新加载所有 sessions
// @Tags Cursor Session Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /admin/cursor/sessions/reload [post]
func ReloadCursorSessionsHandler(c *gin.Context) {
	csm := middleware.GetCursorSessionManager()
	
	if err := csm.ReloadFromDB(); err != nil {
		errorResponse := models.NewErrorResponse(
			fmt.Sprintf("重新加载失败: %v", err),
			"reload_error",
			"reload_failed",
		)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	stats := csm.GetStats()
	c.JSON(http.StatusOK, gin.H{
		"message": "Sessions 重新加载成功",
		"stats":   stats,
	})
}


// MigrateEncryptCursorSessionsHandler 迁移加密 Cursor sessions
// @Summary 将现有明文数据迁移到加密格式
// @Tags Cursor Session Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /admin/cursor/sessions/migrate-encrypt [post]
func MigrateEncryptCursorSessionsHandler(c *gin.Context) {
	csm := middleware.GetCursorSessionManager()
	
	migratedCount, err := csm.MigrateEncryptSessions()
	if err != nil {
		errorResponse := models.NewErrorResponse(
			fmt.Sprintf("迁移失败: %v", err),
			"migration_error",
			"migration_failed",
		)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "数据加密迁移完成",
		"migrated_count": migratedCount,
	})
}
