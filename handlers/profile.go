package handlers

import (
	"Curry2API-go/database"
	"Curry2API-go/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// UpdateUsernameRequest 更新用户名请求
type UpdateUsernameRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
}

// UpdatePasswordRequest 更新密码请求
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// UpdateUsernameHandler 更新用户名
func UpdateUsernameHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			"未登录",
			"unauthorized",
			"unauthorized",
		))
		return
	}

	var req UpdateUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"请求参数无效",
			"invalid_request",
			"invalid_parameters",
		))
		return
	}

	// 检查用户名是否已存在
	newUsername := strings.TrimSpace(req.Username)
	if existingUser, err := database.GetUserByUsername(newUsername); err == nil && existingUser != nil {
		if existingUser.ID != userID.(int64) {
			c.JSON(http.StatusConflict, models.NewErrorResponse(
				"用户名已被使用",
				"username_exists",
				"username_exists",
			))
			return
		}
	}

	// 更新用户名
	if err := database.UpdateUsername(userID.(int64), newUsername); err != nil {
		logrus.Errorf("Failed to update username: %v", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"更新用户名失败",
			"internal_error",
			"update_failed",
		))
		return
	}

	logrus.Infof("User %d updated username to %s", userID.(int64), newUsername)

	c.JSON(http.StatusOK, gin.H{
		"message":  "用户名更新成功",
		"username": newUsername,
	})
}

// UpdatePasswordHandler 更新密码
func UpdatePasswordHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			"未登录",
			"unauthorized",
			"unauthorized",
		))
		return
	}

	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"请求参数无效",
			"invalid_request",
			"invalid_parameters",
		))
		return
	}

	// 获取用户信息
	user, err := database.GetUserByID(userID.(int64))
	if err != nil {
		logrus.Errorf("Failed to get user: %v", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"获取用户信息失败",
			"internal_error",
			"get_user_failed",
		))
		return
	}

	// 验证旧密码
	if !database.ValidatePassword(user, req.OldPassword) {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"原密码错误",
			"invalid_password",
			"invalid_old_password",
		))
		return
	}

	// 更新密码
	if err := database.UpdateUserPassword(userID.(int64), req.NewPassword); err != nil {
		logrus.Errorf("Failed to update password: %v", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"更新密码失败",
			"internal_error",
			"update_failed",
		))
		return
	}

	logrus.Infof("User %d updated password", userID.(int64))

	c.JSON(http.StatusOK, gin.H{
		"message": "密码更新成功",
	})
}
