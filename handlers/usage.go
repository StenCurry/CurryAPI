package handlers

import (
	"Curry2API-go/database"
	"Curry2API-go/models"
	"Curry2API-go/services"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetUserUsageStats retrieves usage statistics for the authenticated user
func GetUserUsageStats(c *gin.Context) {
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

	// Parse query parameters for filtering
	filter := database.UsageFilter{}

	// Parse start_date
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Invalid start_date format. Expected YYYY-MM-DD",
				"invalid_request_error",
				"invalid_date_format",
			))
			return
		}
		filter.StartDate = &startDate
	}

	// Parse end_date
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Invalid end_date format. Expected YYYY-MM-DD",
				"invalid_request_error",
				"invalid_date_format",
			))
			return
		}
		// Set to end of day
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		filter.EndDate = &endDate
	}

	// Parse model filter
	if model := c.Query("model"); model != "" {
		filter.Model = &model
	}

	// Get usage statistics from database
	stats, err := database.GetUserUsageStats(userID, filter)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get user usage stats")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve usage statistics",
			"internal_error",
			"database_error",
		))
		return
	}

	// Check if user has any usage data
	if stats.TotalRequests == 0 {
		c.JSON(http.StatusOK, gin.H{
			"total_requests":     0,
			"total_tokens":       0,
			"prompt_tokens":      0,
			"completion_tokens":  0,
			"by_model":           []interface{}{},
			"recent_calls":       []interface{}{},
			"message":            "No usage data found. Start making API calls to see your statistics here.",
		})
		return
	}

	// Format response with all required fields
	response := gin.H{
		"total_requests":    stats.TotalRequests,
		"total_tokens":      stats.TotalTokens,
		"prompt_tokens":     stats.PromptTokens,
		"completion_tokens": stats.CompletionTokens,
		"by_model":          formatModelBreakdown(stats.ByModel),
		"recent_calls":      formatRecentCalls(stats.RecentCalls),
	}

	c.JSON(http.StatusOK, response)
}

// GetUserRecentCalls retrieves recent API calls for the authenticated user
func GetUserRecentCalls(c *gin.Context) {
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

	// Parse limit parameter (default 50, max 100)
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 100 {
				limit = 100
			}
		}
	}

	// Parse offset parameter for pagination
	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Create filter for recent calls
	filter := database.UsageFilter{
		Limit:  limit,
		Offset: offset,
	}

	// Query recent usage records
	records, err := database.GetUsageRecordsByUser(userID, filter)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get recent calls")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve recent calls",
			"internal_error",
			"database_error",
		))
		return
	}

	// Check if user has any usage records
	if len(records) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"calls":   []interface{}{},
			"total":   0,
			"limit":   limit,
			"offset":  offset,
			"message": "No API calls found. Start making requests to see your call history here.",
		})
		return
	}

	// Format recent calls with model, tokens, status, timestamp
	calls := make([]gin.H, 0, len(records))
	for _, record := range records {
		call := gin.H{
			"id":                record.ID,
			"model":             record.Model,
			"prompt_tokens":     record.PromptTokens,
			"completion_tokens": record.CompletionTokens,
			"total_tokens":      record.TotalTokens,
			"status":            record.StatusCode,
			"timestamp":         record.RequestTime.Format(time.RFC3339),
			"duration_ms":       record.DurationMs,
		}

		// Include error message if present
		if record.ErrorMessage != "" {
			call["error"] = record.ErrorMessage
		}

		// Include token name if present
		if record.TokenName != "" {
			call["token_name"] = record.TokenName
		}

		calls = append(calls, call)
	}

	// Sort by timestamp descending (already done in query)
	response := gin.H{
		"calls":  calls,
		"total":  len(calls),
		"limit":  limit,
		"offset": offset,
	}

	c.JSON(http.StatusOK, response)
}

// Helper function to format model breakdown
func formatModelBreakdown(byModel map[string]database.ModelStats) []gin.H {
	breakdown := make([]gin.H, 0, len(byModel))
	for model, stats := range byModel {
		breakdown = append(breakdown, gin.H{
			"model":             model,
			"request_count":     stats.RequestCount,
			"total_tokens":      stats.TotalTokens,
			"prompt_tokens":     stats.PromptTokens,
			"completion_tokens": stats.CompletionTokens,
		})
	}
	return breakdown
}

