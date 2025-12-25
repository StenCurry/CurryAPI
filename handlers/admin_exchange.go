package handlers

import (
	"Curry2API-go/database"
	"Curry2API-go/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AdminGetAllExchangesHandler retrieves all exchange records with optional filters
// GET /api/admin/exchanges
// Query params:
//   - user_id: filter by user ID (optional)
//   - start_date: filter by start date in RFC3339 format (optional)
//   - end_date: filter by end date in RFC3339 format (optional)
//   - limit: pagination limit (default 20, max 100)
//   - offset: pagination offset (default 0)
// Requirements: 6.1, 6.2, 6.3, 6.4
func AdminGetAllExchangesHandler(c *gin.Context) {
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

	// Parse optional user_id filter
	var userID *int64
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		parsedUserID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Invalid user_id format",
				"validation_error",
				"invalid_user_id",
			))
			return
		}
		userID = &parsedUserID
	}

	// Parse optional date filters
	var startDate, endDate *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		parsed, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			// Try parsing date-only format
			parsed, err = time.Parse("2006-01-02", startDateStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, models.NewErrorResponse(
					"Invalid start_date format. Use RFC3339 or YYYY-MM-DD",
					"validation_error",
					"invalid_start_date",
				))
				return
			}
		}
		startDate = &parsed
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		parsed, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			// Try parsing date-only format
			parsed, err = time.Parse("2006-01-02", endDateStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, models.NewErrorResponse(
					"Invalid end_date format. Use RFC3339 or YYYY-MM-DD",
					"validation_error",
					"invalid_end_date",
				))
				return
			}
			// Set end date to end of day
			parsed = parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
		endDate = &parsed
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

	// Get exchange records from database
	records, total, err := database.GetAllExchangeRecords(userID, startDate, endDate, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to get all exchange records")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve exchange records",
			"internal_error",
			"database_error",
		))
		return
	}

	// Format records for response
	formattedRecords := make([]gin.H, 0, len(records))
	for _, record := range records {
		formattedRecords = append(formattedRecords, gin.H{
			"id":                record.ID,
			"user_id":           record.UserID,
			"username":          record.Username,
			"email":             record.Email,
			"game_coins_amount": record.GameCoinsAmount,
			"usd_amount":        record.USDAmount,
			"exchange_rate":     record.ExchangeRate,
			"status":            record.Status,
			"created_at":        record.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"records": formattedRecords,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// AdminGetExchangeStatsHandler retrieves exchange statistics
// GET /api/admin/exchanges/stats
// Requirements: 6.5
func AdminGetExchangeStatsHandler(c *gin.Context) {
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

	// Get exchange statistics from database
	stats, err := database.GetExchangeStats()
	if err != nil {
		logrus.WithError(err).Error("Failed to get exchange statistics")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve exchange statistics",
			"internal_error",
			"database_error",
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_count": stats.TotalCount,
		"total_usd":   stats.TotalUSD,
	})
}
