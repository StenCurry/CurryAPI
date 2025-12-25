package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"
)

// Constants for game coin system
const (
	InitialGameCoins     = 100.0  // Initial game coins for new users
	MinimumExchangeAmount = 1.0   // Minimum exchange amount
	DailyExchangeLimit   = 1000.0 // Daily exchange limit
	ExchangeRate         = 1.0    // 1 game coin = $1 USD
)

// Game coin transaction types
const (
	GameTxTypeInitial  = "initial"
	GameTxTypeBet      = "game_bet"
	GameTxTypeWin      = "game_win"
	GameTxTypeExchange = "exchange"
	GameTxTypeReset    = "reset"
	GameTxTypeMigrate  = "migrate"
)

// Game types
const (
	GameTypeWheel  = "wheel"
	GameTypeCoin   = "coin"
	GameTypeNumber = "number"
)

// Errors for game coin system
var (
	ErrGameBalanceNotFound     = errors.New("game balance record not found")
	ErrInsufficientGameCoins   = errors.New("insufficient game coins")
	ErrInvalidAmount           = errors.New("invalid amount")
	ErrBelowMinimumExchange    = errors.New("amount below minimum exchange")
	ErrDailyLimitExceeded      = errors.New("daily exchange limit exceeded")
)

// UserGameBalance represents a user's game coin balance record
type UserGameBalance struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	Balance        float64   `json:"balance"`
	TotalWon       float64   `json:"total_won"`
	TotalLost      float64   `json:"total_lost"`
	TotalExchanged float64   `json:"total_exchanged"`
	GamesPlayed    int       `json:"games_played"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}


// GameCoinTransaction represents a game coin transaction record
type GameCoinTransaction struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	Type         string    `json:"type"`
	GameType     string    `json:"game_type,omitempty"`
	Amount       float64   `json:"amount"`
	BalanceAfter float64   `json:"balance_after"`
	Description  string    `json:"description,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// GameRecord represents a single game round record
type GameRecord struct {
	ID        int64           `json:"id"`
	UserID    int64           `json:"user_id"`
	GameType  string          `json:"game_type"`
	BetAmount float64         `json:"bet_amount"`
	Result    string          `json:"result"`
	Payout    float64         `json:"payout"`
	NetProfit float64         `json:"net_profit"`
	Details   json.RawMessage `json:"details"`
	CreatedAt time.Time       `json:"created_at"`
}

// GameStats represents aggregated game statistics for a user
type GameStats struct {
	GamesPlayed int     `json:"games_played"`
	Wins        int     `json:"wins"`
	Losses      int     `json:"losses"`
	WinRate     string  `json:"win_rate"`
	NetProfit   string  `json:"net_profit"`
	TotalWon    string  `json:"total_won"`
	TotalLost   string  `json:"total_lost"`
}

// LeaderboardEntry represents a single entry in the leaderboard
type LeaderboardEntry struct {
	Rank          int     `json:"rank"`
	UserID        int64   `json:"user_id"`
	Username      string  `json:"username"`
	TotalWinnings float64 `json:"total_winnings"`
	GamesPlayed   int     `json:"games_played"`
}

// Game result constants
const (
	GameResultWin  = "win"
	GameResultLose = "lose"
)

// roundToTwoDecimals rounds a float64 to 2 decimal places
func roundToTwoDecimals(val float64) float64 {
	return math.Round(val*100) / 100
}