// Helper function to format recent calls
func formatRecentCalls(recentCalls []database.UsageRecord) []gin.H {
	calls := make([]gin.H, 0, len(recentCalls))
	for _, record := range recentCalls {
		call := gin.H{
			"id":                record.ID,
			"model":             record.Model,
			"prompt_tokens":     record.PromptTokens,
			"completion_tokens": record.CompletionTokens,
			"total_tokens":      record.TotalTokens,
			"status":            record.StatusCode,
			"timestamp":         record.RequestTime.Format(time.RFC3339),
			"duration_ms":       record.DurationMs,
		}

		if record.ErrorMessage != "" {
			call["error"] = record.ErrorMessage
		}

		if record.TokenName != "" {
			call["token_name"] = record.TokenName
		}

		calls = append(calls, call)
	}
	return calls
}

// GetUserUsageTrends retrieves usage trends over time for the authenticated user
func GetUserUsageTrends(c *gin.Context) {
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

	// Parse days parameter (default 7, max 90)
	days := 7
	if daysStr := c.Query("days"); daysStr != "" {
		parsedDays, err := strconv.Atoi(daysStr)
		if err == nil && parsedDays > 0 {
			days = parsedDays
			if days > 90 {
				days = 90
			}
		}
	}

	// Get daily usage trends from database for this user
	trends, err := database.GetDailyUsageTrends(&userID, days)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get user usage trends")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve usage trends",
			"internal_error",
			"database_error",
		))
		return
	}

	// Format trends for chart display
	formattedTrends := make([]gin.H, 0, len(trends))
	for _, trend := range trends {
		formattedTrends = append(formattedTrends, gin.H{
			"date":              trend.Date.Format("2006-01-02"),
			"total_tokens":      trend.TotalTokens,
			"prompt_tokens":     trend.PromptTokens,
			"completion_tokens": trend.CompletionTokens,
			"request_count":     trend.Requests,
		})
	}

	response := gin.H{
		"days":   days,
		"trends": formattedTrends,
	}

	c.JSON(http.StatusOK, response)
}

// GetAdminUsageStats retrieves system-wide usage statistics for administrators
func GetAdminUsageStats(c *gin.Context) {
	// Parse query parameters for filtering
	filter := database.UsageFilter{}

	// Parse start_date
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Invalid start_date format. Expected YYYY-MM-DD",
				"invalid_request_error",
				"invalid_date_format",
			))
			return
		}
		filter.StartDate = &startDate
	}

	// Parse end_date
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Invalid end_date format. Expected YYYY-MM-DD",
				"invalid_request_error",
				"invalid_date_format",
			))
			return
		}
		// Set to end of day
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		filter.EndDate = &endDate
	}

	// Parse model filter
	if model := c.Query("model"); model != "" {
		filter.Model = &model
	}

	// Get aggregate statistics from database
	stats, err := database.GetAllUsageStats(filter)
	if err != nil {
		logrus.WithError(err).Error("Failed to get admin usage stats")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve usage statistics",
			"internal_error",
			"database_error",
		))
		return
	}

	// Format response
	response := gin.H{
		"total_users":    stats.TotalUsers,
		"total_requests": stats.TotalRequests,
		"total_tokens":   stats.TotalTokens,
		"top_users":      formatTopUsers(stats.TopUsers),
		"top_models":     formatTopModels(stats.TopModels),
	}

	c.JSON(http.StatusOK, response)
}

// Helper function to format top users
func formatTopUsers(topUsers []database.UserUsageSummary) []gin.H {
	users := make([]gin.H, 0, len(topUsers))
	for _, user := range topUsers {
		users = append(users, gin.H{
			"user_id":     user.UserID,
			"username":    user.Username,
			"requests":    user.Requests,
			"total_tokens": user.TotalTokens,
		})
	}
	return users
}

// Helper function to format top models
func formatTopModels(topModels []database.ModelStats) []gin.H {
	models := make([]gin.H, 0, len(topModels))
	for _, model := range topModels {
		models = append(models, gin.H{
			"model":             model.Model,
			"request_count":     model.RequestCount,
			"total_tokens":      model.TotalTokens,
			"prompt_tokens":     model.PromptTokens,
			"completion_tokens": model.CompletionTokens,
		})
	}
	return models
}

