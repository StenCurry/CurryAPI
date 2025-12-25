package handlers

import (
	"Curry2API-go/database"
	"Curry2API-go/models"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Request/Response types for game coin handlers

// DeductGameCoinsRequest represents the request body for deducting game coins
type DeductGameCoinsRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	GameType    string  `json:"game_type" binding:"required"`
	Description string  `json:"description"`
}

// AddGameCoinsRequest represents the request body for adding game coins
type AddGameCoinsRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	GameType    string  `json:"game_type" binding:"required"`
	Description string  `json:"description"`
}

// MigrateLocalStorageRequest represents the request body for migrating localStorage data
type MigrateLocalStorageRequest struct {
	Balance     float64 `json:"balance" binding:"required,gte=0"`
	TotalWon    float64 `json:"total_won" binding:"gte=0"`
	TotalLost   float64 `json:"total_lost" binding:"gte=0"`
	GamesPlayed int     `json:"games_played" binding:"gte=0"`
}

// GetGameBalanceHandler retrieves the current user's game coin balance
// GET /api/game/balance
// Requirements: 1.3, 7.3
func GetGameBalanceHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return // Error response already sent
	}

	// Get or create user game balance
	balance, err := database.GetOrCreateUserGameBalance(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get game balance")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve game balance",
			"internal_error",
			"database_error",
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance":         balance.Balance,
		"total_won":       balance.TotalWon,
		"total_lost":      balance.TotalLost,
		"total_exchanged": balance.TotalExchanged,
		"games_played":    balance.GamesPlayed,
		"created_at":      balance.CreatedAt,
		"updated_at":      balance.UpdatedAt,
	})
}

// DeductGameCoinsHandler deducts game coins from user's balance (for betting)
// POST /api/game/deduct
// Requirements: 1.2, 7.1
func DeductGameCoinsHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return // Error response already sent
	}

	var req DeductGameCoinsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid request format: "+err.Error(),
			"validation_error",
			"invalid_request",
		))
		return
	}

	// Validate game type
	if !isValidGameType(req.GameType) {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid game type",
			"validation_error",
			"invalid_game_type",
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

	// Deduct game coins
	transaction, err := database.DeductGameCoins(userID, req.Amount, req.GameType, req.Description)
	if err != nil {
		if err == database.ErrInsufficientGameCoins {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Insufficient game coin balance",
				"validation_error",
				"insufficient_balance",
			))
			return
		}
		if err == database.ErrInvalidAmount {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Invalid amount",
				"validation_error",
				"invalid_amount",
			))
			return
		}
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to deduct game coins")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to deduct game coins",
			"internal_error",
			"database_error",
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"transaction":   transaction,
		"balance_after": transaction.BalanceAfter,
	})
}

// AddGameCoinsHandler adds game coins to user's balance (for winning)
// POST /api/game/add
// Requirements: 1.2, 7.2
func AddGameCoinsHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return // Error response already sent
	}

	var req AddGameCoinsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid request format: "+err.Error(),
			"validation_error",
			"invalid_request",
		))
		return
	}

	// Validate game type
	if !isValidGameType(req.GameType) {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid game type",
			"validation_error",
			"invalid_game_type",
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

	// Add game coins
	transaction, err := database.AddGameCoins(userID, req.Amount, req.GameType, req.Description)
	if err != nil {
		if err == database.ErrInvalidAmount {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Invalid amount",
				"validation_error",
				"invalid_amount",
			))
			return
		}
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to add game coins")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to add game coins",
			"internal_error",
			"database_error",
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"transaction":   transaction,
		"balance_after": transaction.BalanceAfter,
	})
}

// ResetGameCoinsHandler resets user's game coin balance to initial value
// POST /api/game/reset
// Requirements: 8.2
func ResetGameCoinsHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return // Error response already sent
	}

	// Ensure user has a game balance record first
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

	// Reset game coins
	balance, err := database.ResetGameCoins(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to reset game coins")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to reset game coins",
			"internal_error",
			"database_error",
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"balance":         balance.Balance,
		"total_won":       balance.TotalWon,
		"total_lost":      balance.TotalLost,
		"total_exchanged": balance.TotalExchanged,
		"games_played":    balance.GamesPlayed,
	})
}

