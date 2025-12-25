package handlers

import (
	"Curry2API-go/database"
	"Curry2API-go/models"
	"Curry2API-go/services"
	"Curry2API-go/utils"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// trackUsageFromContext extracts context info and tracks usage
// This function handles both successful and failed requests
// It is safe to call even if tracking fails - errors are logged but don't affect the response
func trackUsageFromContext(c *gin.Context, usage *models.Usage, statusCode int, errorMsg string) {
	// Extract request start time
	requestStartTime, exists := c.Get("request_start_time")
	if !exists {
		logrus.Debug("request_start_time not found in context, skipping usage tracking")
		return
	}
	startTime, ok := requestStartTime.(time.Time)
	if !ok {
		logrus.Debug("invalid request_start_time type in context")
		return
	}
	
	// Extract model
	requestModel, exists := c.Get("request_model")
	if !exists {
		logrus.Debug("request_model not found in context")
		return
	}
	model, ok := requestModel.(string)
	if !ok {
		logrus.Debug("invalid request_model type in context")
		return
	}
	
	// Extract usage info
	usageInfoRaw, exists := c.Get("usage_info")
	if !exists {
		logrus.Debug("usage_info not found in context, skipping usage tracking")
		return
	}
	usageInfo, ok := usageInfoRaw.(*utils.UsageContextInfo)
	if !ok {
		logrus.Debug("invalid usage_info type in context")
		return
	}
	
	// Calculate response time and duration
	responseTime := time.Now()
	duration := responseTime.Sub(startTime)
	
	// Prepare usage record
	var promptTokens, completionTokens, totalTokens int
	if usage != nil {
		promptTokens = usage.PromptTokens
		completionTokens = usage.CompletionTokens
		totalTokens = usage.TotalTokens
	}
	
	// Get cursor session if available
	cursorSession := ""
	if sessionRaw, exists := c.Get("cursor_session"); exists {
		if session, ok := sessionRaw.(string); ok {
			cursorSession = session
			logrus.WithField("cursor_session", cursorSession).Debug("Got cursor_session from context")
		}
	} else {
		logrus.Debug("cursor_session not found in context")
	}
	
	// Track usage with the usage tracker service
	tracker := services.GetUsageTracker()
	record := &services.UsageRecord{
		UserID:           usageInfo.UserID,
		Username:         usageInfo.Username,
		APIToken:         usageInfo.APIToken,
		TokenName:        usageInfo.TokenName,
		Model:            model,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
		CursorSession:    cursorSession,
		StatusCode:       statusCode,
		ErrorMessage:     errorMsg,
		RequestTime:      startTime,
		ResponseTime:     responseTime,
		Duration:         duration,
	}
	
	if err := tracker.TrackUsage(record); err != nil {
		logrus.WithError(err).Warn("Failed to track usage")
	}
	
	// Update Cursor Session usage count and token quota asynchronously
	if cursorSession != "" && cursorSession != "x-is-human-fallback" {
		go func() {
			success := statusCode >= 200 && statusCode < 300
			logrus.WithFields(logrus.Fields{
				"cursor_session": cursorSession,
				"status_code":    statusCode,
				"success":        success,
				"total_tokens":   totalTokens,
			}).Info("Updating cursor session usage")
			
			// Update usage count (success/fail tracking)
			if err := database.UpdateCursorSessionUsage(cursorSession, success); err != nil {
				logrus.WithError(err).WithField("cursor_session", cursorSession).Warn("Failed to update cursor session usage count")
			}
			
			// Update daily token usage for successful requests
			if success && totalTokens > 0 {
				if err := database.UpdateSessionQuotaUsage(cursorSession, int64(totalTokens)); err != nil {
					logrus.WithError(err).WithFields(logrus.Fields{
						"cursor_session": cursorSession,
						"tokens":         totalTokens,
					}).Warn("Failed to update cursor session daily_token_used")
				} else {
					logrus.WithFields(logrus.Fields{
						"cursor_session": cursorSession,
						"tokens":         totalTokens,
					}).Debug("Cursor session daily_token_used updated")
				}
			}
		}()
	}

	// Update API key last_used_at timestamp asynchronously
	// Only update for successful requests to avoid updating on errors
	if statusCode >= 200 && statusCode < 300 {
		go func() {
			if err := database.UpdateAPIKeyLastUsed(usageInfo.APIToken, responseTime); err != nil {
				logrus.WithError(err).Debug("Failed to update API key last_used_at")
			}
		}()
	}

	// Deduct balance for successful API calls with token usage
	// Requirements: 2.2, 11.1, 11.2
	if statusCode >= 200 && statusCode < 300 && totalTokens > 0 {
		go deductBalanceForUsage(usageInfo.UserID, totalTokens, usageInfo.APIToken, model)
	}
}

// deductBalanceForUsage deducts balance based on token usage
// This function runs asynchronously to avoid blocking the response
// Requirements: 2.2 - Deduct cost from user's balance after API call
// Requirements: 12.2 - Update token quota_used after API call
func deductBalanceForUsage(userID int64, tokens int, apiToken, model string) {
	// Calculate cost: $1 = 1,000,000 tokens
	cost := database.CalculateCost(tokens)

	logrus.WithFields(logrus.Fields{
		"user_id":   userID,
		"tokens":    tokens,
		"cost":      cost,
		"api_token": apiToken,
		"model":     model,
	}).Debug("Deducting balance for API usage")

	// Update token quota_used
	// Requirements: 12.2 - Track token's consumed amount separately
	if err := database.UpdateTokenQuotaUsed(apiToken, cost); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"api_token": apiToken,
			"cost":      cost,
		}).Warn("Failed to update token quota_used")
	} else {
		logrus.WithFields(logrus.Fields{
			"api_token": apiToken,
			"cost":      cost,
		}).Debug("Token quota_used updated")
	}

	// Deduct balance and create transaction record
	transaction, err := database.DeductBalance(userID, tokens, apiToken, model)
	if err != nil {
		// Log error but don't fail - balance deduction failure shouldn't affect API response
		if errors.Is(err, database.ErrBalanceNotFound) {
			// User doesn't have a balance record yet - this is expected for users
			// created before the balance system was implemented
			logrus.WithFields(logrus.Fields{
				"user_id": userID,
			}).Debug("User has no balance record, skipping balance deduction")
			return
		}
		logrus.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"tokens":  tokens,
			"cost":    cost,
		}).Warn("Failed to deduct balance for API usage")
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id":       userID,
		"tokens":        tokens,
		"cost":          cost,
		"balance_after": transaction.BalanceAfter,
		"transaction_id": transaction.ID,
	}).Info("Balance deducted for API usage")
}
