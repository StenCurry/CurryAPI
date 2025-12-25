package handlers

import (
	"Curry2API-go/database"
	"Curry2API-go/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetBalanceHandler retrieves the current user's balance
// GET /api/balance
// Requirements: 6.1
func GetBalanceHandler(c *gin.Context) {
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

	// Get user balance from database
	balance, err := database.GetUserBalance(userID)
	if err != nil {
		if err == database.ErrBalanceNotFound {
			// Auto-create balance record for existing users who don't have one
			logrus.WithField("user_id", userID).Info("Creating balance record for existing user")
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
			logrus.WithError(err).WithField("user_id", userID).Error("Failed to get user balance")
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				"Failed to retrieve balance",
				"internal_error",
				"database_error",
			))
			return
		}
	}

	// Return balance information
	c.JSON(http.StatusOK, gin.H{
		"balance":         balance.Balance,
		"status":          balance.Status,
		"referral_code":   balance.ReferralCode,
		"total_consumed":  balance.TotalConsumed,
		"total_recharged": balance.TotalRecharged,
		"created_at":      balance.CreatedAt,
		"updated_at":      balance.UpdatedAt,
	})
}


// GetTransactionsHandler retrieves paginated transaction history for the current user
// GET /api/balance/transactions
// Query params: limit (default 20, max 100), offset (default 0)
// Requirements: 6.2, 6.3
func GetTransactionsHandler(c *gin.Context) {
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

	// Get transactions from database
	transactions, total, err := database.GetBalanceTransactions(userID, limit, offset)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get balance transactions")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve transactions",
			"internal_error",
			"database_error",
		))
		return
	}

	// Format transactions for response
	formattedTransactions := make([]gin.H, 0, len(transactions))
	for _, tx := range transactions {
		txData := gin.H{
			"id":            tx.ID,
			"type":          tx.Type,
			"amount":        tx.Amount,
			"balance_after": tx.BalanceAfter,
			"tokens":        tx.Tokens,
			"description":   tx.Description,
			"created_at":    tx.CreatedAt,
		}

		// Include optional fields if present
		if tx.Model != "" {
			txData["model"] = tx.Model
		}
		if tx.APIToken != "" {
			// Mask the API token for security
			txData["api_token"] = maskAPIToken(tx.APIToken)
		}
		if tx.RelatedUserID != nil {
			txData["related_user_id"] = *tx.RelatedUserID
		}
		if tx.AdminID != nil {
			txData["admin_id"] = *tx.AdminID
		}

		formattedTransactions = append(formattedTransactions, txData)
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": formattedTransactions,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
	})
}

// maskAPIToken masks an API token for display (shows first 4 and last 4 characters)
func maskAPIToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}
