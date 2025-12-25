package database

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"math/big"
	"time"
)

// Constants for balance system
const (
	InitialBalance     = 50.0      // Initial balance in USD
	TokensPerDollar    = 1000000   // 1 USD = 1,000,000 tokens
	BalanceStatusActive    = "active"
	BalanceStatusExhausted = "exhausted"
	ReferralCodeLength     = 6 // 6-character referral code with uppercase letters and numbers
)

// Transaction types
const (
	TransactionTypeInitial       = "initial"
	TransactionTypeAPIUsage      = "api_usage"
	TransactionTypeReferralBonus = "referral_bonus"
	TransactionTypeAdminAdjust   = "admin_adjust"
)

// Errors
var (
	ErrBalanceNotFound      = errors.New("balance record not found")
	ErrInsufficientBalance  = errors.New("insufficient balance")
	ErrBalanceExhausted     = errors.New("balance exhausted")
	ErrReferralCodeNotFound = errors.New("referral code not found")
	ErrReferralCodeExists   = errors.New("referral code already exists")
)

// UserBalance represents a user's balance record
type UserBalance struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	Balance        float64   `json:"balance"`
	Status         string    `json:"status"`
	ReferralCode   string    `json:"referral_code"`
	TotalConsumed  float64   `json:"total_consumed"`
	TotalRecharged float64   `json:"total_recharged"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// BalanceTransaction represents a balance transaction record
type BalanceTransaction struct {
	ID            int64      `json:"id"`
	UserID        int64      `json:"user_id"`
	Type          string     `json:"type"`
	Amount        float64    `json:"amount"`
	BalanceAfter  float64    `json:"balance_after"`
	Tokens        int        `json:"tokens"`
	Description   string     `json:"description"`
	RelatedUserID *int64     `json:"related_user_id,omitempty"`
	AdminID       *int64     `json:"admin_id,omitempty"`
	APIToken      string     `json:"api_token,omitempty"`
	Model         string     `json:"model,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}


// generateReferralCode generates a unique 6-character alphanumeric referral code (uppercase letters and numbers)
func generateReferralCode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, ReferralCodeLength)
	
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		code[i] = charset[n.Int64()]
	}
	
	return string(code), nil
}

// generateUniqueReferralCode generates a referral code that doesn't exist in the database
func generateUniqueReferralCode() (string, error) {
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		code, err := generateReferralCode()
		if err != nil {
			return "", err
		}
		
		// Check if code already exists
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM user_balances WHERE referral_code = ?)", code).Scan(&exists)
		if err != nil {
			return "", err
		}
		
		if !exists {
			return code, nil
		}
	}
	
	return "", errors.New("failed to generate unique referral code after max attempts")
}

