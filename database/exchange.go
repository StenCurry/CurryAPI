package database

import (
	"database/sql"
	"time"
)

// ExchangeRecord represents a game coin to USD exchange record
type ExchangeRecord struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	GameCoinsAmount float64   `json:"game_coins_amount"`
	USDAmount       float64   `json:"usd_amount"`
	ExchangeRate    float64   `json:"exchange_rate"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

// ExchangeRecordWithUser represents an exchange record with user info for admin view
type ExchangeRecordWithUser struct {
	ExchangeRecord
	Username string `json:"username"`
	Email    string `json:"email"`
}

// ExchangeStats represents exchange statistics for admin
type ExchangeStats struct {
	TotalCount int     `json:"total_count"`
	TotalUSD   float64 `json:"total_usd"`
}

// ExchangeGameCoins exchanges game coins for account balance (USD)
// This is an atomic transaction that:
// 1. Validates the exchange amount
// 2. Checks daily exchange limit
// 3. Deducts game coins from user's game balance
// 4. Adds USD to user's account balance
// 5. Creates exchange record
// 6. Creates transaction records for both game coins and account balance
// Requirements: 2.1, 2.4, 2.5, 5.1
func ExchangeGameCoins(userID int64, amount float64) (*ExchangeRecord, error) {
	// Validate amount
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if amount < MinimumExchangeAmount {
		return nil, ErrBelowMinimumExchange
	}

	amount = roundToTwoDecimals(amount)
	usdAmount := amount * ExchangeRate // 1:1 rate

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Check daily exchange limit
	todayExchanged, err := getTodayExchangeAmountTx(tx, userID)
	if err != nil {
		return nil, err
	}
	if todayExchanged+amount > DailyExchangeLimit {
		return nil, ErrDailyLimitExceeded
	}

	// Get current game coin balance with lock
	var currentGameBalance float64
	err = tx.QueryRow(
		`SELECT balance FROM user_game_balances WHERE user_id = ? FOR UPDATE`,
		userID,
	).Scan(&currentGameBalance)

	if err == sql.ErrNoRows {
		return nil, ErrGameBalanceNotFound
	}
	if err != nil {
		return nil, err
	}

	// Check sufficient game coins
	if currentGameBalance < amount {
		return nil, ErrInsufficientGameCoins
	}

	now := time.Now()
	newGameBalance := roundToTwoDecimals(currentGameBalance - amount)

	// Deduct game coins
	_, err = tx.Exec(
		`UPDATE user_game_balances SET balance = ?, total_exchanged = total_exchanged + ?, updated_at = ?
		 WHERE user_id = ?`,
		newGameBalance, amount, now, userID,
	)
	if err != nil {
		return nil, err
	}

	// Create game coin transaction record (negative amount for exchange)
	_, err = tx.Exec(
		`INSERT INTO game_coin_transactions (user_id, type, game_type, amount, balance_after, description, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, GameTxTypeExchange, nil, -amount, newGameBalance, "Exchange to account balance", now,
	)
	if err != nil {
		return nil, err
	}

	// Get current account balance with lock
	var currentAccountBalance float64
	var accountStatus string
	err = tx.QueryRow(
		`SELECT balance, status FROM user_balances WHERE user_id = ? FOR UPDATE`,
		userID,
	).Scan(&currentAccountBalance, &accountStatus)

	if err == sql.ErrNoRows {
		return nil, ErrBalanceNotFound
	}
	if err != nil {
		return nil, err
	}

	newAccountBalance := currentAccountBalance + usdAmount
	newStatus := accountStatus
	// If balance was exhausted and now positive, set to active
	if accountStatus == BalanceStatusExhausted && newAccountBalance > 0 {
		newStatus = BalanceStatusActive
	}

	// Add USD to account balance
	_, err = tx.Exec(
		`UPDATE user_balances SET balance = ?, status = ?, total_recharged = total_recharged + ?, updated_at = ?
		 WHERE user_id = ?`,
		newAccountBalance, newStatus, usdAmount, now, userID,
	)
	if err != nil {
		return nil, err
	}

	// Create account balance transaction record
	_, err = tx.Exec(
		`INSERT INTO balance_transactions (user_id, type, amount, balance_after, tokens, description, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, "game_exchange", usdAmount, newAccountBalance, 0, "Exchange from game coins", now,
	)
	if err != nil {
		return nil, err
	}

	// Re-enable tokens if status changed from exhausted to active
	if accountStatus == BalanceStatusExhausted && newStatus == BalanceStatusActive {
		_, err = tx.Exec(`UPDATE api_keys SET is_active = TRUE WHERE user_id = ?`, userID)
		if err != nil {
			return nil, err
		}
	}

	// Create exchange record
	result, err := tx.Exec(
		`INSERT INTO exchange_records (user_id, game_coins_amount, usd_amount, exchange_rate, status, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		userID, amount, usdAmount, ExchangeRate, "completed", now,
	)
	if err != nil {
		return nil, err
	}

	exchangeID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &ExchangeRecord{
		ID:              exchangeID,
		UserID:          userID,
		GameCoinsAmount: amount,
		USDAmount:       usdAmount,
		ExchangeRate:    ExchangeRate,
		Status:          "completed",
		CreatedAt:       now,
	}, nil
}