// GetAdminUsageTrends retrieves usage trends over time for administrators
func GetAdminUsageTrends(c *gin.Context) {
	// Parse days parameter (default 30, max 365)
	days := 30
	if daysStr := c.Query("days"); daysStr != "" {
		parsedDays, err := strconv.Atoi(daysStr)
		if err == nil && parsedDays > 0 {
			days = parsedDays
			if days > 365 {
				days = 365
			}
		}
	}

	// Parse view parameter (daily, weekly, monthly)
	view := c.Query("view")
	if view == "" {
		view = "daily"
	}

	// Validate view parameter
	if view != "daily" && view != "weekly" && view != "monthly" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid view parameter. Expected 'daily', 'weekly', or 'monthly'",
			"invalid_request_error",
			"invalid_view",
		))
		return
	}

	// Parse user_id filter (optional)
	var userID *int64
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		parsedUserID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err == nil {
			userID = &parsedUserID
		}
	}

	// Get daily usage trends from database
	trends, err := database.GetDailyUsageTrends(userID, days)
	if err != nil {
		logrus.WithError(err).Error("Failed to get usage trends")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve usage trends",
			"internal_error",
			"database_error",
		))
		return
	}

	// Format trends based on view
	var formattedTrends []gin.H
	var growthRate float64

	switch view {
	case "daily":
		formattedTrends = formatDailyTrends(trends)
		growthRate = calculateGrowthRate(trends)
	case "weekly":
		formattedTrends = aggregateWeeklyTrends(trends)
		growthRate = calculateGrowthRate(trends)
	case "monthly":
		formattedTrends = aggregateMonthlyTrends(trends)
		growthRate = calculateGrowthRate(trends)
	}

	// Format response for chart display
	response := gin.H{
		"view":        view,
		"days":        days,
		"trends":      formattedTrends,
		"growth_rate": growthRate,
	}

	c.JSON(http.StatusOK, response)
}

// Helper function to format daily trends
func formatDailyTrends(trends []database.DailyStats) []gin.H {
	formatted := make([]gin.H, 0, len(trends))
	for _, trend := range trends {
		formatted = append(formatted, gin.H{
			"date":         trend.Date.Format("2006-01-02"),
			"requests":     trend.Requests,
			"total_tokens": trend.TotalTokens,
		})
	}
	return formatted
}

// Helper function to aggregate weekly trends
func aggregateWeeklyTrends(trends []database.DailyStats) []gin.H {
	if len(trends) == 0 {
		return []gin.H{}
	}

	weeklyMap := make(map[string]*struct {
		StartDate   time.Time
		Requests    int
		TotalTokens int64
	})

	for _, trend := range trends {
		// Get the start of the week (Monday)
		year, week := trend.Date.ISOWeek()
		weekKey := fmt.Sprintf("%d-W%02d", year, week)

		if _, exists := weeklyMap[weekKey]; !exists {
			// Calculate Monday of this week
			weekday := int(trend.Date.Weekday())
			if weekday == 0 {
				weekday = 7 // Sunday
			}
			monday := trend.Date.AddDate(0, 0, -(weekday - 1))

			weeklyMap[weekKey] = &struct {
				StartDate   time.Time
				Requests    int
				TotalTokens int64
			}{
				StartDate: monday,
			}
		}

		weeklyMap[weekKey].Requests += trend.Requests
		weeklyMap[weekKey].TotalTokens += trend.TotalTokens
	}

	// Convert map to slice and sort by date
	formatted := make([]gin.H, 0, len(weeklyMap))
	for _, week := range weeklyMap {
		formatted = append(formatted, gin.H{
			"date":         week.StartDate.Format("2006-01-02"),
			"requests":     week.Requests,
			"total_tokens": week.TotalTokens,
		})
	}

	return formatted
}

