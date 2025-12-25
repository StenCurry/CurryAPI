package handlers

import (
	"Curry2API-go/database"
	"Curry2API-go/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ExchangeGameCoinsRequest represents the request body for exchanging game coins
type ExchangeGameCoinsRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

// ExchangeGameCoinsHandler exchanges game coins for account balance (USD)
// POST /api/game/exchange
// Requirements: 2.1, 2.2, 2.3, 2.6, 2.7
func ExchangeGameCoinsHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return // Error response already sent
	}

	var req ExchangeGameCoinsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid request format: "+err.Error(),
			"validation_error",
			"invalid_request",
		))
		return
	}

	// Validate amount is positive
	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid exchange amount",
			"validation_error",
			"invalid_amount",
		))
		return
	}

	// Validate minimum exchange amount
	if req.Amount < database.MinimumExchangeAmount {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Minimum exchange amount is 1 game coin",
			"validation_error",
			"below_minimum",
		))
		return
	}

	// Ensure user has a game balance record
	_, err = database.GetOrCreateUserGameBalance(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get/create game balance")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to access game balance",
			"internal_error",
			"database_error",
		))
		return
	}


	// Execute exchange
	exchangeRecord, err := database.ExchangeGameCoins(userID, req.Amount)
	if err != nil {
		switch err {
		case database.ErrInsufficientGameCoins:
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Insufficient game coin balance",
				"validation_error",
				"insufficient_balance",
			))
			return
		case database.ErrInvalidAmount:
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Invalid exchange amount",
				"validation_error",
				"invalid_amount",
			))
			return
		case database.ErrBelowMinimumExchange:
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Minimum exchange amount is 1 game coin",
				"validation_error",
				"below_minimum",
			))
			return
		case database.ErrDailyLimitExceeded:
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Daily exchange limit exceeded",
				"validation_error",
				"daily_limit_exceeded",
			))
			return
		case database.ErrGameBalanceNotFound:
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Game balance not found",
				"validation_error",
				"balance_not_found",
			))
			return
		case database.ErrBalanceNotFound:
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Account balance not found",
				"validation_error",
				"account_balance_not_found",
			))
			return
		default:
			logrus.WithError(err).WithField("user_id", userID).Error("Failed to exchange game coins")
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				"Failed to process exchange",
				"internal_error",
				"database_error",
			))
			return
		}
	}

	// Get updated balances
	gameBalance, _ := database.GetUserGameBalance(userID)
	accountBalance, _ := database.GetUserBalance(userID)

	var newGameBalance, newAccountBalance float64
	if gameBalance != nil {
		newGameBalance = gameBalance.Balance
	}
	if accountBalance != nil {
		newAccountBalance = accountBalance.Balance
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"exchange_record": gin.H{
			"id":                exchangeRecord.ID,
			"game_coins_amount": exchangeRecord.GameCoinsAmount,
			"usd_amount":        exchangeRecord.USDAmount,
			"exchange_rate":     exchangeRecord.ExchangeRate,
			"status":            exchangeRecord.Status,
			"created_at":        exchangeRecord.CreatedAt,
		},
		"new_game_balance":    newGameBalance,
		"new_account_balance": newAccountBalance,
	})
}

// GetExchangeHistoryHandler retrieves paginated exchange history for the current user
// GET /api/game/exchange/history
// Query params: limit (default 20, max 100), offset (default 0)
// Requirements: 3.1, 3.2, 3.3
func GetExchangeHistoryHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return // Error response already sent
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

	// Get exchange history from database
	records, total, err := database.GetExchangeHistory(userID, limit, offset)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get exchange history")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve exchange history",
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

// PurchaseGameCoinsRequest represents the request body for purchasing game coins with USD
type PurchaseGameCoinsRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

// PurchaseGameCoinsHandler exchanges account balance (USD) for game coins
// POST /api/game/purchase
func PurchaseGameCoinsHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return // Error response already sent
	}

	var req PurchaseGameCoinsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid request format: "+err.Error(),
			"validation_error",
			"invalid_request",
		))
		return
	}

	// Validate amount is positive
	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid purchase amount",
			"validation_error",
			"invalid_amount",
		))
		return
	}

	// Validate minimum purchase amount
	if req.Amount < database.MinimumExchangeAmount {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Minimum purchase amount is $1",
			"validation_error",
			"below_minimum",
		))
		return
	}

	// Execute purchase
	exchangeRecord, err := database.ExchangeUSDToGameCoins(userID, req.Amount)
	if err != nil {
		switch err {
		case database.ErrInsufficientBalance:
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Insufficient account balance",
				"validation_error",
				"insufficient_balance",
			))
			return
		case database.ErrInvalidAmount:
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Invalid purchase amount",
				"validation_error",
				"invalid_amount",
			))
			return
		case database.ErrBelowMinimumExchange:
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Minimum purchase amount is $1",
				"validation_error",
				"below_minimum",
			))
			return
		case database.ErrBalanceNotFound:
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Account balance not found",
				"validation_error",
				"account_balance_not_found",
			))
			return
		default:
			logrus.WithError(err).WithField("user_id", userID).Error("Failed to purchase game coins")
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				"Failed to process purchase",
				"internal_error",
				"database_error",
			))
			return
		}
	}

	// Get updated balances
	gameBalance, _ := database.GetUserGameBalance(userID)
	accountBalance, _ := database.GetUserBalance(userID)

	var newGameBalance, newAccountBalance float64
	if gameBalance != nil {
		newGameBalance = gameBalance.Balance
	}
	if accountBalance != nil {
		newAccountBalance = accountBalance.Balance
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"purchase_record": gin.H{
			"id":                exchangeRecord.ID,
			"game_coins_amount": exchangeRecord.GameCoinsAmount,
			"usd_amount":        exchangeRecord.USDAmount,
			"exchange_rate":     exchangeRecord.ExchangeRate,
			"status":            exchangeRecord.Status,
			"created_at":        exchangeRecord.CreatedAt,
		},
		"new_game_balance":    newGameBalance,
		"new_account_balance": newAccountBalance,
	})
}

// GetTodayExchangeAmountHandler retrieves today's exchange amount and remaining limit
// GET /api/game/exchange/today
// Requirements: 2.7
func GetTodayExchangeAmountHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return // Error response already sent
	}

	// Get today's exchange amount
	todayAmount, err := database.GetTodayExchangeAmount(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get today's exchange amount")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve today's exchange amount",
			"internal_error",
			"database_error",
		))
		return
	}

	remaining := database.DailyExchangeLimit - todayAmount
	if remaining < 0 {
		remaining = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"amount":    todayAmount,
		"limit":     database.DailyExchangeLimit,
		"remaining": remaining,
	})
}
