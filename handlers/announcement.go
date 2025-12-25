package handlers

import (
	"Curry2API-go/database"
	"Curry2API-go/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// CreateAnnouncementRequest 创建公告请求
type CreateAnnouncementRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// CreateAnnouncementHandler 创建新公告
// @Summary 创建新公告
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateAnnouncementRequest true "公告信息"
// @Success 201 {object} database.Announcement
// @Router /admin/announcements [post]
func CreateAnnouncementHandler(c *gin.Context) {
	var req CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse := models.NewErrorResponse(
			"标题和内容不能为空",
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

	// 创建公告
	announcement, err := database.CreateAnnouncement(req.Title, req.Content, userID.(int64))
	if err != nil {
		logrus.WithError(err).Error("Failed to create announcement")
		errorResponse := models.NewErrorResponse(
			"服务器内部错误",
			"internal_error",
			"create_announcement_failed",
		)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	c.JSON(http.StatusCreated, announcement)
}

// ListAllAnnouncementsHandler 获取所有公告列表
// @Summary 获取所有公告列表
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param limit query int false "返回数量" default(10)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {object} map[string]interface{}
// @Router /admin/announcements [get]
func ListAllAnnouncementsHandler(c *gin.Context) {
	// 获取分页参数
	limit := 10
	offset := 0
	
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// 获取公告列表
	announcements, total, err := database.GetAnnouncements(limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to get announcements")
		errorResponse := models.NewErrorResponse(
			"服务器内部错误",
			"internal_error",
			"get_announcements_failed",
		)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":         total,
		"announcements": announcements,
	})
}

// DeleteAnnouncementHandler 删除公告
// @Summary 删除公告
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param id path int true "公告ID"
// @Success 200 {object} map[string]interface{}
// @Router /admin/announcements/{id} [delete]
func DeleteAnnouncementHandler(c *gin.Context) {
	// 获取公告ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errorResponse := models.NewErrorResponse(
			"无效的公告ID",
			"validation_error",
			"invalid_id",
		)
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// 删除公告
	err = database.DeleteAnnouncement(id)
	if err == database.ErrAnnouncementNotFound {
		errorResponse := models.NewErrorResponse(
			"公告不存在",
			"not_found",
			"announcement_not_found",
		)
		c.JSON(http.StatusNotFound, errorResponse)
		return
	}
	if err != nil {
		logrus.WithError(err).Error("Failed to delete announcement")
		errorResponse := models.NewErrorResponse(
			"服务器内部错误",
			"internal_error",
			"delete_announcement_failed",
		)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "公告删除成功",
	})
}

// ListAnnouncementsHandler 获取公告列表（包含阅读状态）
// @Summary 获取公告列表
// @Tags User
// @Security BearerAuth
// @Produce json
// @Param limit query int false "返回数量" default(10)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {object} map[string]interface{}
// @Router /announcements [get]
func ListAnnouncementsHandler(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse := models.NewErrorResponse(
			"需要登录才能访问此资源",
			"unauthorized",
			"user_not_found",
		)
		c.JSON(http.StatusUnauthorized, errorResponse)
		return
	}

	// 获取分页参数
	limit := 10
	offset := 0
	
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// 获取带阅读状态的公告列表
	announcements, total, err := database.GetAnnouncementsWithReadStatus(userID.(int64), limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to get announcements with read status")
		errorResponse := models.NewErrorResponse(
			"服务器内部错误",
			"internal_error",
			"get_announcements_failed",
		)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":         total,
		"announcements": announcements,
	})
}

// GetUnreadCountHandler 获取未读公告数量
// @Summary 获取未读公告数量
// @Tags User
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /announcements/unread-count [get]
func GetUnreadCountHandler(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse := models.NewErrorResponse(
			"需要登录才能访问此资源",
			"unauthorized",
			"user_not_found",
		)
		c.JSON(http.StatusUnauthorized, errorResponse)
		return
	}

	// 获取未读数量
	count, err := database.GetUnreadCount(userID.(int64))
	if err != nil {
		logrus.WithError(err).Error("Failed to get unread count")
		errorResponse := models.NewErrorResponse(
			"服务器内部错误",
			"internal_error",
			"get_unread_count_failed",
		)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": count,
	})
}

// MarkAsReadHandler 标记公告为已读
// @Summary 标记公告为已读
// @Tags User
// @Security BearerAuth
// @Produce json
// @Param id path int true "公告ID"
// @Success 200 {object} map[string]interface{}
// @Router /announcements/{id}/read [post]
func MarkAsReadHandler(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse := models.NewErrorResponse(
			"需要登录才能访问此资源",
			"unauthorized",
			"user_not_found",
		)
		c.JSON(http.StatusUnauthorized, errorResponse)
		return
	}

	// 获取公告ID
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errorResponse := models.NewErrorResponse(
			"无效的公告ID",
			"validation_error",
			"invalid_id",
		)
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// 检查公告是否存在
	_, err = database.GetAnnouncementByID(id)
	if err == database.ErrAnnouncementNotFound {
		errorResponse := models.NewErrorResponse(
			"公告不存在",
			"not_found",
			"announcement_not_found",
		)
		c.JSON(http.StatusNotFound, errorResponse)
		return
	}
	if err != nil {
		logrus.WithError(err).Error("Failed to get announcement")
		errorResponse := models.NewErrorResponse(
			"服务器内部错误",
			"internal_error",
			"get_announcement_failed",
		)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	// 标记为已读
	err = database.MarkAsRead(id, userID.(int64))
	if err != nil {
		logrus.WithError(err).Error("Failed to mark as read")
		errorResponse := models.NewErrorResponse(
			"服务器内部错误",
			"internal_error",
			"mark_as_read_failed",
		)
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "公告已标记为已读",
	})
}
