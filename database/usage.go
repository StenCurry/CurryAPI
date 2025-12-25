package database

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"github.com/sirupsen/logrus"
)

// UsageRecord represents a single API usage record
type UsageRecord struct {
	ID               int64     `db:"id"`
	UserID           int64     `db:"user_id"`
	Username         string    `db:"username"`
	APIToken         string    `db:"api_token"`
	TokenName        string    `db:"token_name"`
	Model            string    `db:"model"`
	PromptTokens     int       `db:"prompt_tokens"`
	CompletionTokens int       `db:"completion_tokens"`
	TotalTokens      int       `db:"total_tokens"`
	CursorSession    string    `db:"cursor_session"`
	StatusCode       int       `db:"status_code"`
	ErrorMessage     string    `db:"error_message"`
	RequestTime      time.Time `db:"request_time"`
	ResponseTime     time.Time `db:"response_time"`
	DurationMs       int       `db:"duration_ms"`
	CreatedAt        time.Time `db:"created_at"`
}

// UsageFilter represents filtering options for usage queries
type UsageFilter struct {
	StartDate *time.Time
	EndDate   *time.Time
	UserID    *int64
	Model     *string
	Limit     int
	Offset    int
}

// UsageStats represents aggregated usage statistics
type UsageStats struct {
	TotalRequests    int
	TotalTokens      int64
	PromptTokens     int64
	CompletionTokens int64
	ByModel          map[string]ModelStats
	RecentCalls      []UsageRecord
	DailyUsage       []DailyStats
}

// ModelStats represents usage statistics for a specific model
type ModelStats struct {
	Model            string
	RequestCount     int
	TotalTokens      int64
	PromptTokens     int64
	CompletionTokens int64
}

// DailyStats represents usage statistics for a specific day
type DailyStats struct {
	Date             time.Time
	Requests         int
	TotalTokens      int64
	PromptTokens     int64
	CompletionTokens int64
}

// AggregateStats represents system-wide usage statistics
type AggregateStats struct {
	TotalUsers    int
	TotalRequests int
	TotalTokens   int64
	TopUsers      []UserUsageSummary
	TopModels     []ModelStats
	UsageTrends   []DailyStats
}

// UserUsageSummary represents a summary of a user's usage
type UserUsageSummary struct {
	UserID      int64
	Username    string
	Requests    int
	TotalTokens int64
}