// Helper function to aggregate monthly trends
func aggregateMonthlyTrends(trends []database.DailyStats) []gin.H {
	if len(trends) == 0 {
		return []gin.H{}
	}

	monthlyMap := make(map[string]*struct {
		StartDate   time.Time
		Requests    int
		TotalTokens int64
	})

	for _, trend := range trends {
		monthKey := trend.Date.Format("2006-01")

		if _, exists := monthlyMap[monthKey]; !exists {
			// Get first day of month
			firstDay := time.Date(trend.Date.Year(), trend.Date.Month(), 1, 0, 0, 0, 0, trend.Date.Location())
			monthlyMap[monthKey] = &struct {
				StartDate   time.Time
				Requests    int
				TotalTokens int64
			}{
				StartDate: firstDay,
			}
		}

		monthlyMap[monthKey].Requests += trend.Requests
		monthlyMap[monthKey].TotalTokens += trend.TotalTokens
	}

	// Convert map to slice
	formatted := make([]gin.H, 0, len(monthlyMap))
	for _, month := range monthlyMap {
		formatted = append(formatted, gin.H{
			"date":         month.StartDate.Format("2006-01-02"),
			"requests":     month.Requests,
			"total_tokens": month.TotalTokens,
		})
	}

	return formatted
}

// Helper function to calculate growth rate
func calculateGrowthRate(trends []database.DailyStats) float64 {
	if len(trends) < 2 {
		return 0.0
	}

	// Calculate average of first half vs second half
	midpoint := len(trends) / 2
	var firstHalfTotal, secondHalfTotal int64

	for i := 0; i < midpoint; i++ {
		firstHalfTotal += trends[i].TotalTokens
	}
	for i := midpoint; i < len(trends); i++ {
		secondHalfTotal += trends[i].TotalTokens
	}

	if firstHalfTotal == 0 {
		if secondHalfTotal > 0 {
			return 100.0
		}
		return 0.0
	}

	growthRate := ((float64(secondHalfTotal) - float64(firstHalfTotal)) / float64(firstHalfTotal)) * 100.0
	return growthRate
}

// GetAdminCursorSessionUsage retrieves usage statistics grouped by Cursor session
func GetAdminCursorSessionUsage(c *gin.Context) {
	// Parse query parameters for filtering
	filter := database.UsageFilter{}

	// Parse start_date
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Invalid start_date format. Expected YYYY-MM-DD",
				"invalid_request_error",
				"invalid_date_format",
			))
			return
		}
		filter.StartDate = &startDate
	}

	// Parse end_date
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Invalid end_date format. Expected YYYY-MM-DD",
				"invalid_request_error",
				"invalid_date_format",
			))
			return
		}
		// Set to end of day
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		filter.EndDate = &endDate
	}

	// Get Cursor session usage from database
	sessions, err := database.GetCursorSessionUsage(filter)
	if err != nil {
		logrus.WithError(err).Error("Failed to get cursor session usage")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve cursor session usage",
			"internal_error",
			"database_error",
		))
		return
	}

	// Format response
	formattedSessions := make([]gin.H, 0, len(sessions))
	for _, session := range sessions {
		formattedSessions = append(formattedSessions, gin.H{
			"cursor_session": session.CursorSession,
			"requests":       session.Requests,
			"total_tokens":   session.TotalTokens,
		})
	}

	response := gin.H{
		"sessions": formattedSessions,
		"total":    len(formattedSessions),
	}

	c.JSON(http.StatusOK, response)
}

// ExportUsageData exports usage data as CSV for administrators
func ExportUsageData(c *gin.Context) {
	// Parse date range from query parameters
	filter := database.UsageFilter{}

	// Parse start_date (required for export)
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Invalid start_date format. Expected YYYY-MM-DD",
				"invalid_request_error",
				"invalid_date_format",
			))
			return
		}
		filter.StartDate = &startDate
	}

	// Parse end_date (required for export)
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Invalid end_date format. Expected YYYY-MM-DD",
				"invalid_request_error",
				"invalid_date_format",
			))
			return
		}
		// Set to end of day
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		filter.EndDate = &endDate
	}

	// Parse optional user_id filter
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err == nil {
			filter.UserID = &userID
		}
	}

	// Parse optional model filter
	if model := c.Query("model"); model != "" {
		filter.Model = &model
	}

	// Set appropriate CSV headers
	filename := fmt.Sprintf("usage_export_%s.csv", time.Now().Format("2006-01-02_15-04-05"))
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Cache-Control", "no-cache")

	// Stream CSV data directly to response
	if err := database.StreamUsageRecordsCSV(c.Writer, filter); err != nil {
		logrus.WithError(err).Error("Failed to export usage data")
		// Note: We can't send JSON error after starting CSV stream
		// The error will be logged and the stream will be incomplete
		return
	}
}