// CreateUserBalance creates a new balance record for a user with initial balance of $50
// Requirements: 1.1, 4.1, 4.2
func CreateUserBalance(userID int64) (*UserBalance, error) {
	// Generate unique referral code
	referralCode, err := generateUniqueReferralCode()
	if err != nil {
		return nil, err
	}
	
	now := time.Now()
	
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	
	// Insert balance record
	result, err := tx.Exec(
		`INSERT INTO user_balances (user_id, balance, status, referral_code, total_consumed, total_recharged, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, InitialBalance, BalanceStatusActive, referralCode, 0, InitialBalance, now, now,
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
		`INSERT INTO balance_transactions (user_id, type, amount, balance_after, tokens, description, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, TransactionTypeInitial, InitialBalance, InitialBalance, 0, "Initial balance", now,
	)
	if err != nil {
		return nil, err
	}
	
	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	
	return &UserBalance{
		ID:             balanceID,
		UserID:         userID,
		Balance:        InitialBalance,
		Status:         BalanceStatusActive,
		ReferralCode:   referralCode,
		TotalConsumed:  0,
		TotalRecharged: InitialBalance,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}


// GetUserBalance retrieves a user's balance record
// Requirements: 6.1
func GetUserBalance(userID int64) (*UserBalance, error) {
	balance := &UserBalance{}
	
	err := db.QueryRow(
		`SELECT id, user_id, balance, status, referral_code, total_consumed, total_recharged, created_at, updated_at
		 FROM user_balances WHERE user_id = ?`,
		userID,
	).Scan(&balance.ID, &balance.UserID, &balance.Balance, &balance.Status, &balance.ReferralCode,
		&balance.TotalConsumed, &balance.TotalRecharged, &balance.CreatedAt, &balance.UpdatedAt)
	
	if err == sql.ErrNoRows {
		return nil, ErrBalanceNotFound
	}
	if err != nil {
		return nil, err
	}
	
	return balance, nil
}

// CalculateCost calculates the cost in USD from token count
// $1 = 1,000,000 tokens
// Requirements: 2.1
func CalculateCost(tokens int) float64 {
	return float64(tokens) / float64(TokensPerDollar)
}


// DeductBalance deducts balance based on token usage and creates a transaction record
// Requirements: 2.1, 2.2, 2.3
func DeductBalance(userID int64, tokens int, apiToken, model string) (*BalanceTransaction, error) {
	cost := CalculateCost(tokens)
	
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	
	// Get current balance with lock
	var currentBalance float64
	var status string
	err = tx.QueryRow(
		`SELECT balance, status FROM user_balances WHERE user_id = ? FOR UPDATE`,
		userID,
	).Scan(&currentBalance, &status)
	
	if err == sql.ErrNoRows {
		return nil, ErrBalanceNotFound
	}
	if err != nil {
		return nil, err
	}
	
	// Calculate new balance
	newBalance := currentBalance - cost
	newStatus := status
	
	// Check if balance becomes exhausted
	if newBalance <= 0 {
		newStatus = BalanceStatusExhausted
	}
	
	now := time.Now()
	
	// Update balance
	_, err = tx.Exec(
		`UPDATE user_balances SET balance = ?, status = ?, total_consumed = total_consumed + ?, updated_at = ?
		 WHERE user_id = ?`,
		newBalance, newStatus, cost, now, userID,
	)
	if err != nil {
		return nil, err
	}
	
	// Create transaction record
	description := "API usage"
	if model != "" {
		description = "API usage: " + model
	}
	
	result, err := tx.Exec(
		`INSERT INTO balance_transactions (user_id, type, amount, balance_after, tokens, description, api_token, model, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, TransactionTypeAPIUsage, -cost, newBalance, tokens, description, apiToken, model, now,
	)
	if err != nil {
		return nil, err
	}
	
	txID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	
	// If status changed to exhausted, disable all user tokens
	if newStatus == BalanceStatusExhausted && status != BalanceStatusExhausted {
		_, err = tx.Exec(`UPDATE api_keys SET is_active = FALSE WHERE user_id = ?`, userID)
		if err != nil {
			return nil, err
		}
	}
	
	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	
	return &BalanceTransaction{
		ID:           txID,
		UserID:       userID,
		Type:         TransactionTypeAPIUsage,
		Amount:       -cost,
		BalanceAfter: newBalance,
		Tokens:       tokens,
		Description:  description,
		APIToken:     apiToken,
		Model:        model,
		CreatedAt:    now,
	}, nil
}


// AddBalance adds balance to a user's account and creates a transaction record
// Re-enables tokens if status changes from exhausted to active
// Requirements: 3.3, 8.1, 8.2
func AddBalance(userID int64, amount float64, description string, adminID *int64, relatedUserID *int64, txType string) (*BalanceTransaction, error) {
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	
	// Get current balance with lock
	var currentBalance float64
	var currentStatus string
	err = tx.QueryRow(
		`SELECT balance, status FROM user_balances WHERE user_id = ? FOR UPDATE`,
		userID,
	).Scan(&currentBalance, &currentStatus)
	
	if err == sql.ErrNoRows {
		return nil, ErrBalanceNotFound
	}
	if err != nil {
		return nil, err
	}
	
	// Calculate new balance
	newBalance := currentBalance + amount
	newStatus := currentStatus
	
	// If balance was exhausted and now positive, set to active
	if currentStatus == BalanceStatusExhausted && newBalance > 0 {
		newStatus = BalanceStatusActive
	}
	
	now := time.Now()
	
	// Update balance
	_, err = tx.Exec(
		`UPDATE user_balances SET balance = ?, status = ?, total_recharged = total_recharged + ?, updated_at = ?
		 WHERE user_id = ?`,
		newBalance, newStatus, amount, now, userID,
	)
	if err != nil {
		return nil, err
	}
	
	// Create transaction record
	result, err := tx.Exec(
		`INSERT INTO balance_transactions (user_id, type, amount, balance_after, tokens, description, admin_id, related_user_id, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, txType, amount, newBalance, 0, description, adminID, relatedUserID, now,
	)
	if err != nil {
		return nil, err
	}
	
	txID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	
	// If status changed from exhausted to active, re-enable all user tokens
	if currentStatus == BalanceStatusExhausted && newStatus == BalanceStatusActive {
		_, err = tx.Exec(`UPDATE api_keys SET is_active = TRUE WHERE user_id = ?`, userID)
		if err != nil {
			return nil, err
		}
	}
	
	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	
	return &BalanceTransaction{
		ID:            txID,
		UserID:        userID,
		Type:          txType,
		Amount:        amount,
		BalanceAfter:  newBalance,
		Tokens:        0,
		Description:   description,
		AdminID:       adminID,
		RelatedUserID: relatedUserID,
		CreatedAt:     now,
	}, nil
}


// UpdateBalanceStatus updates the balance status and handles token enable/disable
// Requirements: 2.4, 3.1
func UpdateBalanceStatus(userID int64, status string) error {
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// Get current status
	var currentStatus string
	err = tx.QueryRow(
		`SELECT status FROM user_balances WHERE user_id = ? FOR UPDATE`,
		userID,
	).Scan(&currentStatus)
	
	if err == sql.ErrNoRows {
		return ErrBalanceNotFound
	}
	if err != nil {
		return err
	}
	
	// Update status
	_, err = tx.Exec(
		`UPDATE user_balances SET status = ?, updated_at = ? WHERE user_id = ?`,
		status, time.Now(), userID,
	)
	if err != nil {
		return err
	}
	
	// Handle token status based on balance status change
	if status == BalanceStatusExhausted && currentStatus != BalanceStatusExhausted {
		// Disable all user tokens when balance becomes exhausted
		_, err = tx.Exec(`UPDATE api_keys SET is_active = FALSE WHERE user_id = ?`, userID)
		if err != nil {
			return err
		}
	} else if status == BalanceStatusActive && currentStatus == BalanceStatusExhausted {
		// Re-enable all user tokens when balance becomes active
		_, err = tx.Exec(`UPDATE api_keys SET is_active = TRUE WHERE user_id = ?`, userID)
		if err != nil {
			return err
		}
	}
	
	// Commit transaction
	return tx.Commit()
}

// CheckAndUpdateBalanceStatus checks if balance is <= 0 and updates status to exhausted
// Returns true if status was changed to exhausted
// Requirements: 2.4, 3.1
func CheckAndUpdateBalanceStatus(userID int64) (bool, error) {
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	
	// Get current balance and status
	var balance float64
	var status string
	err = tx.QueryRow(
		`SELECT balance, status FROM user_balances WHERE user_id = ? FOR UPDATE`,
		userID,
	).Scan(&balance, &status)
	
	if err == sql.ErrNoRows {
		return false, ErrBalanceNotFound
	}
	if err != nil {
		return false, err
	}
	
	// If balance <= 0 and not already exhausted, update status
	if balance <= 0 && status != BalanceStatusExhausted {
		_, err = tx.Exec(
			`UPDATE user_balances SET status = ?, updated_at = ? WHERE user_id = ?`,
			BalanceStatusExhausted, time.Now(), userID,
		)
		if err != nil {
			return false, err
		}
		
		// Disable all user tokens
		_, err = tx.Exec(`UPDATE api_keys SET is_active = FALSE WHERE user_id = ?`, userID)
		if err != nil {
			return false, err
		}
		
		if err := tx.Commit(); err != nil {
			return false, err
		}
		return true, nil
	}
	
	return false, tx.Commit()
}

// GetBalanceTransactions retrieves paginated transaction history for a user
// Requirements: 6.2, 6.3
func GetBalanceTransactions(userID int64, limit, offset int) ([]*BalanceTransaction, int, error) {
	// Get total count
	var total int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM balance_transactions WHERE user_id = ?`,
		userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	// Get transactions
	rows, err := db.Query(
		`SELECT id, user_id, type, amount, balance_after, tokens, description, related_user_id, admin_id, api_token, model, created_at
		 FROM balance_transactions WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var transactions []*BalanceTransaction
	for rows.Next() {
		tx := &BalanceTransaction{}
		var relatedUserID, adminID sql.NullInt64
		var apiToken, model sql.NullString
		
		err := rows.Scan(&tx.ID, &tx.UserID, &tx.Type, &tx.Amount, &tx.BalanceAfter, &tx.Tokens,
			&tx.Description, &relatedUserID, &adminID, &apiToken, &model, &tx.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		
		if relatedUserID.Valid {
			tx.RelatedUserID = &relatedUserID.Int64
		}
		if adminID.Valid {
			tx.AdminID = &adminID.Int64
		}
		if apiToken.Valid {
			tx.APIToken = apiToken.String
		}
		if model.Valid {
			tx.Model = model.String
		}
		
		transactions = append(transactions, tx)
	}
	
	return transactions, total, nil
}

// ============================================
// Referral System Functions
// ============================================

// ReferralBonus is the bonus amount for referrals in USD
const ReferralBonus = 50.0

// Referral represents a referral relationship record
type Referral struct {
	ID          int64     `json:"id"`
	ReferrerID  int64     `json:"referrer_id"`
	RefereeID   int64     `json:"referee_id"`
	BonusAmount float64   `json:"bonus_amount"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// ReferralStats represents referral statistics for a user
type ReferralStats struct {
	TotalReferrals int     `json:"total_referrals"`
	TotalBonus     float64 `json:"total_bonus"`
}

// ReferredUser represents a referred user with registration date
type ReferredUser struct {
	UserID       int64     `json:"user_id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	RegisteredAt time.Time `json:"registered_at"`
	BonusAmount  float64   `json:"bonus_amount"`
}

// GetUserByReferralCode finds a user by their referral code
// Requirements: 5.1
func GetUserByReferralCode(referralCode string) (*UserBalance, error) {
	balance := &UserBalance{}
	
	err := db.QueryRow(
		`SELECT id, user_id, balance, status, referral_code, total_consumed, total_recharged, created_at, updated_at
		 FROM user_balances WHERE referral_code = ?`,
		referralCode,
	).Scan(&balance.ID, &balance.UserID, &balance.Balance, &balance.Status, &balance.ReferralCode,
		&balance.TotalConsumed, &balance.TotalRecharged, &balance.CreatedAt, &balance.UpdatedAt)
	
	if err == sql.ErrNoRows {
		return nil, ErrReferralCodeNotFound
	}
	if err != nil {
		return nil, err
	}
	
	return balance, nil
}


// Errors for referral system
var (
	ErrSelfReferral       = errors.New("self referral not allowed")
	ErrReferralExists     = errors.New("referral relationship already exists")
)

// CreateReferral creates a referral relationship record
// Requirements: 5.3
func CreateReferral(referrerID, refereeID int64, bonusAmount float64) (*Referral, error) {
	// Prevent self-referral
	if referrerID == refereeID {
		return nil, ErrSelfReferral
	}
	
	now := time.Now()
	
	result, err := db.Exec(
		`INSERT INTO referrals (referrer_id, referee_id, bonus_amount, status, created_at)
		 VALUES (?, ?, ?, 'completed', ?)`,
		referrerID, refereeID, bonusAmount, now,
	)
	if err != nil {
		// Check for duplicate entry (referee_id is unique)
		return nil, err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	
	return &Referral{
		ID:          id,
		ReferrerID:  referrerID,
		RefereeID:   refereeID,
		BonusAmount: bonusAmount,
		Status:      "completed",
		CreatedAt:   now,
	}, nil
}


// ProcessReferralBonus processes the referral bonus for both referrer and referee
// Adds $50 to referrer balance and $50 to referee balance (extra)
// Creates transaction records for both users
// Requirements: 5.1, 5.2, 5.4
func ProcessReferralBonus(referralCode string, refereeID int64) (*Referral, error) {
	// Find referrer by referral code
	referrerBalance, err := GetUserByReferralCode(referralCode)
	if err != nil {
		return nil, err
	}
	
	referrerID := referrerBalance.UserID
	
	// Prevent self-referral
	if referrerID == refereeID {
		return nil, ErrSelfReferral
	}
	
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	
	now := time.Now()
	
	// 1. Add bonus to referrer's balance
	var referrerCurrentBalance float64
	var referrerStatus string
	err = tx.QueryRow(
		`SELECT balance, status FROM user_balances WHERE user_id = ? FOR UPDATE`,
		referrerID,
	).Scan(&referrerCurrentBalance, &referrerStatus)
	if err != nil {
		return nil, err
	}
	
	referrerNewBalance := referrerCurrentBalance + ReferralBonus
	referrerNewStatus := referrerStatus
	if referrerStatus == BalanceStatusExhausted && referrerNewBalance > 0 {
		referrerNewStatus = BalanceStatusActive
	}
	
	_, err = tx.Exec(
		`UPDATE user_balances SET balance = ?, status = ?, total_recharged = total_recharged + ?, updated_at = ?
		 WHERE user_id = ?`,
		referrerNewBalance, referrerNewStatus, ReferralBonus, now, referrerID,
	)
	if err != nil {
		return nil, err
	}
	
	// Create transaction record for referrer
	_, err = tx.Exec(
		`INSERT INTO balance_transactions (user_id, type, amount, balance_after, tokens, description, related_user_id, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		referrerID, TransactionTypeReferralBonus, ReferralBonus, referrerNewBalance, 0,
		"Referral bonus - new user registered", refereeID, now,
	)
	if err != nil {
		return nil, err
	}
	
	// Re-enable referrer's tokens if status changed from exhausted to active
	if referrerStatus == BalanceStatusExhausted && referrerNewStatus == BalanceStatusActive {
		_, err = tx.Exec(`UPDATE api_keys SET is_active = TRUE WHERE user_id = ?`, referrerID)
		if err != nil {
			return nil, err
		}
	}
	
	// 2. Add bonus to referee's balance
	var refereeCurrentBalance float64
	var refereeStatus string
	err = tx.QueryRow(
		`SELECT balance, status FROM user_balances WHERE user_id = ? FOR UPDATE`,
		refereeID,
	).Scan(&refereeCurrentBalance, &refereeStatus)
	if err != nil {
		return nil, err
	}
	
	refereeNewBalance := refereeCurrentBalance + ReferralBonus
	refereeNewStatus := refereeStatus
	if refereeStatus == BalanceStatusExhausted && refereeNewBalance > 0 {
		refereeNewStatus = BalanceStatusActive
	}
	
	_, err = tx.Exec(
		`UPDATE user_balances SET balance = ?, status = ?, total_recharged = total_recharged + ?, updated_at = ?
		 WHERE user_id = ?`,
		refereeNewBalance, refereeNewStatus, ReferralBonus, now, refereeID,
	)
	if err != nil {
		return nil, err
	}
	
	// Create transaction record for referee
	_, err = tx.Exec(
		`INSERT INTO balance_transactions (user_id, type, amount, balance_after, tokens, description, related_user_id, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		refereeID, TransactionTypeReferralBonus, ReferralBonus, refereeNewBalance, 0,
		"Referral bonus - registered with referral code", referrerID, now,
	)
	if err != nil {
		return nil, err
	}
	
	// Re-enable referee's tokens if status changed from exhausted to active
	if refereeStatus == BalanceStatusExhausted && refereeNewStatus == BalanceStatusActive {
		_, err = tx.Exec(`UPDATE api_keys SET is_active = TRUE WHERE user_id = ?`, refereeID)
		if err != nil {
			return nil, err
		}
	}
	
	// 3. Create referral relationship record
	result, err := tx.Exec(
		`INSERT INTO referrals (referrer_id, referee_id, bonus_amount, status, created_at)
		 VALUES (?, ?, ?, 'completed', ?)`,
		referrerID, refereeID, ReferralBonus, now,
	)
	if err != nil {
		return nil, err
	}
	
	referralID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	
	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	
	return &Referral{
		ID:          referralID,
		ReferrerID:  referrerID,
		RefereeID:   refereeID,
		BonusAmount: ReferralBonus,
		Status:      "completed",
		CreatedAt:   now,
	}, nil
}


// GetReferralStats returns referral statistics for a user
// Returns total referrals count and bonus earned
// Requirements: 7.1, 7.2
func GetReferralStats(userID int64) (*ReferralStats, error) {
	stats := &ReferralStats{}
	
	err := db.QueryRow(
		`SELECT COUNT(*), COALESCE(SUM(bonus_amount), 0)
		 FROM referrals WHERE referrer_id = ?`,
		userID,
	).Scan(&stats.TotalReferrals, &stats.TotalBonus)
	
	if err != nil {
		return nil, err
	}
	
	return stats, nil
}


// GetReferralList returns a list of referred users with registration dates
// Requirements: 7.3
func GetReferralList(userID int64, limit, offset int) ([]*ReferredUser, int, error) {
	// Get total count
	var total int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM referrals WHERE referrer_id = ?`,
		userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	// Get referred users with their info
	rows, err := db.Query(
		`SELECT r.referee_id, u.username, u.email, r.created_at, r.bonus_amount
		 FROM referrals r
		 JOIN users u ON r.referee_id = u.id
		 WHERE r.referrer_id = ?
		 ORDER BY r.created_at DESC
		 LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var referredUsers []*ReferredUser
	for rows.Next() {
		user := &ReferredUser{}
		err := rows.Scan(&user.UserID, &user.Username, &user.Email, &user.RegisteredAt, &user.BonusAmount)
		if err != nil {
			return nil, 0, err
		}
		referredUsers = append(referredUsers, user)
	}
	
	return referredUsers, total, nil
}


// GetAllUserBalances retrieves all user balances with pagination
// Used by admin to view all users' balance information
func GetAllUserBalances(limit, offset int) ([]*UserBalance, int, error) {
	// Get total count
	var total int
	err := db.QueryRow(`SELECT COUNT(*) FROM user_balances`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get balances with user info
	rows, err := db.Query(
		`SELECT ub.id, ub.user_id, ub.balance, ub.status, ub.referral_code, 
		        ub.total_consumed, ub.total_recharged, ub.created_at, ub.updated_at
		 FROM user_balances ub
		 ORDER BY ub.created_at DESC
		 LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var balances []*UserBalance
	for rows.Next() {
		balance := &UserBalance{}
		err := rows.Scan(&balance.ID, &balance.UserID, &balance.Balance, &balance.Status,
			&balance.ReferralCode, &balance.TotalConsumed, &balance.TotalRecharged,
			&balance.CreatedAt, &balance.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		balances = append(balances, balance)
	}

	return balances, total, nil
}

// UserBalanceWithInfo represents a user balance with additional user info
type UserBalanceWithInfo struct {
	UserBalance
	Username string `json:"username"`
	Email    string `json:"email"`
}

// GetAllUserBalancesWithInfo retrieves all user balances with user info and pagination
// Used by admin to view all users' balance information with usernames
func GetAllUserBalancesWithInfo(limit, offset int) ([]*UserBalanceWithInfo, int, error) {
	// Get total count
	var total int
	err := db.QueryRow(`SELECT COUNT(*) FROM user_balances`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get balances with user info
	rows, err := db.Query(
		`SELECT ub.id, ub.user_id, ub.balance, ub.status, ub.referral_code, 
		        ub.total_consumed, ub.total_recharged, ub.created_at, ub.updated_at,
		        u.username, u.email
		 FROM user_balances ub
		 JOIN users u ON ub.user_id = u.id
		 ORDER BY ub.created_at DESC
		 LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var balances []*UserBalanceWithInfo
	for rows.Next() {
		balance := &UserBalanceWithInfo{}
		err := rows.Scan(&balance.ID, &balance.UserID, &balance.Balance, &balance.Status,
			&balance.ReferralCode, &balance.TotalConsumed, &balance.TotalRecharged,
			&balance.CreatedAt, &balance.UpdatedAt, &balance.Username, &balance.Email)
		if err != nil {
			return nil, 0, err
		}
		balances = append(balances, balance)
	}

	return balances, total, nil
}