// GetGameTransactionsHandler retrieves paginated game coin transaction history
// GET /api/game/transactions
// Query params: limit (default 20, max 100), offset (default 0)
// Requirements: 1.6
func GetGameTransactionsHandler(c *gin.Context) {
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

	// Get transactions from database
	transactions, total, err := database.GetGameCoinTransactions(userID, limit, offset)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get game transactions")
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
			"created_at":    tx.CreatedAt,
		}

		if tx.GameType != "" {
			txData["game_type"] = tx.GameType
		}
		if tx.Description != "" {
			txData["description"] = tx.Description
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

// MigrateLocalStorageHandler migrates game coin data from localStorage to database
// POST /api/game/migrate
// Requirements: 1.5
func MigrateLocalStorageHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return // Error response already sent
	}

	var req MigrateLocalStorageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid request format: "+err.Error(),
			"validation_error",
			"invalid_request",
		))
		return
	}

	// Migrate localStorage data
	balance, err := database.MigrateLocalStorageData(userID, req.Balance, req.TotalWon, req.TotalLost, req.GamesPlayed)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to migrate localStorage data")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to migrate game data",
			"internal_error",
			"database_error",
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"balance":         balance.Balance,
		"total_won":       balance.TotalWon,
		"total_lost":      balance.TotalLost,
		"total_exchanged": balance.TotalExchanged,
		"games_played":    balance.GamesPlayed,
		"migrated":        true,
	})
}

// Helper functions

// getUserIDFromContext extracts and validates user_id from gin context
func getUserIDFromContext(c *gin.Context) (int64, error) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			"User not authenticated",
			"authentication_error",
			"missing_user_id",
		))
		return 0, database.ErrGameBalanceNotFound
	}

	userID, ok := userIDInterface.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Invalid user ID format",
			"internal_error",
			"invalid_user_id_type",
		))
		return 0, database.ErrGameBalanceNotFound
	}

	return userID, nil
}

// isValidGameType checks if the game type is valid
func isValidGameType(gameType string) bool {
	switch gameType {
	case database.GameTypeWheel, database.GameTypeCoin, database.GameTypeNumber:
		return true
	default:
		return false
	}
}


// CreateGameRecordRequest represents the request body for creating a game record
type CreateGameRecordRequest struct {
	GameType  string          `json:"game_type" binding:"required"`
	BetAmount float64         `json:"bet_amount" binding:"required,gt=0"`
	Result    string          `json:"result" binding:"required"`
	Payout    float64         `json:"payout" binding:"gte=0"`
	Details   json.RawMessage `json:"details"`
}

// CreateGameRecordHandler creates a new game record
// POST /api/game/record
// Requirements: 1.1, 7.1
func CreateGameRecordHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return // Error response already sent
	}

	var req CreateGameRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid request format: "+err.Error(),
			"validation_error",
			"invalid_request",
		))
		return
	}

	// Validate game_type is one of: wheel, coin, number
	if !isValidGameType(req.GameType) {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid game type. Must be one of: wheel, coin, number",
			"validation_error",
			"invalid_game_type",
		))
		return
	}

	// Validate result is one of: win, lose
	if req.Result != database.GameResultWin && req.Result != database.GameResultLose {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid result. Must be one of: win, lose",
			"validation_error",
			"invalid_result",
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

	// Create game record
	record, err := database.CreateGameRecord(userID, req.GameType, req.BetAmount, req.Result, req.Payout, req.Details)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to create game record")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to create game record",
			"internal_error",
			"database_error",
		))
		return
	}

	// Get updated stats
	stats, err := database.GetGameStats(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get game stats")
		// Still return the record even if stats fail
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"record":  record,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"record":  record,
		"stats":   stats,
	})
}


// GetGameRecordsHandler retrieves paginated game records for the current user
// GET /api/game/records
// Query params: limit (default 10, max 100), offset (default 0)
// Requirements: 1.5, 1.6, 7.2
func GetGameRecordsHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return // Error response already sent
	}

	// Parse pagination parameters
	limit := 10
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

	// Get game records from database
	records, total, err := database.GetGameRecords(userID, limit, offset)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get game records")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve game records",
			"internal_error",
			"database_error",
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"records": records,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}


// GetGameStatsHandler retrieves game statistics for the current user
// GET /api/game/stats
// Requirements: 2.1
func GetGameStatsHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return // Error response already sent
	}

	// Get game stats from database
	stats, err := database.GetGameStats(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get game stats")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve game statistics",
			"internal_error",
			"database_error",
		))
		return
	}

	c.JSON(http.StatusOK, stats)
}


// GetLeaderboardHandler retrieves the global leaderboard
// GET /api/game/leaderboard
// Query params: sort (winnings/games, default winnings), limit (default 10)
// Requirements: 3.1, 3.2, 3.3
func GetLeaderboardHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return // Error response already sent
	}

	// Parse query parameters
	sortBy := c.DefaultQuery("sort", "winnings")
	if sortBy != "winnings" && sortBy != "games" {
		sortBy = "winnings"
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Get leaderboard from database
	entries, currentUser, totalPlayers, err := database.GetLeaderboard(userID, sortBy, limit)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get leaderboard")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to retrieve leaderboard",
			"internal_error",
			"database_error",
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entries":       entries,
		"current_user":  currentUser,
		"total_players": totalPlayers,
	})
}