// UpdateRetentionRequest represents the request body for updating retention period
type UpdateRetentionRequest struct {
	RetentionDays int `json:"retention_days" binding:"required"`
}

// GetRetentionConfig retrieves the current retention configuration
func GetRetentionConfig(c *gin.Context) {
	cleanupService := services.GetUsageCleanupService()
	config := cleanupService.GetConfig()

	response := gin.H{
		"enabled":         config.Enabled,
		"retention_days":  config.RetentionDays,
		"schedule_hour":   config.ScheduleHour,
		"schedule_minute": config.ScheduleMinute,
		"last_cleanup":    cleanupService.GetLastCleanup().Format(time.RFC3339),
		"is_running":      cleanupService.IsRunning(),
	}

	// Include last error if any
	if lastErr := cleanupService.GetLastError(); lastErr != nil {
		response["last_error"] = lastErr.Error()
	}

	c.JSON(http.StatusOK, response)
}

// UpdateRetentionConfig updates the retention period configuration
func UpdateRetentionConfig(c *gin.Context) {
	var req UpdateRetentionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid request format",
			"validation_error",
			"invalid_request",
		))
		return
	}

	// Validate minimum retention period (7 days)
	if req.RetentionDays < 7 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Retention period must be at least 7 days",
			"validation_error",
			"retention_too_short",
		))
		return
	}

	// Validate maximum retention period (365 days)
	if req.RetentionDays > 365 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Retention period cannot exceed 365 days",
			"validation_error",
			"retention_too_long",
		))
		return
	}

	// Update the retention period
	cleanupService := services.GetUsageCleanupService()
	if err := cleanupService.UpdateRetentionDays(req.RetentionDays); err != nil {
		logrus.WithError(err).Error("Failed to update retention period")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to update retention period",
			"internal_error",
			"update_failed",
		))
		return
	}

	logrus.Infof("Retention period updated to %d days by admin", req.RetentionDays)

	c.JSON(http.StatusOK, gin.H{
		"message":        "Retention period updated successfully",
		"retention_days": req.RetentionDays,
	})
}

// TriggerCleanupNow triggers an immediate cleanup operation
func TriggerCleanupNow(c *gin.Context) {
	cleanupService := services.GetUsageCleanupService()

	// Check if cleanup is enabled
	if !cleanupService.GetConfig().Enabled {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Cleanup service is disabled",
			"service_error",
			"cleanup_disabled",
		))
		return
	}

	// Run cleanup immediately
	deletedCount, err := cleanupService.RunCleanupNow()
	if err != nil {
		logrus.WithError(err).Error("Manual cleanup failed")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			fmt.Sprintf("Cleanup failed: %v", err),
			"internal_error",
			"cleanup_failed",
		))
		return
	}

	logrus.Infof("Manual cleanup completed: deleted %d records", deletedCount)

	c.JSON(http.StatusOK, gin.H{
		"message":       "Cleanup completed successfully",
		"deleted_count": deletedCount,
	})
}

// GetCleanupStats retrieves statistics about records eligible for cleanup
func GetCleanupStats(c *gin.Context) {
	cleanupService := services.GetUsageCleanupService()
	config := cleanupService.GetConfig()

	// Calculate cutoff date
	cutoffDate := time.Now().AddDate(0, 0, -config.RetentionDays)

	// Count records older than retention period
	count, err := database.CountUsageRecordsOlderThan(cutoffDate)
	if err != nil {
		logrus.WithError(err).Error("Failed to count old records")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve cleanup statistics",
			"internal_error",
			"database_error",
		))
		return
	}

	response := gin.H{
		"retention_days":       config.RetentionDays,
		"cutoff_date":          cutoffDate.Format("2006-01-02"),
		"records_to_delete":    count,
		"last_cleanup":         cleanupService.GetLastCleanup().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}