// getTodayExchangeAmountTx gets today's total exchange amount within a transaction
func getTodayExchangeAmountTx(tx *sql.Tx, userID int64) (float64, error) {
	var total sql.NullFloat64
	today := time.Now().Format("2006-01-02")

	err := tx.QueryRow(
		`SELECT SUM(game_coins_amount) FROM exchange_records 
		 WHERE user_id = ? AND DATE(created_at) = ? AND status = 'completed'`,
		userID, today,
	).Scan(&total)

	if err != nil {
		return 0, err
	}

	if total.Valid {
		return total.Float64, nil
	}
	return 0, nil
}

// GetTodayExchangeAmount gets today's total exchange amount for a user
// Requirements: 2.7
func GetTodayExchangeAmount(userID int64) (float64, error) {
	var total sql.NullFloat64
	today := time.Now().Format("2006-01-02")

	err := db.QueryRow(
		`SELECT SUM(game_coins_amount) FROM exchange_records 
		 WHERE user_id = ? AND DATE(created_at) = ? AND status = 'completed'`,
		userID, today,
	).Scan(&total)

	if err != nil {
		return 0, err
	}

	if total.Valid {
		return total.Float64, nil
	}
	return 0, nil
}

// GetExchangeHistory retrieves paginated exchange history for a user
// Records are sorted by created_at in descending order
// Requirements: 3.1, 3.3
func GetExchangeHistory(userID int64, limit, offset int) ([]*ExchangeRecord, int, error) {
	// Get total count
	var total int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM exchange_records WHERE user_id = ?`,
		userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get records sorted by created_at DESC
	rows, err := db.Query(
		`SELECT id, user_id, game_coins_amount, usd_amount, exchange_rate, status, created_at
		 FROM exchange_records WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var records []*ExchangeRecord
	for rows.Next() {
		record := &ExchangeRecord{}
		err := rows.Scan(&record.ID, &record.UserID, &record.GameCoinsAmount,
			&record.USDAmount, &record.ExchangeRate, &record.Status, &record.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		records = append(records, record)
	}

	return records, total, nil
}

// GetAllExchangeRecords retrieves all exchange records with optional filters for admin
// Supports filtering by user ID and date range
// Requirements: 6.1, 6.2, 6.3, 6.4
func GetAllExchangeRecords(userID *int64, startDate, endDate *time.Time, limit, offset int) ([]*ExchangeRecordWithUser, int, error) {
	// Build query with optional filters
	baseQuery := `FROM exchange_records er JOIN users u ON er.user_id = u.id WHERE 1=1`
	args := []interface{}{}

	if userID != nil {
		baseQuery += ` AND er.user_id = ?`
		args = append(args, *userID)
	}
	if startDate != nil {
		baseQuery += ` AND er.created_at >= ?`
		args = append(args, *startDate)
	}
	if endDate != nil {
		baseQuery += ` AND er.created_at <= ?`
		args = append(args, *endDate)
	}

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) ` + baseQuery
	err := db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get records with user info
	selectQuery := `SELECT er.id, er.user_id, er.game_coins_amount, er.usd_amount, er.exchange_rate, 
	                er.status, er.created_at, u.username, u.email ` + baseQuery +
		` ORDER BY er.created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := db.Query(selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var records []*ExchangeRecordWithUser
	for rows.Next() {
		record := &ExchangeRecordWithUser{}
		err := rows.Scan(&record.ID, &record.UserID, &record.GameCoinsAmount,
			&record.USDAmount, &record.ExchangeRate, &record.Status, &record.CreatedAt,
			&record.Username, &record.Email)
		if err != nil {
			return nil, 0, err
		}
		records = append(records, record)
	}

	return records, total, nil
}

// GetExchangeStats retrieves exchange statistics for admin
// Requirements: 6.5
func GetExchangeStats() (*ExchangeStats, error) {
	stats := &ExchangeStats{}

	err := db.QueryRow(
		`SELECT COUNT(*), COALESCE(SUM(usd_amount), 0) FROM exchange_records WHERE status = 'completed'`,
	).Scan(&stats.TotalCount, &stats.TotalUSD)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// ExchangeUSDToGameCoins exchanges account balance (USD) for game coins
// This is an atomic transaction that:
// 1. Validates the exchange amount
// 2. Deducts USD from user's account balance
// 3. Adds game coins to user's game balance
// 4. Creates exchange record
// 5. Creates transaction records for both account balance and game coins
func ExchangeUSDToGameCoins(userID int64, usdAmount float64) (*ExchangeRecord, error) {
	// Validate amount
	if usdAmount <= 0 {
		return nil, ErrInvalidAmount
	}
	if usdAmount < MinimumExchangeAmount {
		return nil, ErrBelowMinimumExchange
	}

	usdAmount = roundToTwoDecimals(usdAmount)
	gameCoinsAmount := usdAmount * ExchangeRate // 1:1 rate

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now()

	// Get current account balance with lock
	var currentAccountBalance float64
	var accountStatus string
	err = tx.QueryRow(
		`SELECT balance, status FROM user_balances WHERE user_id = ? FOR UPDATE`,
		userID,
	).Scan(&currentAccountBalance, &accountStatus)

	if err == sql.ErrNoRows {
		return nil, ErrBalanceNotFound
	}
	if err != nil {
		return nil, err
	}

	// Check sufficient account balance
	if currentAccountBalance < usdAmount {
		return nil, ErrInsufficientBalance
	}

	newAccountBalance := roundToTwoDecimals(currentAccountBalance - usdAmount)
	newStatus := accountStatus
	// If balance becomes zero or negative, set to exhausted
	if newAccountBalance <= 0 {
		newStatus = BalanceStatusExhausted
	}

	// Deduct USD from account balance
	_, err = tx.Exec(
		`UPDATE user_balances SET balance = ?, status = ?, updated_at = ?
		 WHERE user_id = ?`,
		newAccountBalance, newStatus, now, userID,
	)
	if err != nil {
		return nil, err
	}

	// Create account balance transaction record (negative amount for exchange)
	_, err = tx.Exec(
		`INSERT INTO balance_transactions (user_id, type, amount, balance_after, tokens, description, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, "game_purchase", -usdAmount, newAccountBalance, 0, "Purchase game coins", now,
	)
	if err != nil {
		return nil, err
	}

	// Disable API keys if balance exhausted
	if newStatus == BalanceStatusExhausted {
		_, err = tx.Exec(`UPDATE api_keys SET is_active = FALSE WHERE user_id = ?`, userID)
		if err != nil {
			return nil, err
		}
	}

	// Get current game coin balance with lock
	var currentGameBalance float64
	err = tx.QueryRow(
		`SELECT balance FROM user_game_balances WHERE user_id = ? FOR UPDATE`,
		userID,
	).Scan(&currentGameBalance)

	if err == sql.ErrNoRows {
		// Create game balance if not exists
		_, err = tx.Exec(
			`INSERT INTO user_game_balances (user_id, balance, total_won, total_lost, total_exchanged, games_played, created_at, updated_at)
			 VALUES (?, ?, 0, 0, 0, 0, ?, ?)`,
			userID, gameCoinsAmount, now, now,
		)
		if err != nil {
			return nil, err
		}
		currentGameBalance = 0
	} else if err != nil {
		return nil, err
	}

	newGameBalance := roundToTwoDecimals(currentGameBalance + gameCoinsAmount)

	// Add game coins (only if balance already existed)
	if currentGameBalance > 0 || err == nil {
		_, err = tx.Exec(
			`UPDATE user_game_balances SET balance = ?, updated_at = ?
			 WHERE user_id = ?`,
			newGameBalance, now, userID,
		)
		if err != nil {
			return nil, err
		}
	}

	// Create game coin transaction record (positive amount for purchase)
	_, err = tx.Exec(
		`INSERT INTO game_coin_transactions (user_id, type, game_type, amount, balance_after, description, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, "purchase", nil, gameCoinsAmount, newGameBalance, "Purchased with account balance", now,
	)
	if err != nil {
		return nil, err
	}

	// Create exchange record (with negative game_coins_amount to indicate reverse direction)
	result, err := tx.Exec(
		`INSERT INTO exchange_records (user_id, game_coins_amount, usd_amount, exchange_rate, status, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		userID, -gameCoinsAmount, -usdAmount, ExchangeRate, "completed", now,
	)
	if err != nil {
		return nil, err
	}

	exchangeID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &ExchangeRecord{
		ID:              exchangeID,
		UserID:          userID,
		GameCoinsAmount: gameCoinsAmount,  // Return positive for display
		USDAmount:       usdAmount,        // Return positive for display
		ExchangeRate:    ExchangeRate,
		Status:          "completed",
		CreatedAt:       now,
	}, nil
}