// CreateUserGameBalance creates a new game balance record for a user with initial 100 game coins
// Requirements: 1.1
func CreateUserGameBalance(userID int64) (*UserGameBalance, error) {
	now := time.Now()

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Insert game balance record
	result, err := tx.Exec(
		`INSERT INTO user_game_balances (user_id, balance, total_won, total_lost, total_exchanged, games_played, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, InitialGameCoins, 0, 0, 0, 0, now, now,
	)
	if err != nil {
		return nil, err
	}

	balanceID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Create initial transaction record
	_, err = tx.Exec(
		`INSERT INTO game_coin_transactions (user_id, type, game_type, amount, balance_after, description, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, GameTxTypeInitial, nil, InitialGameCoins, InitialGameCoins, "Initial game coins", now,
	)
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &UserGameBalance{
		ID:             balanceID,
		UserID:         userID,
		Balance:        InitialGameCoins,
		TotalWon:       0,
		TotalLost:      0,
		TotalExchanged: 0,
		GamesPlayed:    0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}


// GetUserGameBalance retrieves a user's game coin balance record
// Requirements: 1.3
func GetUserGameBalance(userID int64) (*UserGameBalance, error) {
	balance := &UserGameBalance{}

	err := db.QueryRow(
		`SELECT id, user_id, balance, total_won, total_lost, total_exchanged, games_played, created_at, updated_at
		 FROM user_game_balances WHERE user_id = ?`,
		userID,
	).Scan(&balance.ID, &balance.UserID, &balance.Balance, &balance.TotalWon, &balance.TotalLost,
		&balance.TotalExchanged, &balance.GamesPlayed, &balance.CreatedAt, &balance.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrGameBalanceNotFound
	}
	if err != nil {
		return nil, err
	}

	return balance, nil
}

// GetOrCreateUserGameBalance retrieves or creates a user's game coin balance
// Requirements: 1.1, 1.3
func GetOrCreateUserGameBalance(userID int64) (*UserGameBalance, error) {
	balance, err := GetUserGameBalance(userID)
	if err == ErrGameBalanceNotFound {
		return CreateUserGameBalance(userID)
	}
	return balance, err
}

// DeductGameCoins deducts game coins from user's balance (for betting)
// Requirements: 1.2, 7.1
func DeductGameCoins(userID int64, amount float64, gameType, description string) (*GameCoinTransaction, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	amount = roundToTwoDecimals(amount)

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Get current balance with lock
	var currentBalance float64
	err = tx.QueryRow(
		`SELECT balance FROM user_game_balances WHERE user_id = ? FOR UPDATE`,
		userID,
	).Scan(&currentBalance)

	if err == sql.ErrNoRows {
		return nil, ErrGameBalanceNotFound
	}
	if err != nil {
		return nil, err
	}

	// Check sufficient balance
	if currentBalance < amount {
		return nil, ErrInsufficientGameCoins
	}

	// Calculate new balance
	newBalance := roundToTwoDecimals(currentBalance - amount)
	now := time.Now()

	// Update balance and stats
	_, err = tx.Exec(
		`UPDATE user_game_balances SET balance = ?, total_lost = total_lost + ?, games_played = games_played + 1, updated_at = ?
		 WHERE user_id = ?`,
		newBalance, amount, now, userID,
	)
	if err != nil {
		return nil, err
	}

	// Create transaction record (negative amount for deduction)
	result, err := tx.Exec(
		`INSERT INTO game_coin_transactions (user_id, type, game_type, amount, balance_after, description, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, GameTxTypeBet, gameType, -amount, newBalance, description, now,
	)
	if err != nil {
		return nil, err
	}

	txID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &GameCoinTransaction{
		ID:           txID,
		UserID:       userID,
		Type:         GameTxTypeBet,
		GameType:     gameType,
		Amount:       -amount,
		BalanceAfter: newBalance,
		Description:  description,
		CreatedAt:    now,
	}, nil
}


// AddGameCoins adds game coins to user's balance (for winning)
// Requirements: 1.2, 7.2
func AddGameCoins(userID int64, amount float64, gameType, description string) (*GameCoinTransaction, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	amount = roundToTwoDecimals(amount)

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Get current balance with lock
	var currentBalance float64
	err = tx.QueryRow(
		`SELECT balance FROM user_game_balances WHERE user_id = ? FOR UPDATE`,
		userID,
	).Scan(&currentBalance)

	if err == sql.ErrNoRows {
		return nil, ErrGameBalanceNotFound
	}
	if err != nil {
		return nil, err
	}

	// Calculate new balance
	newBalance := roundToTwoDecimals(currentBalance + amount)
	now := time.Now()

	// Update balance and stats
	_, err = tx.Exec(
		`UPDATE user_game_balances SET balance = ?, total_won = total_won + ?, updated_at = ?
		 WHERE user_id = ?`,
		newBalance, amount, now, userID,
	)
	if err != nil {
		return nil, err
	}

	// Create transaction record (positive amount for addition)
	result, err := tx.Exec(
		`INSERT INTO game_coin_transactions (user_id, type, game_type, amount, balance_after, description, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, GameTxTypeWin, gameType, amount, newBalance, description, now,
	)
	if err != nil {
		return nil, err
	}

	txID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &GameCoinTransaction{
		ID:           txID,
		UserID:       userID,
		Type:         GameTxTypeWin,
		GameType:     gameType,
		Amount:       amount,
		BalanceAfter: newBalance,
		Description:  description,
		CreatedAt:    now,
	}, nil
}

// ResetGameCoins resets user's game coin balance to initial value and clears history
// Requirements: 8.2, 8.3, 8.4
func ResetGameCoins(userID int64) (*UserGameBalance, error) {
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Check if user has game balance
	var balanceID int64
	err = tx.QueryRow(
		`SELECT id FROM user_game_balances WHERE user_id = ? FOR UPDATE`,
		userID,
	).Scan(&balanceID)

	if err == sql.ErrNoRows {
		return nil, ErrGameBalanceNotFound
	}
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// Delete all previous transaction records (clear history)
	_, err = tx.Exec(
		`DELETE FROM game_coin_transactions WHERE user_id = ?`,
		userID,
	)
	if err != nil {
		return nil, err
	}

	// Reset balance to initial value and clear stats
	_, err = tx.Exec(
		`UPDATE user_game_balances SET balance = ?, total_won = 0, total_lost = 0, games_played = 0, updated_at = ?
		 WHERE user_id = ?`,
		InitialGameCoins, now, userID,
	)
	if err != nil {
		return nil, err
	}

	// Create reset transaction record
	_, err = tx.Exec(
		`INSERT INTO game_coin_transactions (user_id, type, game_type, amount, balance_after, description, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, GameTxTypeReset, nil, InitialGameCoins, InitialGameCoins, "Game coins reset", now,
	)
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &UserGameBalance{
		ID:             balanceID,
		UserID:         userID,
		Balance:        InitialGameCoins,
		TotalWon:       0,
		TotalLost:      0,
		TotalExchanged: 0,
		GamesPlayed:    0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}


// GetGameCoinTransactions retrieves paginated game coin transaction history for a user
// Requirements: 1.6
func GetGameCoinTransactions(userID int64, limit, offset int) ([]*GameCoinTransaction, int, error) {
	// Get total count
	var total int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM game_coin_transactions WHERE user_id = ?`,
		userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get transactions
	rows, err := db.Query(
		`SELECT id, user_id, type, game_type, amount, balance_after, description, created_at
		 FROM game_coin_transactions WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []*GameCoinTransaction
	for rows.Next() {
		tx := &GameCoinTransaction{}
		var gameType, description sql.NullString

		err := rows.Scan(&tx.ID, &tx.UserID, &tx.Type, &gameType, &tx.Amount, &tx.BalanceAfter, &description, &tx.CreatedAt)
		if err != nil {
			return nil, 0, err
		}

		if gameType.Valid {
			tx.GameType = gameType.String
		}
		if description.Valid {
			tx.Description = description.String
		}

		transactions = append(transactions, tx)
	}

	return transactions, total, nil
}

// MigrateLocalStorageData migrates game coin data from localStorage to database
// Requirements: 1.5
func MigrateLocalStorageData(userID int64, balance, totalWon, totalLost float64, gamesPlayed int) (*UserGameBalance, error) {
	// Round values to 2 decimal places
	balance = roundToTwoDecimals(balance)
	totalWon = roundToTwoDecimals(totalWon)
	totalLost = roundToTwoDecimals(totalLost)

	now := time.Now()

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Check if user already has game balance
	var existingID int64
	err = tx.QueryRow(
		`SELECT id FROM user_game_balances WHERE user_id = ?`,
		userID,
	).Scan(&existingID)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// If balance already exists, don't migrate
	if err == nil {
		// Return existing balance
		tx.Rollback()
		return GetUserGameBalance(userID)
	}

	// Insert migrated balance record
	result, err := tx.Exec(
		`INSERT INTO user_game_balances (user_id, balance, total_won, total_lost, total_exchanged, games_played, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, balance, totalWon, totalLost, 0, gamesPlayed, now, now,
	)
	if err != nil {
		return nil, err
	}

	balanceID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Create migration transaction record
	_, err = tx.Exec(
		`INSERT INTO game_coin_transactions (user_id, type, game_type, amount, balance_after, description, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, GameTxTypeMigrate, nil, balance, balance, "Migrated from localStorage", now,
	)
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &UserGameBalance{
		ID:             balanceID,
		UserID:         userID,
		Balance:        balance,
		TotalWon:       totalWon,
		TotalLost:      totalLost,
		TotalExchanged: 0,
		GamesPlayed:    gamesPlayed,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// CreateGameRecord creates a new game record and updates user statistics atomically
// Requirements: 1.1, 2.3
func CreateGameRecord(userID int64, gameType string, betAmount float64, result string, payout float64, details json.RawMessage) (*GameRecord, error) {
	// Validate game type
	if gameType != GameTypeWheel && gameType != GameTypeCoin && gameType != GameTypeNumber {
		return nil, fmt.Errorf("invalid game type: %s", gameType)
	}

	// Validate result
	if result != GameResultWin && result != GameResultLose {
		return nil, fmt.Errorf("invalid result: %s", result)
	}

	// Round amounts
	betAmount = roundToTwoDecimals(betAmount)
	payout = roundToTwoDecimals(payout)
	netProfit := roundToTwoDecimals(payout - betAmount)

	now := time.Now()

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Insert game record
	recordResult, err := tx.Exec(
		`INSERT INTO game_records (user_id, game_type, bet_amount, result, payout, net_profit, details, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, gameType, betAmount, result, payout, netProfit, details, now,
	)
	if err != nil {
		return nil, err
	}

	recordID, err := recordResult.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Update user_game_balances: increment games_played, update total_won/total_lost
	// Note: We don't update wins column here as stats are calculated from game_records table
	var updateQuery string
	if result == GameResultWin {
		updateQuery = `UPDATE user_game_balances 
			SET games_played = games_played + 1, 
			    total_won = total_won + ?, 
			    updated_at = ? 
			WHERE user_id = ?`
		_, err = tx.Exec(updateQuery, payout, now, userID)
	} else {
		updateQuery = `UPDATE user_game_balances 
			SET games_played = games_played + 1, 
			    total_lost = total_lost + ?, 
			    updated_at = ? 
			WHERE user_id = ?`
		_, err = tx.Exec(updateQuery, betAmount, now, userID)
	}
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &GameRecord{
		ID:        recordID,
		UserID:    userID,
		GameType:  gameType,
		BetAmount: betAmount,
		Result:    result,
		Payout:    payout,
		NetProfit: netProfit,
		Details:   details,
		CreatedAt: now,
	}, nil
}

// GetGameRecords retrieves paginated game records for a user
// Requirements: 1.5, 1.6
func GetGameRecords(userID int64, limit, offset int) ([]*GameRecord, int, error) {
	// Validate and cap limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	// Get total count
	var total int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM game_records WHERE user_id = ?`,
		userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get records sorted by created_at DESC
	rows, err := db.Query(
		`SELECT id, user_id, game_type, bet_amount, result, payout, net_profit, details, created_at
		 FROM game_records WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var records []*GameRecord
	for rows.Next() {
		record := &GameRecord{}
		var details sql.NullString

		err := rows.Scan(&record.ID, &record.UserID, &record.GameType, &record.BetAmount,
			&record.Result, &record.Payout, &record.NetProfit, &details, &record.CreatedAt)
		if err != nil {
			return nil, 0, err
		}

		if details.Valid {
			record.Details = json.RawMessage(details.String)
		}

		records = append(records, record)
	}

	return records, total, nil
}

// GetGameStats retrieves aggregated game statistics for a user
// Requirements: 2.1, 2.4, 2.5
func GetGameStats(userID int64) (*GameStats, error) {
	// Query stats directly from game_records table for accuracy
	var gamesPlayed, wins int
	var totalPayout, totalBet float64

	// Get total games and wins from game_records
	err := db.QueryRow(
		`SELECT 
			COUNT(*) as games_played,
			SUM(CASE WHEN result = 'win' THEN 1 ELSE 0 END) as wins,
			COALESCE(SUM(payout), 0) as total_payout,
			COALESCE(SUM(bet_amount), 0) as total_bet
		 FROM game_records WHERE user_id = ?`,
		userID,
	).Scan(&gamesPlayed, &wins, &totalPayout, &totalBet)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// If no records, return zero stats
	if gamesPlayed == 0 {
		return &GameStats{
			GamesPlayed: 0,
			Wins:        0,
			Losses:      0,
			WinRate:     "0.0",
			NetProfit:   "0.00",
			TotalWon:    "0.00",
			TotalLost:   "0.00",
		}, nil
	}

	// Calculate losses
	losses := gamesPlayed - wins

	// Calculate win rate with one decimal precision
	var winRate float64
	if gamesPlayed > 0 {
		winRate = float64(wins) / float64(gamesPlayed) * 100
	}

	// Calculate net profit (total payout - total bet)
	netProfit := totalPayout - totalBet

	return &GameStats{
		GamesPlayed: gamesPlayed,
		Wins:        wins,
		Losses:      losses,
		WinRate:     fmt.Sprintf("%.1f", winRate),
		NetProfit:   fmt.Sprintf("%.2f", netProfit),
		TotalWon:    fmt.Sprintf("%.2f", totalPayout),
		TotalLost:   fmt.Sprintf("%.2f", totalBet),
	}, nil
}

// GetLeaderboard retrieves the global leaderboard
// Requirements: 3.1, 3.2, 4.2
func GetLeaderboard(currentUserID int64, sortBy string, limit int) ([]*LeaderboardEntry, *LeaderboardEntry, int, error) {
	// Validate and set defaults
	if limit <= 0 {
		limit = 10
	}
	if sortBy != "winnings" && sortBy != "games" {
		sortBy = "winnings"
	}

	// Determine sort column
	var orderBy string
	if sortBy == "winnings" {
		orderBy = "(total_won - total_lost) DESC"
	} else {
		orderBy = "games_played DESC"
	}

	// Get total players count
	var totalPlayers int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM user_game_balances WHERE games_played > 0`,
	).Scan(&totalPlayers)
	if err != nil {
		return nil, nil, 0, err
	}

	// Get top N entries with rank
	query := fmt.Sprintf(`
		SELECT ugb.user_id, u.username, (ugb.total_won - ugb.total_lost) as total_winnings, ugb.games_played
		FROM user_game_balances ugb
		JOIN users u ON ugb.user_id = u.id
		WHERE ugb.games_played > 0
		ORDER BY %s
		LIMIT ?
	`, orderBy)

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, nil, 0, err
	}
	defer rows.Close()

	var entries []*LeaderboardEntry
	rank := 0
	currentUserInTop := false

	for rows.Next() {
		rank++
		entry := &LeaderboardEntry{Rank: rank}
		err := rows.Scan(&entry.UserID, &entry.Username, &entry.TotalWinnings, &entry.GamesPlayed)
		if err != nil {
			return nil, nil, 0, err
		}
		entries = append(entries, entry)

		if entry.UserID == currentUserID {
			currentUserInTop = true
		}
	}

	// Get current user's entry if not in top N
	var currentUserEntry *LeaderboardEntry
	if !currentUserInTop && currentUserID > 0 {
		// Get current user's rank and stats
		rankQuery := fmt.Sprintf(`
			SELECT COUNT(*) + 1 as rank
			FROM user_game_balances
			WHERE games_played > 0 AND %s > (
				SELECT COALESCE(%s, 0)
				FROM user_game_balances
				WHERE user_id = ?
			)
		`, func() string {
			if sortBy == "winnings" {
				return "(total_won - total_lost)"
			}
			return "games_played"
		}(), func() string {
			if sortBy == "winnings" {
				return "(total_won - total_lost)"
			}
			return "games_played"
		}())

		var userRank int
		err := db.QueryRow(rankQuery, currentUserID).Scan(&userRank)
		if err != nil && err != sql.ErrNoRows {
			return nil, nil, 0, err
		}

		// Get current user's stats
		var username string
		var totalWinnings float64
		var gamesPlayed int
		err = db.QueryRow(`
			SELECT u.username, (ugb.total_won - ugb.total_lost) as total_winnings, ugb.games_played
			FROM user_game_balances ugb
			JOIN users u ON ugb.user_id = u.id
			WHERE ugb.user_id = ? AND ugb.games_played > 0
		`, currentUserID).Scan(&username, &totalWinnings, &gamesPlayed)

		if err == nil {
			currentUserEntry = &LeaderboardEntry{
				Rank:          userRank,
				UserID:        currentUserID,
				Username:      username,
				TotalWinnings: totalWinnings,
				GamesPlayed:   gamesPlayed,
			}
		}
	}

	return entries, currentUserEntry, totalPlayers, nil
}
