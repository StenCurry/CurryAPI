package handlers

import (
	"Curry2API-go/database"
	"Curry2API-go/models"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetReferralCodeHandler retrieves the current user's referral code and link
// GET /api/referral/code
// Requirements: 4.3
func GetReferralCodeHandler(c *gin.Context) {
	// Extract user_id from session context
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			"User not authenticated",
			"authentication_error",
			"missing_user_id",
		))
		return
	}

	userID, ok := userIDInterface.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Invalid user ID format",
			"internal_error",
			"invalid_user_id_type",
		))
		return
	}

	// Get user balance (which contains the referral code)
	balance, err := database.GetUserBalance(userID)
	if err != nil {
		if err == database.ErrBalanceNotFound {
			// Auto-create balance record for existing users who don't have one
			logrus.WithField("user_id", userID).Info("Creating balance record for existing user (referral)")
			balance, err = database.CreateUserBalance(userID)
			if err != nil {
				logrus.WithError(err).WithField("user_id", userID).Error("Failed to create balance for existing user")
				c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
					"Failed to create balance record",
					"internal_error",
					"database_error",
				))
				return
			}
		} else {
			logrus.WithError(err).WithField("user_id", userID).Error("Failed to get user balance for referral code")
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				"Failed to retrieve referral code",
				"internal_error",
				"database_error",
			))
			return
		}
	}

	// Generate referral link (points to login page with referral code)
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:5173"
	}
	referralLink := fmt.Sprintf("%s/login?ref=%s", baseURL, balance.ReferralCode)

	c.JSON(http.StatusOK, gin.H{
		"referral_code": balance.ReferralCode,
		"referral_link": referralLink,
	})
}


// GetReferralStatsHandler retrieves referral statistics for the current user
// GET /api/referral/stats
// Requirements: 7.1, 7.2
func GetReferralStatsHandler(c *gin.Context) {
	// Extract user_id from session context
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			"User not authenticated",
			"authentication_error",
			"missing_user_id",
		))
		return
	}

	userID, ok := userIDInterface.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Invalid user ID format",
			"internal_error",
			"invalid_user_id_type",
		))
		return
	}

	// Get referral statistics
	stats, err := database.GetReferralStats(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get referral stats")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve referral statistics",
			"internal_error",
			"database_error",
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_referrals": stats.TotalReferrals,
		"total_bonus":     stats.TotalBonus,
	})
}

// GetReferralListHandler retrieves the list of referred users for the current user
// GET /api/referral/list
// Query params: limit (default 20, max 100), offset (default 0)
// Requirements: 7.3
func GetReferralListHandler(c *gin.Context) {
	// Extract user_id from session context
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			"User not authenticated",
			"authentication_error",
			"missing_user_id",
		))
		return
	}

	userID, ok := userIDInterface.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Invalid user ID format",
			"internal_error",
			"invalid_user_id_type",
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

	// Get referral list
	referrals, total, err := database.GetReferralList(userID, limit, offset)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get referral list")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve referral list",
			"internal_error",
			"database_error",
		))
		return
	}

	// Format referrals for response
	formattedReferrals := make([]gin.H, 0, len(referrals))
	for _, ref := range referrals {
		formattedReferrals = append(formattedReferrals, gin.H{
			"user_id":       ref.UserID,
			"username":      ref.Username,
			"email":         maskEmail(ref.Email),
			"registered_at": ref.RegisteredAt,
			"bonus_amount":  ref.BonusAmount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"referrals": formattedReferrals,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// maskEmail masks an email address for privacy (shows first 2 chars and domain)
func maskEmail(email string) string {
	if len(email) < 5 {
		return "****"
	}
	
	atIndex := -1
	for i, c := range email {
		if c == '@' {
			atIndex = i
			break
		}
	}
	
	if atIndex <= 0 {
		return "****"
	}
	
	// Show first 2 characters, then mask, then show domain
	prefix := email[:2]
	if atIndex > 2 {
		prefix += "****"
	}
	return prefix + email[atIndex:]
}
