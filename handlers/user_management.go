package handlers

import (
	"Curry2API-go/database"
	"Curry2API-go/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetUserHandler 获取单个用户信息
func GetUserHandler(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"无效的用户ID",
			"invalid_request",
			"invalid_user_id",
		))
		return
	}

	user, err := database.GetUserByID(userID)
	if err != nil {
		if err == database.ErrUserNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				"用户不存在",
				"not_found",
				"user_not_found",
			))
			return
		}
		logrus.Errorf("Failed to get user: %v", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"获取用户信息失败",
			"internal_error",
			"get_user_failed",
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": user.CreatedAt,
			"last_login": user.LastLogin,
			"is_active":  user.IsActive,
		},
	})
}

// UpdateUserRoleRequest 更新用户角色请求
type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=user admin"`
}

// UpdateUserRoleHandler 更新用户角色
func UpdateUserRoleHandler(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"无效的用户ID",
			"invalid_request",
			"invalid_user_id",
		))
		return
	}

	var req UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"请求参数无效",
			"invalid_request",
			"invalid_parameters",
		))
		return
	}

	// 检查用户是否存在
	user, err := database.GetUserByID(userID)
	if err != nil {
		if err == database.ErrUserNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				"用户不存在",
				"not_found",
				"user_not_found",
			))
			return
		}
		logrus.Errorf("Failed to get user: %v", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"获取用户信息失败",
			"internal_error",
			"get_user_failed",
		))
		return
	}

	// 更新角色
	if err := database.UpdateUserRole(userID, req.Role); err != nil {
		logrus.Errorf("Failed to update user role: %v", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"更新用户角色失败",
			"internal_error",
			"update_role_failed",
		))
		return
	}

	logrus.Infof("User %s role updated from %s to %s", user.Username, user.Role, req.Role)

	c.JSON(http.StatusOK, gin.H{
		"message": "用户角色更新成功",
		"user": gin.H{
			"id":       userID,
			"username": user.Username,
			"role":     req.Role,
		},
	})
}

// ToggleUserStatusHandler 启用/禁用用户
func ToggleUserStatusHandler(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"无效的用户ID",
			"invalid_request",
			"invalid_user_id",
		))
		return
	}

	// 检查用户是否存在
	user, err := database.GetUserByID(userID)
	if err != nil {
		if err == database.ErrUserNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				"用户不存在",
				"not_found",
				"user_not_found",
			))
			return
		}
		logrus.Errorf("Failed to get user: %v", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"获取用户信息失败",
			"internal_error",
			"get_user_failed",
		))
		return
	}

	// 切换状态
	newStatus := !user.IsActive
	if err := database.UpdateUserStatus(userID, newStatus); err != nil {
		logrus.Errorf("Failed to update user status: %v", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"更新用户状态失败",
			"internal_error",
			"update_status_failed",
		))
		return
	}

	statusText := "禁用"
	if newStatus {
		statusText = "启用"
	}
	logrus.Infof("User %s status changed to %s (API keys also %s)", user.Username, statusText, statusText)

	c.JSON(http.StatusOK, gin.H{
		"message":   "用户状态更新成功",
		"is_active": newStatus,
	})
}

// DeleteUserHandler 删除用户（软删除）
func DeleteUserHandler(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"无效的用户ID",
			"invalid_request",
			"invalid_user_id",
		))
		return
	}

	// 检查用户是否存在
	user, err := database.GetUserByID(userID)
	if err != nil {
		if err == database.ErrUserNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				"用户不存在",
				"not_found",
				"user_not_found",
			))
			return
		}
		logrus.Errorf("Failed to get user: %v", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"获取用户信息失败",
			"internal_error",
			"get_user_failed",
		))
		return
	}

	// 删除用户（软删除）
	if err := database.DeleteUser(userID); err != nil {
		logrus.Errorf("Failed to delete user: %v", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"删除用户失败",
			"internal_error",
			"delete_user_failed",
		))
		return
	}

	logrus.Infof("User %s deleted", user.Username)

	c.JSON(http.StatusOK, gin.H{
		"message": "用户删除成功",
	})
}