// InsertUsageRecord inserts a single usage record into the database
func InsertUsageRecord(record *UsageRecord) error {
	dbConn, err := GetDB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `
		INSERT INTO usage_records (
			user_id, username, api_token, token_name, model,
			prompt_tokens, completion_tokens, total_tokens,
			cursor_session, status_code, error_message,
			request_time, response_time, duration_ms
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := dbConn.Exec(query,
		record.UserID,
		record.Username,
		record.APIToken,
		record.TokenName,
		record.Model,
		record.PromptTokens,
		record.CompletionTokens,
		record.TotalTokens,
		record.CursorSession,
		record.StatusCode,
		record.ErrorMessage,
		record.RequestTime,
		record.ResponseTime,
		record.DurationMs,
	)

	if err != nil {
		return fmt.Errorf("failed to insert usage record: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		logrus.Warnf("Failed to get last insert ID: %v", err)
	} else {
		record.ID = id
	}

	return nil
}

// BatchInsertUsageRecords inserts multiple usage records in a single transaction
func BatchInsertUsageRecords(records []*UsageRecord) error {
	if len(records) == 0 {
		return nil
	}

	dbConn, err := GetDB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Start transaction
	tx, err := dbConn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	query := `
		INSERT INTO usage_records (
			user_id, username, api_token, token_name, model,
			prompt_tokens, completion_tokens, total_tokens,
			cursor_session, status_code, error_message,
			request_time, response_time, duration_ms
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, record := range records {
		_, err := stmt.Exec(
			record.UserID,
			record.Username,
			record.APIToken,
			record.TokenName,
			record.Model,
			record.PromptTokens,
			record.CompletionTokens,
			record.TotalTokens,
			record.CursorSession,
			record.StatusCode,
			record.ErrorMessage,
			record.RequestTime,
			record.ResponseTime,
			record.DurationMs,
		)
		if err != nil {
			return fmt.Errorf("failed to insert record in batch: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	logrus.Infof("Successfully inserted %d usage records in batch", len(records))
	return nil
}

// GetUsageRecordsByUser retrieves usage records for a specific user with optional filtering
func GetUsageRecordsByUser(userID int64, filter UsageFilter) ([]*UsageRecord, error) {
	dbConn, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `
		SELECT id, user_id, username, api_token, token_name, model,
			   prompt_tokens, completion_tokens, total_tokens,
			   cursor_session, status_code, error_message,
			   request_time, response_time, duration_ms, created_at
		FROM usage_records
		WHERE user_id = ?
	`
	args := []interface{}{userID}

	// Apply filters
	if filter.StartDate != nil {
		query += " AND request_time >= ?"
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil {
		query += " AND request_time <= ?"
		args = append(args, *filter.EndDate)
	}
	if filter.Model != nil {
		query += " AND model = ?"
		args = append(args, *filter.Model)
	}

	query += " ORDER BY request_time DESC"

	// Apply pagination
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query usage records: %w", err)
	}
	defer rows.Close()

	var records []*UsageRecord
	for rows.Next() {
		record := &UsageRecord{}
		err := rows.Scan(
			&record.ID,
			&record.UserID,
			&record.Username,
			&record.APIToken,
			&record.TokenName,
			&record.Model,
			&record.PromptTokens,
			&record.CompletionTokens,
			&record.TotalTokens,
			&record.CursorSession,
			&record.StatusCode,
			&record.ErrorMessage,
			&record.RequestTime,
			&record.ResponseTime,
			&record.DurationMs,
			&record.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage record: %w", err)
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating usage records: %w", err)
	}

	return records, nil
}

// GetUsageRecordsByToken retrieves usage records for a specific API token with optional filtering
func GetUsageRecordsByToken(token string, filter UsageFilter) ([]*UsageRecord, error) {
	dbConn, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `
		SELECT id, user_id, username, api_token, token_name, model,
			   prompt_tokens, completion_tokens, total_tokens,
			   cursor_session, status_code, error_message,
			   request_time, response_time, duration_ms, created_at
		FROM usage_records
		WHERE api_token = ?
	`
	args := []interface{}{token}

	// Apply filters
	if filter.StartDate != nil {
		query += " AND request_time >= ?"
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil {
		query += " AND request_time <= ?"
		args = append(args, *filter.EndDate)
	}
	if filter.Model != nil {
		query += " AND model = ?"
		args = append(args, *filter.Model)
	}

	query += " ORDER BY request_time DESC"

	// Apply pagination
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query usage records: %w", err)
	}
	defer rows.Close()

	var records []*UsageRecord
	for rows.Next() {
		record := &UsageRecord{}
		err := rows.Scan(
			&record.ID,
			&record.UserID,
			&record.Username,
			&record.APIToken,
			&record.TokenName,
			&record.Model,
			&record.PromptTokens,
			&record.CompletionTokens,
			&record.TotalTokens,
			&record.CursorSession,
			&record.StatusCode,
			&record.ErrorMessage,
			&record.RequestTime,
			&record.ResponseTime,
			&record.DurationMs,
			&record.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage record: %w", err)
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating usage records: %w", err)
	}

	return records, nil
}

// GetUsageRecordsByDateRange retrieves usage records within a specific date range
func GetUsageRecordsByDateRange(start, end time.Time) ([]*UsageRecord, error) {
	dbConn, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `
		SELECT id, user_id, username, api_token, token_name, model,
			   prompt_tokens, completion_tokens, total_tokens,
			   cursor_session, status_code, error_message,
			   request_time, response_time, duration_ms, created_at
		FROM usage_records
		WHERE request_time >= ? AND request_time <= ?
		ORDER BY request_time DESC
	`

	rows, err := dbConn.Query(query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query usage records: %w", err)
	}
	defer rows.Close()

	var records []*UsageRecord
	for rows.Next() {
		record := &UsageRecord{}
		err := rows.Scan(
			&record.ID,
			&record.UserID,
			&record.Username,
			&record.APIToken,
			&record.TokenName,
			&record.Model,
			&record.PromptTokens,
			&record.CompletionTokens,
			&record.TotalTokens,
			&record.CursorSession,
			&record.StatusCode,
			&record.ErrorMessage,
			&record.RequestTime,
			&record.ResponseTime,
			&record.DurationMs,
			&record.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage record: %w", err)
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating usage records: %w", err)
	}

	return records, nil
}

// GetUserUsageStats retrieves aggregated usage statistics for a specific user
func GetUserUsageStats(userID int64, filter UsageFilter) (*UsageStats, error) {
	dbConn, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	stats := &UsageStats{
		ByModel: make(map[string]ModelStats),
	}

	// Build base query with filters
	query := `
		SELECT 
			COUNT(*) as total_requests,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(prompt_tokens), 0) as prompt_tokens,
			COALESCE(SUM(completion_tokens), 0) as completion_tokens
		FROM usage_records
		WHERE user_id = ?
	`
	args := []interface{}{userID}

	if filter.StartDate != nil {
		query += " AND request_time >= ?"
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil {
		query += " AND request_time <= ?"
		args = append(args, *filter.EndDate)
	}

	// Get overall stats
	err = dbConn.QueryRow(query, args...).Scan(
		&stats.TotalRequests,
		&stats.TotalTokens,
		&stats.PromptTokens,
		&stats.CompletionTokens,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get user usage stats: %w", err)
	}

	// Get breakdown by model
	modelQuery := `
		SELECT 
			model,
			COUNT(*) as request_count,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(prompt_tokens), 0) as prompt_tokens,
			COALESCE(SUM(completion_tokens), 0) as completion_tokens
		FROM usage_records
		WHERE user_id = ?
	`
	modelArgs := []interface{}{userID}

	if filter.StartDate != nil {
		modelQuery += " AND request_time >= ?"
		modelArgs = append(modelArgs, *filter.StartDate)
	}
	if filter.EndDate != nil {
		modelQuery += " AND request_time <= ?"
		modelArgs = append(modelArgs, *filter.EndDate)
	}

	modelQuery += " GROUP BY model"

	rows, err := dbConn.Query(modelQuery, modelArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get model breakdown: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var modelStats ModelStats
		err := rows.Scan(
			&modelStats.Model,
			&modelStats.RequestCount,
			&modelStats.TotalTokens,
			&modelStats.PromptTokens,
			&modelStats.CompletionTokens,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan model stats: %w", err)
		}
		stats.ByModel[modelStats.Model] = modelStats
	}

	// Get recent calls
	limit := 50
	if filter.Limit > 0 {
		limit = filter.Limit
	}
	recentFilter := UsageFilter{
		StartDate: filter.StartDate,
		EndDate:   filter.EndDate,
		Limit:     limit,
	}
	recentCalls, err := GetUsageRecordsByUser(userID, recentFilter)
	if err != nil {
		logrus.Warnf("Failed to get recent calls: %v", err)
	} else {
		for _, record := range recentCalls {
			stats.RecentCalls = append(stats.RecentCalls, *record)
		}
	}

	return stats, nil
}

// GetAllUsageStats retrieves system-wide aggregated usage statistics
func GetAllUsageStats(filter UsageFilter) (*AggregateStats, error) {
	dbConn, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	stats := &AggregateStats{}

	// Build base query with filters
	query := `
		SELECT 
			COUNT(DISTINCT user_id) as total_users,
			COUNT(*) as total_requests,
			COALESCE(SUM(total_tokens), 0) as total_tokens
		FROM usage_records
		WHERE 1=1
	`
	args := []interface{}{}

	if filter.StartDate != nil {
		query += " AND request_time >= ?"
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil {
		query += " AND request_time <= ?"
		args = append(args, *filter.EndDate)
	}

	// Get overall stats
	err = dbConn.QueryRow(query, args...).Scan(
		&stats.TotalUsers,
		&stats.TotalRequests,
		&stats.TotalTokens,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get aggregate stats: %w", err)
	}

	// Get top users
	topUsersQuery := `
		SELECT 
			user_id,
			username,
			COUNT(*) as requests,
			COALESCE(SUM(total_tokens), 0) as total_tokens
		FROM usage_records
		WHERE 1=1
	`
	topUsersArgs := []interface{}{}

	if filter.StartDate != nil {
		topUsersQuery += " AND request_time >= ?"
		topUsersArgs = append(topUsersArgs, *filter.StartDate)
	}
	if filter.EndDate != nil {
		topUsersQuery += " AND request_time <= ?"
		topUsersArgs = append(topUsersArgs, *filter.EndDate)
	}

	topUsersQuery += " GROUP BY user_id, username ORDER BY total_tokens DESC LIMIT 10"

	rows, err := dbConn.Query(topUsersQuery, topUsersArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get top users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var userSummary UserUsageSummary
		err := rows.Scan(
			&userSummary.UserID,
			&userSummary.Username,
			&userSummary.Requests,
			&userSummary.TotalTokens,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user summary: %w", err)
		}
		stats.TopUsers = append(stats.TopUsers, userSummary)
	}

	// Get top models
	topModelsQuery := `
		SELECT 
			model,
			COUNT(*) as request_count,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(prompt_tokens), 0) as prompt_tokens,
			COALESCE(SUM(completion_tokens), 0) as completion_tokens
		FROM usage_records
		WHERE 1=1
	`
	topModelsArgs := []interface{}{}

	if filter.StartDate != nil {
		topModelsQuery += " AND request_time >= ?"
		topModelsArgs = append(topModelsArgs, *filter.StartDate)
	}
	if filter.EndDate != nil {
		topModelsQuery += " AND request_time <= ?"
		topModelsArgs = append(topModelsArgs, *filter.EndDate)
	}

	topModelsQuery += " GROUP BY model ORDER BY request_count DESC"

	rows, err = dbConn.Query(topModelsQuery, topModelsArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get top models: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var modelStats ModelStats
		err := rows.Scan(
			&modelStats.Model,
			&modelStats.RequestCount,
			&modelStats.TotalTokens,
			&modelStats.PromptTokens,
			&modelStats.CompletionTokens,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan model stats: %w", err)
		}
		stats.TopModels = append(stats.TopModels, modelStats)
	}

	return stats, nil
}

// GetModelUsageBreakdown retrieves usage breakdown by model
func GetModelUsageBreakdown(userID *int64, filter UsageFilter) (map[string]ModelStats, error) {
	dbConn, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `
		SELECT 
			model,
			COUNT(*) as request_count,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(prompt_tokens), 0) as prompt_tokens,
			COALESCE(SUM(completion_tokens), 0) as completion_tokens
		FROM usage_records
		WHERE 1=1
	`
	args := []interface{}{}

	if userID != nil {
		query += " AND user_id = ?"
		args = append(args, *userID)
	}
	if filter.StartDate != nil {
		query += " AND request_time >= ?"
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil {
		query += " AND request_time <= ?"
		args = append(args, *filter.EndDate)
	}

	query += " GROUP BY model"

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get model breakdown: %w", err)
	}
	defer rows.Close()

	breakdown := make(map[string]ModelStats)
	for rows.Next() {
		var modelStats ModelStats
		err := rows.Scan(
			&modelStats.Model,
			&modelStats.RequestCount,
			&modelStats.TotalTokens,
			&modelStats.PromptTokens,
			&modelStats.CompletionTokens,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan model stats: %w", err)
		}
		breakdown[modelStats.Model] = modelStats
	}

	return breakdown, nil
}

// GetDailyUsageTrends retrieves daily usage trends for the specified number of days
func GetDailyUsageTrends(userID *int64, days int) ([]DailyStats, error) {
	dbConn, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `
		SELECT 
			DATE(request_time) as date,
			COUNT(*) as requests,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(prompt_tokens), 0) as prompt_tokens,
			COALESCE(SUM(completion_tokens), 0) as completion_tokens
		FROM usage_records
		WHERE request_time >= DATE_SUB(NOW(), INTERVAL ? DAY)
	`
	args := []interface{}{days}

	if userID != nil {
		query += " AND user_id = ?"
		args = append(args, *userID)
	}

	query += " GROUP BY DATE(request_time) ORDER BY date ASC"

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily trends: %w", err)
	}
	defer rows.Close()

	var trends []DailyStats
	for rows.Next() {
		var stats DailyStats
		err := rows.Scan(
			&stats.Date,
			&stats.Requests,
			&stats.TotalTokens,
			&stats.PromptTokens,
			&stats.CompletionTokens,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan daily stats: %w", err)
		}
		trends = append(trends, stats)
	}

	return trends, nil
}

// CursorSessionStats represents usage statistics for a Cursor session
type CursorSessionStats struct {
	CursorSession string
	Requests      int
	TotalTokens   int64
}

// GetCursorSessionUsage retrieves usage statistics grouped by Cursor session
func GetCursorSessionUsage(filter UsageFilter) ([]CursorSessionStats, error) {
	dbConn, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `
		SELECT 
			cursor_session,
			COUNT(*) as requests,
			COALESCE(SUM(total_tokens), 0) as total_tokens
		FROM usage_records
		WHERE cursor_session IS NOT NULL AND cursor_session != ''
	`
	args := []interface{}{}

	if filter.StartDate != nil {
		query += " AND request_time >= ?"
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil {
		query += " AND request_time <= ?"
		args = append(args, *filter.EndDate)
	}

	query += " GROUP BY cursor_session ORDER BY requests DESC"

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get cursor session usage: %w", err)
	}
	defer rows.Close()

	var sessions []CursorSessionStats
	for rows.Next() {
		var stats CursorSessionStats
		err := rows.Scan(
			&stats.CursorSession,
			&stats.Requests,
			&stats.TotalTokens,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cursor session stats: %w", err)
		}
		sessions = append(sessions, stats)
	}

	return sessions, nil
}

// StreamUsageRecordsCSV streams usage records as CSV directly to the writer
// This function processes records in chunks to avoid loading all data into memory
func StreamUsageRecordsCSV(writer io.Writer, filter UsageFilter) error {
	dbConn, err := GetDB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Create CSV writer
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write CSV header
	header := []string{
		"ID",
		"User ID",
		"Username",
		"API Token",
		"Token Name",
		"Model",
		"Prompt Tokens",
		"Completion Tokens",
		"Total Tokens",
		"Cursor Session",
		"Status Code",
		"Error Message",
		"Request Time",
		"Response Time",
		"Duration (ms)",
		"Created At",
	}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Build query with filters
	query := `
		SELECT id, user_id, username, api_token, token_name, model,
			   prompt_tokens, completion_tokens, total_tokens,
			   cursor_session, status_code, error_message,
			   request_time, response_time, duration_ms, created_at
		FROM usage_records
		WHERE 1=1
	`
	args := []interface{}{}

	if filter.UserID != nil {
		query += " AND user_id = ?"
		args = append(args, *filter.UserID)
	}
	if filter.StartDate != nil {
		query += " AND request_time >= ?"
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil {
		query += " AND request_time <= ?"
		args = append(args, *filter.EndDate)
	}
	if filter.Model != nil {
		query += " AND model = ?"
		args = append(args, *filter.Model)
	}

	query += " ORDER BY request_time DESC"

	// Execute query
	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return fmt.Errorf("failed to query usage records: %w", err)
	}
	defer rows.Close()

	// Process records in chunks to avoid memory issues
	const chunkSize = 1000
	recordCount := 0
	rowBuffer := make([][]string, 0, chunkSize)

	for rows.Next() {
		var record UsageRecord
		err := rows.Scan(
			&record.ID,
			&record.UserID,
			&record.Username,
			&record.APIToken,
			&record.TokenName,
			&record.Model,
			&record.PromptTokens,
			&record.CompletionTokens,
			&record.TotalTokens,
			&record.CursorSession,
			&record.StatusCode,
			&record.ErrorMessage,
			&record.RequestTime,
			&record.ResponseTime,
			&record.DurationMs,
			&record.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to scan usage record: %w", err)
		}

		// Convert record to CSV row
		row := []string{
			fmt.Sprintf("%d", record.ID),
			fmt.Sprintf("%d", record.UserID),
			record.Username,
			record.APIToken,
			record.TokenName,
			record.Model,
			fmt.Sprintf("%d", record.PromptTokens),
			fmt.Sprintf("%d", record.CompletionTokens),
			fmt.Sprintf("%d", record.TotalTokens),
			record.CursorSession,
			fmt.Sprintf("%d", record.StatusCode),
			record.ErrorMessage,
			record.RequestTime.Format(time.RFC3339),
			record.ResponseTime.Format(time.RFC3339),
			fmt.Sprintf("%d", record.DurationMs),
			record.CreatedAt.Format(time.RFC3339),
		}

		rowBuffer = append(rowBuffer, row)
		recordCount++

		// Write chunk when buffer is full
		if len(rowBuffer) >= chunkSize {
			if err := csvWriter.WriteAll(rowBuffer); err != nil {
				return fmt.Errorf("failed to write CSV chunk: %w", err)
			}
			csvWriter.Flush()
			if err := csvWriter.Error(); err != nil {
				return fmt.Errorf("CSV writer error: %w", err)
			}
			rowBuffer = rowBuffer[:0] // Clear buffer
		}
	}

	// Write remaining records
	if len(rowBuffer) > 0 {
		if err := csvWriter.WriteAll(rowBuffer); err != nil {
			return fmt.Errorf("failed to write final CSV chunk: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating usage records: %w", err)
	}

	logrus.Infof("Successfully exported %d usage records to CSV", recordCount)
	return nil
}

// AggregateUsageStats represents preserved aggregate statistics
type AggregateUsageStats struct {
	ID               int64     `db:"id"`
	PeriodType       string    `db:"period_type"` // daily, weekly, monthly
	PeriodStart      time.Time `db:"period_start"`
	PeriodEnd        time.Time `db:"period_end"`
	UserID           *int64    `db:"user_id"` // NULL for system-wide aggregates
	Model            *string   `db:"model"`   // NULL for all-model aggregates
	TotalRequests    int       `db:"total_requests"`
	TotalTokens      int64     `db:"total_tokens"`
	PromptTokens     int64     `db:"prompt_tokens"`
	CompletionTokens int64     `db:"completion_tokens"`
	CreatedAt        time.Time `db:"created_at"`
}

// DeleteOldUsageRecords deletes usage records older than the cutoff date in batches
// Returns the total number of records deleted
func DeleteOldUsageRecords(cutoffDate time.Time, batchSize int) (int64, error) {
	dbConn, err := GetDB()
	if err != nil {
		return 0, fmt.Errorf("failed to get database connection: %w", err)
	}

	var totalDeleted int64

	// Delete in batches to avoid locking the table for too long
	for {
		query := `
			DELETE FROM usage_records 
			WHERE request_time < ? 
			LIMIT ?
		`

		result, err := dbConn.Exec(query, cutoffDate, batchSize)
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to delete batch: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to get rows affected: %w", err)
		}

		totalDeleted += rowsAffected
		logrus.Debugf("Deleted batch of %d records (total: %d)", rowsAffected, totalDeleted)

		// If we deleted fewer than batchSize, we're done
		if rowsAffected < int64(batchSize) {
			break
		}

		// Small delay between batches to reduce database load
		time.Sleep(100 * time.Millisecond)
	}

	return totalDeleted, nil
}

// PreserveUsageAggregates calculates and stores aggregate statistics before deletion
func PreserveUsageAggregates(cutoffDate time.Time) error {
	dbConn, err := GetDB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Ensure the aggregate table exists
	if err := ensureAggregateTableExists(dbConn); err != nil {
		return fmt.Errorf("failed to ensure aggregate table exists: %w", err)
	}

	// Calculate daily aggregates for records that will be deleted
	if err := preserveDailyAggregates(dbConn, cutoffDate); err != nil {
		return fmt.Errorf("failed to preserve daily aggregates: %w", err)
	}

	// Calculate user-level aggregates
	if err := preserveUserAggregates(dbConn, cutoffDate); err != nil {
		return fmt.Errorf("failed to preserve user aggregates: %w", err)
	}

	// Calculate model-level aggregates
	if err := preserveModelAggregates(dbConn, cutoffDate); err != nil {
		return fmt.Errorf("failed to preserve model aggregates: %w", err)
	}

	logrus.Info("Successfully preserved usage aggregates")
	return nil
}

// ensureAggregateTableExists creates the aggregate_usage_stats table if it doesn't exist
func ensureAggregateTableExists(dbConn *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS aggregate_usage_stats (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			period_type VARCHAR(20) NOT NULL COMMENT 'daily, weekly, monthly, user, model',
			period_start DATETIME NOT NULL,
			period_end DATETIME NOT NULL,
			user_id INT NULL COMMENT 'NULL for system-wide aggregates',
			model VARCHAR(100) NULL COMMENT 'NULL for all-model aggregates',
			total_requests INT NOT NULL DEFAULT 0,
			total_tokens BIGINT NOT NULL DEFAULT 0,
			prompt_tokens BIGINT NOT NULL DEFAULT 0,
			completion_tokens BIGINT NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_period_type (period_type, period_start),
			INDEX idx_user_period (user_id, period_type, period_start),
			INDEX idx_model_period (model, period_type, period_start),
			UNIQUE KEY uk_aggregate (period_type, period_start, period_end, user_id, model)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`

	_, err := dbConn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create aggregate table: %w", err)
	}

	return nil
}

// preserveDailyAggregates preserves daily system-wide aggregates
func preserveDailyAggregates(dbConn *sql.DB, cutoffDate time.Time) error {
	query := `
		INSERT INTO aggregate_usage_stats 
			(period_type, period_start, period_end, user_id, model, total_requests, total_tokens, prompt_tokens, completion_tokens)
		SELECT 
			'daily' as period_type,
			DATE(request_time) as period_start,
			DATE_ADD(DATE(request_time), INTERVAL 1 DAY) as period_end,
			NULL as user_id,
			NULL as model,
			COUNT(*) as total_requests,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(prompt_tokens), 0) as prompt_tokens,
			COALESCE(SUM(completion_tokens), 0) as completion_tokens
		FROM usage_records
		WHERE request_time < ?
		GROUP BY DATE(request_time)
		ON DUPLICATE KEY UPDATE
			total_requests = VALUES(total_requests),
			total_tokens = VALUES(total_tokens),
			prompt_tokens = VALUES(prompt_tokens),
			completion_tokens = VALUES(completion_tokens)
	`

	result, err := dbConn.Exec(query, cutoffDate)
	if err != nil {
		return fmt.Errorf("failed to insert daily aggregates: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	logrus.Debugf("Preserved %d daily aggregate records", rowsAffected)
	return nil
}

// preserveUserAggregates preserves per-user aggregates
func preserveUserAggregates(dbConn *sql.DB, cutoffDate time.Time) error {
	query := `
		INSERT INTO aggregate_usage_stats 
			(period_type, period_start, period_end, user_id, model, total_requests, total_tokens, prompt_tokens, completion_tokens)
		SELECT 
			'user' as period_type,
			DATE(MIN(request_time)) as period_start,
			? as period_end,
			user_id,
			NULL as model,
			COUNT(*) as total_requests,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(prompt_tokens), 0) as prompt_tokens,
			COALESCE(SUM(completion_tokens), 0) as completion_tokens
		FROM usage_records
		WHERE request_time < ?
		GROUP BY user_id
		ON DUPLICATE KEY UPDATE
			total_requests = aggregate_usage_stats.total_requests + VALUES(total_requests),
			total_tokens = aggregate_usage_stats.total_tokens + VALUES(total_tokens),
			prompt_tokens = aggregate_usage_stats.prompt_tokens + VALUES(prompt_tokens),
			completion_tokens = aggregate_usage_stats.completion_tokens + VALUES(completion_tokens)
	`

	result, err := dbConn.Exec(query, cutoffDate, cutoffDate)
	if err != nil {
		return fmt.Errorf("failed to insert user aggregates: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	logrus.Debugf("Preserved %d user aggregate records", rowsAffected)
	return nil
}

// preserveModelAggregates preserves per-model aggregates
func preserveModelAggregates(dbConn *sql.DB, cutoffDate time.Time) error {
	query := `
		INSERT INTO aggregate_usage_stats 
			(period_type, period_start, period_end, user_id, model, total_requests, total_tokens, prompt_tokens, completion_tokens)
		SELECT 
			'model' as period_type,
			DATE(MIN(request_time)) as period_start,
			? as period_end,
			NULL as user_id,
			model,
			COUNT(*) as total_requests,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(prompt_tokens), 0) as prompt_tokens,
			COALESCE(SUM(completion_tokens), 0) as completion_tokens
		FROM usage_records
		WHERE request_time < ?
		GROUP BY model
		ON DUPLICATE KEY UPDATE
			total_requests = aggregate_usage_stats.total_requests + VALUES(total_requests),
			total_tokens = aggregate_usage_stats.total_tokens + VALUES(total_tokens),
			prompt_tokens = aggregate_usage_stats.prompt_tokens + VALUES(prompt_tokens),
			completion_tokens = aggregate_usage_stats.completion_tokens + VALUES(completion_tokens)
	`

	result, err := dbConn.Exec(query, cutoffDate, cutoffDate)
	if err != nil {
		return fmt.Errorf("failed to insert model aggregates: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	logrus.Debugf("Preserved %d model aggregate records", rowsAffected)
	return nil
}

// GetAggregateStats retrieves preserved aggregate statistics
func GetAggregateStats(periodType string, startDate, endDate *time.Time) ([]AggregateUsageStats, error) {
	dbConn, err := GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `
		SELECT id, period_type, period_start, period_end, user_id, model,
			   total_requests, total_tokens, prompt_tokens, completion_tokens, created_at
		FROM aggregate_usage_stats
		WHERE period_type = ?
	`
	args := []interface{}{periodType}

	if startDate != nil {
		query += " AND period_start >= ?"
		args = append(args, *startDate)
	}
	if endDate != nil {
		query += " AND period_end <= ?"
		args = append(args, *endDate)
	}

	query += " ORDER BY period_start ASC"

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query aggregate stats: %w", err)
	}
	defer rows.Close()

	var stats []AggregateUsageStats
	for rows.Next() {
		var s AggregateUsageStats
		err := rows.Scan(
			&s.ID,
			&s.PeriodType,
			&s.PeriodStart,
			&s.PeriodEnd,
			&s.UserID,
			&s.Model,
			&s.TotalRequests,
			&s.TotalTokens,
			&s.PromptTokens,
			&s.CompletionTokens,
			&s.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan aggregate stats: %w", err)
		}
		stats = append(stats, s)
	}

	return stats, nil
}

// CountUsageRecordsOlderThan counts records older than the specified date
func CountUsageRecordsOlderThan(cutoffDate time.Time) (int64, error) {
	dbConn, err := GetDB()
	if err != nil {
		return 0, fmt.Errorf("failed to get database connection: %w", err)
	}

	var count int64
	query := `SELECT COUNT(*) FROM usage_records WHERE request_time < ?`
	err = dbConn.QueryRow(query, cutoffDate).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count old records: %w", err)
	}

	return count, nil
}
