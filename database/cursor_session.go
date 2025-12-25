package database

import (
	"Curry2API-go/models"
	"Curry2API-go/utils"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	ErrCursorSessionNotFound = errors.New("cursor session not found")
	ErrCursorSessionExists   = errors.New("cursor session already exists")
)

// sanitizeEmail 清理邮箱中的空白字符（回车、换行、空格等）
func sanitizeEmail(email string) string {
	// 移除所有空白字符（包括 \r, \n, \t, 空格等）
	email = strings.TrimSpace(email)
	email = strings.ReplaceAll(email, "\r", "")
	email = strings.ReplaceAll(email, "\n", "")
	email = strings.ReplaceAll(email, "\t", "")
	return email
}

// AddCursorSession 添加Cursor Session
func AddCursorSession(email, token, userAgent string, expiresAt time.Time, extraCookies map[string]string) error {
	// 清理邮箱中的空白字符
	email = sanitizeEmail(email)
	
	// 加密 token
	encryptedToken, err := utils.EncryptSensitiveData(token)
	if err != nil {
		logrus.WithError(err).Warn("Failed to encrypt cursor token, storing as plaintext")
		encryptedToken = token
	}
	
	// 序列化并加密 extra_cookies
	extraCookiesJSON, err := json.Marshal(extraCookies)
	if err != nil {
		return err
	}
	encryptedCookies, err := utils.EncryptSensitiveData(string(extraCookiesJSON))
	if err != nil {
		logrus.WithError(err).Warn("Failed to encrypt extra cookies, storing as plaintext")
		encryptedCookies = string(extraCookiesJSON)
	}
	
	now := time.Now()
	// Default quota: 100,000 tokens for free accounts
	defaultQuota := int64(100000)
	
	_, err = db.Exec(
		`INSERT INTO cursor_sessions 
		 (email, token, user_agent, extra_cookies, created_at, last_used, last_check, expires_at, is_valid, usage_count, fail_count,
		  daily_token_limit, daily_token_used, last_reset_date, quota_status, account_type) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		email, encryptedToken, userAgent, encryptedCookies, now, now, now, expiresAt, true, 0, 0,
		defaultQuota, 0, now, "available", "free",
	)
	return err
}

// GetCursorSession 获取Cursor Session
func GetCursorSession(email string) (*models.CursorSessionInfo, error) {
	email = sanitizeEmail(email)
	session := &models.CursorSessionInfo{}
	var userAgent sql.NullString
	var extraCookiesJSON sql.NullString
	var encryptedToken string
	var lastUsed sql.NullTime
	var lastCheck sql.NullTime
	var expiresAt sql.NullTime
	var lastResetDate sql.NullTime
	var quotaStatus sql.NullString
	var accountType sql.NullString
	
	err := db.QueryRow(
		`SELECT email, token, user_agent, extra_cookies, created_at, last_used, last_check, expires_at, is_valid, usage_count, fail_count,
		 daily_token_limit, daily_token_used, last_reset_date, quota_status, account_type
		 FROM cursor_sessions WHERE email = ?`,
		email,
	).Scan(&session.Email, &encryptedToken, &userAgent, &extraCookiesJSON, 
		&session.CreatedAt, &lastUsed, &lastCheck, &expiresAt, 
		&session.IsValid, &session.UsageCount, &session.FailCount,
		&session.DailyTokenLimit, &session.DailyTokenUsed, &lastResetDate,
		&quotaStatus, &accountType)
	
	if err == sql.ErrNoRows {
		return nil, ErrCursorSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	
	// 解密 token
	decryptedToken, err := utils.DecryptSensitiveData(encryptedToken)
	if err != nil {
		logrus.WithError(err).Warn("Failed to decrypt cursor token")
		session.Token = encryptedToken // 回退到原始值
	} else {
		session.Token = decryptedToken
	}
	
	// 处理可能为 NULL 的字段
	if userAgent.Valid {
		session.UserAgent = userAgent.String
	}
	if lastUsed.Valid {
		session.LastUsed = lastUsed.Time
	}
	if lastCheck.Valid {
		session.LastCheck = lastCheck.Time
	}
	if expiresAt.Valid {
		session.ExpiresAt = expiresAt.Time
	}
	if lastResetDate.Valid {
		session.LastResetDate = lastResetDate.Time
	}
	if quotaStatus.Valid {
		session.QuotaStatus = quotaStatus.String
	}
	if accountType.Valid {
		session.AccountType = accountType.String
	}
	
	// 解密并反序列化 extra_cookies
	if extraCookiesJSON.Valid && extraCookiesJSON.String != "" {
		decryptedCookies, err := utils.DecryptSensitiveData(extraCookiesJSON.String)
		if err != nil {
			logrus.WithError(err).Warn("Failed to decrypt extra cookies")
			decryptedCookies = extraCookiesJSON.String
		}
		if err := json.Unmarshal([]byte(decryptedCookies), &session.ExtraCookies); err != nil {
			return nil, err
		}
	}
	
	return session, nil
}

// ListCursorSessions 列出所有Cursor Sessions
func ListCursorSessions() ([]*models.CursorSessionInfo, error) {
	rows, err := db.Query(
		`SELECT email, token, user_agent, extra_cookies, created_at, last_used, last_check, expires_at, is_valid, usage_count, fail_count,
		 daily_token_limit, daily_token_used, last_reset_date, quota_status, account_type
		 FROM cursor_sessions ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var sessions []*models.CursorSessionInfo
	for rows.Next() {
		session := &models.CursorSessionInfo{}
		var userAgent sql.NullString
		var extraCookiesJSON sql.NullString
		var encryptedToken string
		var lastUsed sql.NullTime
		var lastCheck sql.NullTime
		var expiresAt sql.NullTime
		var lastResetDate sql.NullTime
		var quotaStatus sql.NullString
		var accountType sql.NullString
		
		err := rows.Scan(&session.Email, &encryptedToken, &userAgent, &extraCookiesJSON, 
			&session.CreatedAt, &lastUsed, &lastCheck, &expiresAt, 
			&session.IsValid, &session.UsageCount, &session.FailCount,
			&session.DailyTokenLimit, &session.DailyTokenUsed, &lastResetDate,
			&quotaStatus, &accountType)
		if err != nil {
			return nil, err
		}
		
		// 解密 token
		decryptedToken, err := utils.DecryptSensitiveData(encryptedToken)
		if err != nil {
			session.Token = encryptedToken // 回退到原始值
		} else {
			session.Token = decryptedToken
		}
		
		// 处理可能为 NULL 的字段
		if userAgent.Valid {
			session.UserAgent = userAgent.String
		}
		if lastUsed.Valid {
			session.LastUsed = lastUsed.Time
		}
		if lastCheck.Valid {
			session.LastCheck = lastCheck.Time
		}
		if expiresAt.Valid {
			session.ExpiresAt = expiresAt.Time
		}
		if lastResetDate.Valid {
			session.LastResetDate = lastResetDate.Time
		}
		if quotaStatus.Valid {
			session.QuotaStatus = quotaStatus.String
		}
		if accountType.Valid {
			session.AccountType = accountType.String
		}
		
		// 解密并反序列化 extra_cookies
		if extraCookiesJSON.Valid && extraCookiesJSON.String != "" {
			decryptedCookies, err := utils.DecryptSensitiveData(extraCookiesJSON.String)
			if err != nil {
				decryptedCookies = extraCookiesJSON.String
			}
			if err := json.Unmarshal([]byte(decryptedCookies), &session.ExtraCookies); err != nil {
				return nil, err
			}
		}
		
		sessions = append(sessions, session)
	}
	
	return sessions, nil
}

// RemoveCursorSession 删除Cursor Session
func RemoveCursorSession(email string) error {
	email = sanitizeEmail(email)
	result, err := db.Exec(`DELETE FROM cursor_sessions WHERE email = ?`, email)
	if err != nil {
		return err
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rows == 0 {
		return ErrCursorSessionNotFound
	}
	
	return nil
}

// UpdateCursorSessionUsage 更新Cursor Session使用信息
func UpdateCursorSessionUsage(email string, success bool) error {
	email = sanitizeEmail(email)
	now := time.Now()
	
	if success {
		_, err := db.Exec(
			`UPDATE cursor_sessions 
			 SET usage_count = usage_count + 1, last_used = ?, fail_count = 0 
			 WHERE email = ?`,
			now, email,
		)
		return err
	} else {
		_, err := db.Exec(
			`UPDATE cursor_sessions 
			 SET fail_count = fail_count + 1, last_check = ? 
			 WHERE email = ?`,
			now, email,
		)
		return err
	}
}

// UpdateCursorSessionValidity 更新Cursor Session有效性
func UpdateCursorSessionValidity(email string, isValid bool) error {
	email = sanitizeEmail(email)
	_, err := db.Exec(
		`UPDATE cursor_sessions SET is_valid = ?, last_check = ? WHERE email = ?`,
		isValid, time.Now(), email,
	)
	return err
}

// GetCursorSessionStats 获取Cursor Session统计信息
func GetCursorSessionStats() (map[string]interface{}, error) {
	var totalSessions, validSessions int
	var totalUsage int64
	
	err := db.QueryRow(
		`SELECT COUNT(*) as total, 
		 SUM(CASE WHEN is_valid = TRUE THEN 1 ELSE 0 END) as valid,
		 SUM(usage_count) as usage 
		 FROM cursor_sessions`,
	).Scan(&totalSessions, &validSessions, &totalUsage)
	
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"total_sessions": totalSessions,
		"valid_sessions": validSessions,
		"total_usage":    totalUsage,
	}, nil
}

// UpdateSessionStatus 更新Cursor Session状态
func UpdateSessionStatus(email string, isValid bool, failCount int) error {
	email = sanitizeEmail(email)
	_, err := db.Exec(
		`UPDATE cursor_sessions SET is_valid = ?, fail_count = ?, last_check = ? WHERE email = ?`,
		isValid, failCount, time.Now(), email,
	)
	return err
}

// UpdateSessionUsage 更新Cursor Session使用次数
func UpdateSessionUsage(email string) error {
	email = sanitizeEmail(email)
	now := time.Now()
	_, err := db.Exec(
		`UPDATE cursor_sessions 
		 SET usage_count = usage_count + 1, last_used = ?, fail_count = 0 
		 WHERE email = ?`,
		now, email,
	)
	return err
}

// UpdateSessionCheck 更新Cursor Session检查时间
func UpdateSessionCheck(email string, lastCheck time.Time, isValid bool) error {
	email = sanitizeEmail(email)
	_, err := db.Exec(
		`UPDATE cursor_sessions SET last_check = ?, is_valid = ? WHERE email = ?`,
		lastCheck, isValid, email,
	)
	return err
}

// UpdateSessionQuota 更新 session 的配额限制
func UpdateSessionQuota(email string, newLimit int64) error {
	email = sanitizeEmail(email)
	_, err := db.Exec(
		`UPDATE cursor_sessions SET daily_token_limit = ? WHERE email = ?`,
		newLimit, email,
	)
	return err
}

// UpdateSessionQuotaUsage 更新 session 的配额使用量
func UpdateSessionQuotaUsage(email string, tokensUsed int64) error {
	email = sanitizeEmail(email)
	_, err := db.Exec(
		`UPDATE cursor_sessions 
		 SET daily_token_used = daily_token_used + ? 
		 WHERE email = ?`,
		tokensUsed, email,
	)
	return err
}

// UpdateSessionQuotaStatus 更新 session 的配额状态
func UpdateSessionQuotaStatus(email string, status string) error {
	email = sanitizeEmail(email)
	_, err := db.Exec(
		`UPDATE cursor_sessions SET quota_status = ? WHERE email = ?`,
		status, email,
	)
	return err
}

// ResetSessionQuota 重置 session 的每日配额
func ResetSessionQuota(email string) error {
	email = sanitizeEmail(email)
	now := time.Now()
	_, err := db.Exec(
		`UPDATE cursor_sessions 
		 SET daily_token_used = 0, 
		     last_reset_date = ?,
		     quota_status = 'available',
		     is_valid = TRUE,
		     fail_count = 0
		 WHERE email = ?`,
		now, email,
	)
	return err
}

// ResetAllSessionQuotas 重置所有 session 的每日配额
func ResetAllSessionQuotas() error {
	now := time.Now()
	_, err := db.Exec(
		`UPDATE cursor_sessions 
		 SET daily_token_used = 0, 
		     last_reset_date = ?,
		     quota_status = 'available',
		     is_valid = TRUE,
		     fail_count = 0`,
		now,
	)
	return err
}

// GetSessionsNeedingReset 获取需要重置配额的 sessions（超过24小时未重置）
func GetSessionsNeedingReset() ([]*models.CursorSessionInfo, error) {
	cutoffTime := time.Now().Add(-24 * time.Hour)
	
	rows, err := db.Query(
		`SELECT email, token, user_agent, extra_cookies, created_at, last_used, last_check, expires_at, is_valid, usage_count, fail_count,
		 daily_token_limit, daily_token_used, last_reset_date, quota_status, account_type
		 FROM cursor_sessions 
		 WHERE last_reset_date < ?
		 ORDER BY last_reset_date ASC`,
		cutoffTime,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var sessions []*models.CursorSessionInfo
	for rows.Next() {
		session := &models.CursorSessionInfo{}
		var userAgent sql.NullString
		var extraCookiesJSON sql.NullString
		var encryptedToken string
		var lastUsed sql.NullTime
		var lastCheck sql.NullTime
		var expiresAt sql.NullTime
		var lastResetDate sql.NullTime
		var quotaStatus sql.NullString
		var accountType sql.NullString
		
		err := rows.Scan(&session.Email, &encryptedToken, &userAgent, &extraCookiesJSON, 
			&session.CreatedAt, &lastUsed, &lastCheck, &expiresAt, 
			&session.IsValid, &session.UsageCount, &session.FailCount,
			&session.DailyTokenLimit, &session.DailyTokenUsed, &lastResetDate,
			&quotaStatus, &accountType)
		if err != nil {
			return nil, err
		}
		
		// 解密 token
		decryptedToken, err := utils.DecryptSensitiveData(encryptedToken)
		if err != nil {
			session.Token = encryptedToken
		} else {
			session.Token = decryptedToken
		}
		
		// 处理可能为 NULL 的字段
		if userAgent.Valid {
			session.UserAgent = userAgent.String
		}
		if lastUsed.Valid {
			session.LastUsed = lastUsed.Time
		}
		if lastCheck.Valid {
			session.LastCheck = lastCheck.Time
		}
		if expiresAt.Valid {
			session.ExpiresAt = expiresAt.Time
		}
		if lastResetDate.Valid {
			session.LastResetDate = lastResetDate.Time
		}
		if quotaStatus.Valid {
			session.QuotaStatus = quotaStatus.String
		}
		if accountType.Valid {
			session.AccountType = accountType.String
		}
		
		// 解密并反序列化 extra_cookies
		if extraCookiesJSON.Valid && extraCookiesJSON.String != "" {
			decryptedCookies, err := utils.DecryptSensitiveData(extraCookiesJSON.String)
			if err != nil {
				decryptedCookies = extraCookiesJSON.String
			}
			if err := json.Unmarshal([]byte(decryptedCookies), &session.ExtraCookies); err != nil {
				return nil, err
			}
		}
		
		sessions = append(sessions, session)
	}
	
	return sessions, nil
}


// CleanupExpiredSessions 清理过期的 Cursor Sessions
// 只删除 expires_at 不为空且早于当前时间的 session
// 不会删除 expires_at 为 NULL 或零值的 session
func CleanupExpiredSessions() (int64, error) {
	now := time.Now()
	// 只删除有明确过期时间且已过期的 sessions
	// expires_at 必须不为 NULL，不为零值（1970-01-01），且早于当前时间
	result, err := db.Exec(
		`DELETE FROM cursor_sessions 
		 WHERE expires_at IS NOT NULL 
		 AND expires_at > '1970-01-02' 
		 AND expires_at < ?`,
		now,
	)
	if err != nil {
		return 0, err
	}
	
	return result.RowsAffected()
}

// GetExpiredSessions 获取所有过期的 Cursor Sessions（用于日志记录）
// 只返回有明确过期时间且已过期的 sessions
func GetExpiredSessions() ([]*models.CursorSessionInfo, error) {
	now := time.Now()
	rows, err := db.Query(
		`SELECT email, token, user_agent, extra_cookies, created_at, last_used, last_check, expires_at, is_valid, usage_count, fail_count,
		 daily_token_limit, daily_token_used, last_reset_date, quota_status, account_type
		 FROM cursor_sessions 
		 WHERE expires_at IS NOT NULL 
		 AND expires_at > '1970-01-02' 
		 AND expires_at < ?
		 ORDER BY expires_at ASC`,
		now,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var sessions []*models.CursorSessionInfo
	for rows.Next() {
		session := &models.CursorSessionInfo{}
		var userAgent sql.NullString
		var extraCookiesJSON sql.NullString
		var encryptedToken string
		var lastUsed sql.NullTime
		var lastCheck sql.NullTime
		var expiresAt sql.NullTime
		var lastResetDate sql.NullTime
		var quotaStatus sql.NullString
		var accountType sql.NullString
		
		err := rows.Scan(&session.Email, &encryptedToken, &userAgent, &extraCookiesJSON, 
			&session.CreatedAt, &lastUsed, &lastCheck, &expiresAt, 
			&session.IsValid, &session.UsageCount, &session.FailCount,
			&session.DailyTokenLimit, &session.DailyTokenUsed, &lastResetDate,
			&quotaStatus, &accountType)
		if err != nil {
			return nil, err
		}
		
		// 解密 token
		decryptedToken, err := utils.DecryptSensitiveData(encryptedToken)
		if err != nil {
			session.Token = encryptedToken
		} else {
			session.Token = decryptedToken
		}
		
		// 处理可能为 NULL 的字段
		if userAgent.Valid {
			session.UserAgent = userAgent.String
		}
		if lastUsed.Valid {
			session.LastUsed = lastUsed.Time
		}
		if lastCheck.Valid {
			session.LastCheck = lastCheck.Time
		}
		if expiresAt.Valid {
			session.ExpiresAt = expiresAt.Time
		}
		if lastResetDate.Valid {
			session.LastResetDate = lastResetDate.Time
		}
		if quotaStatus.Valid {
			session.QuotaStatus = quotaStatus.String
		}
		if accountType.Valid {
			session.AccountType = accountType.String
		}
		
		// 解密并反序列化 extra_cookies
		if extraCookiesJSON.Valid && extraCookiesJSON.String != "" {
			decryptedCookies, err := utils.DecryptSensitiveData(extraCookiesJSON.String)
			if err != nil {
				decryptedCookies = extraCookiesJSON.String
			}
			if err := json.Unmarshal([]byte(decryptedCookies), &session.ExtraCookies); err != nil {
				return nil, err
			}
		}
		
		sessions = append(sessions, session)
	}
	
	return sessions, nil
}


// MigrateEncryptCursorSessions 迁移现有的明文数据到加密格式
// 这个函数会检查每个 session 的 token 是否已加密，如果没有则加密它
func MigrateEncryptCursorSessions() (int, error) {
	rows, err := db.Query(
		`SELECT email, token, extra_cookies FROM cursor_sessions`,
	)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	migratedCount := 0
	for rows.Next() {
		var email, token string
		var extraCookies sql.NullString

		if err := rows.Scan(&email, &token, &extraCookies); err != nil {
			logrus.WithError(err).Error("Failed to scan cursor session for migration")
			continue
		}

		needsUpdate := false
		var encryptedToken, encryptedCookies string

		// 检查 token 是否已加密
		if !utils.IsEncrypted(token) && token != "" {
			encryptedToken, err = utils.EncryptSensitiveData(token)
			if err != nil {
				logrus.WithError(err).WithField("email", email).Error("Failed to encrypt token")
				continue
			}
			needsUpdate = true
		} else {
			encryptedToken = token
		}

		// 检查 extra_cookies 是否已加密
		if extraCookies.Valid && extraCookies.String != "" && !utils.IsEncrypted(extraCookies.String) {
			encryptedCookies, err = utils.EncryptSensitiveData(extraCookies.String)
			if err != nil {
				logrus.WithError(err).WithField("email", email).Error("Failed to encrypt extra cookies")
				continue
			}
			needsUpdate = true
		} else if extraCookies.Valid {
			encryptedCookies = extraCookies.String
		}

		// 更新数据库
		if needsUpdate {
			_, err = db.Exec(
				`UPDATE cursor_sessions SET token = ?, extra_cookies = ? WHERE email = ?`,
				encryptedToken, encryptedCookies, email,
			)
			if err != nil {
				logrus.WithError(err).WithField("email", email).Error("Failed to update encrypted data")
				continue
			}
			migratedCount++
			logrus.WithField("email", email).Info("Migrated cursor session to encrypted format")
		}
	}

	return migratedCount, nil
}
