package handlers

import (
	"Curry2API-go/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// QuotaUpdateRequest represents a request to update session quota
type QuotaUpdateRequest struct {
	Email    string `json:"email" binding:"required"`
	NewLimit int64  `json:"new_limit" binding:"required,gt=0"`
}

// GetQuotaStats returns quota statistics for all sessions
// GET /api/quota/stats
func (h *Handler) GetQuotaStats(c *gin.Context) {
	quotaMgr := middleware.GetQuotaManager(&h.config.Quota)
	
	if !quotaMgr.IsEnabled() {
		c.JSON(http.StatusOK, gin.H{
			"enabled": false,
			"message": "Quota management is disabled",
		})
		return
	}
	
	stats, err := quotaMgr.GetQuotaStats()
	if err != nil {
		logrus.WithError(err).Error("Failed to get quota statistics")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve quota statistics",
		})
		return
	}
	
	c.JSON(http.StatusOK, stats)
}

// UpdateQuotaLimit updates the quota limit for a specific session
// PUT /api/quota/update
func (h *Handler) UpdateQuotaLimit(c *gin.Context) {
	quotaMgr := middleware.GetQuotaManager(&h.config.Quota)
	
	if !quotaMgr.IsEnabled() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Quota management is disabled",
		})
		return
	}
	
	var req QuotaUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid quota update request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}
	
	if err := quotaMgr.UpdateSessionQuota(req.Email, req.NewLimit); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"email":     req.Email,
			"new_limit": req.NewLimit,
		}).Error("Failed to update session quota")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Quota limit updated successfully",
		"email":   req.Email,
		"new_limit": req.NewLimit,
	})
}

// ResetQuotas manually triggers quota reset for all sessions
// POST /api/quota/reset
func (h *Handler) ResetQuotas(c *gin.Context) {
	quotaMgr := middleware.GetQuotaManager(&h.config.Quota)
	
	if !quotaMgr.IsEnabled() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Quota management is disabled",
		})
		return
	}
	
	if err := quotaMgr.ResetDailyQuotas(); err != nil {
		logrus.WithError(err).Error("Failed to reset quotas")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to reset quotas",
		})
		return
	}
	
	// Get updated stats
	stats, _ := quotaMgr.GetQuotaStats()
	
	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"message":        "All session quotas have been reset",
		"sessions_reset": stats.TotalSessions,
	})
}
