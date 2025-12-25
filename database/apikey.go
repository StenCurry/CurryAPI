package database

import (
	"Curry2API-go/models"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	ErrKeyNotFound      = errors.New("api key not found")
	ErrKeyExists        = errors.New("api key already exists")
	ErrTokenQuotaExceeded = errors.New("token quota exceeded")
	ErrTokenExpired       = errors.New("token has expired")
	ErrModelNotAllowed    = errors.New("model not allowed for this token")
)

// AddAPIKey 添加API密钥
func AddAPIKey(key string, userID *int64) error {
	return AddAPIKeyWithName(key, userID, "")
}

// AddAPIKeyWithName 添加API密钥（带名称）
func AddAPIKeyWithName(key string, userID *int64, tokenName string) error {
	// 生成掩码密钥
	maskedKey := maskKey(key)
	
	_, err := db.Exec(
		"INSERT INTO api_keys (key_value, masked_key, token_name, user_id, created_at, usage_count, is_active) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?)",
		key, maskedKey, tokenName, userID, time.Now(), 0, true,
	)
	if err != nil {
		// Log the error for debugging
		fmt.Printf("AddAPIKeyWithName error: %v\n", err)
	}
	return err
}

// APIKeyOptions contains optional parameters for creating an API key
type APIKeyOptions struct {
	QuotaLimit    *float64   // Quota limit in USD, nil means unlimited
	ExpiresAt     *time.Time // Expiration time, nil means never expires
	AllowedModels []string   // Allowed models, nil/empty means all models
}

// AddAPIKeyWithOptions 添加API密钥（带完整选项）
// Requirements: 12.1, 13.1, 14.1
func AddAPIKeyWithOptions(key string, userID *int64, tokenName string, opts *APIKeyOptions) error {
	maskedKey := maskKey(key)
	
	var quotaLimit *float64
	var expiresAt *time.Time
	var allowedModelsJSON *string
	
	if opts != nil {
		quotaLimit = opts.QuotaLimit
		expiresAt = opts.ExpiresAt
		if len(opts.AllowedModels) > 0 {
			jsonBytes, err := json.Marshal(opts.AllowedModels)
			if err != nil {
				return fmt.Errorf("failed to marshal allowed_models: %w", err)
			}
			jsonStr := string(jsonBytes)
			allowedModelsJSON = &jsonStr
		}
	}
	
	_, err := db.Exec(
		"INSERT INTO api_keys (key_value, masked_key, token_name, user_id, created_at, usage_count, is_active, quota_limit, quota_used, expires_at, allowed_models) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		key, maskedKey, tokenName, userID, time.Now(), 0, true, quotaLimit, 0.0, expiresAt, allowedModelsJSON,
	)
	if err != nil {
		fmt.Printf("AddAPIKeyWithOptions error: %v\n", err)
	}
	return err
}

// GetAPIKey 获取API密钥信息
func GetAPIKey(key string) (*models.KeyInfo, error) {
	keyInfo := &models.KeyInfo{}
	var tokenName sql.NullString
	var lastUsedAt sql.NullTime
	var quotaLimit sql.NullFloat64
	var quotaUsed sql.NullFloat64
	var expiresAt sql.NullTime
	var allowedModelsJSON sql.NullString
	
	err := db.QueryRow(
		"SELECT key_value, masked_key, token_name, user_id, created_at, usage_count, last_used_at, is_active, "+
			"quota_limit, quota_used, expires_at, allowed_models "+
			"FROM api_keys WHERE key_value = ? AND is_active = TRUE",
		key,
	).Scan(&keyInfo.Key, &keyInfo.MaskedKey, &tokenName, &keyInfo.UserID, &keyInfo.CreatedAt, &keyInfo.UsageCount, 
		&lastUsedAt, &keyInfo.IsActive, &quotaLimit, &quotaUsed, &expiresAt, &allowedModelsJSON)
	
	if err == sql.ErrNoRows {
		return nil, ErrKeyNotFound
	}
	if err != nil {
		return nil, err
	}
	
	if tokenName.Valid {
		keyInfo.TokenName = tokenName.String
	}
	if lastUsedAt.Valid {
		keyInfo.LastUsedAt = &lastUsedAt.Time
	}
	if quotaLimit.Valid {
		keyInfo.QuotaLimit = &quotaLimit.Float64
	}
	if quotaUsed.Valid {
		keyInfo.QuotaUsed = quotaUsed.Float64
	}
	if expiresAt.Valid {
		keyInfo.ExpiresAt = &expiresAt.Time
	}
	if allowedModelsJSON.Valid && allowedModelsJSON.String != "" {
		var models []string
		if err := json.Unmarshal([]byte(allowedModelsJSON.String), &models); err == nil {
			keyInfo.AllowedModels = models
		}
	}
	
	return keyInfo, nil
}

// ListAPIKeys 列出所有API密钥（包含用户名）
func ListAPIKeys() ([]*models.KeyInfo, error) {
	rows, err := db.Query(
		"SELECT k.key_value, k.masked_key, k.token_name, k.user_id, k.created_at, k.usage_count, k.last_used_at, k.is_active, " +
			"k.quota_limit, k.quota_used, k.expires_at, k.allowed_models, u.username " +
			"FROM api_keys k " +
			"LEFT JOIN users u ON k.user_id = u.id " +
			"WHERE k.is_active = TRUE " +
			"ORDER BY k.created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var keys []*models.KeyInfo
	for rows.Next() {
		key := &models.KeyInfo{}
		var username sql.NullString
		var tokenName sql.NullString
		var lastUsedAt sql.NullTime
		var quotaLimit sql.NullFloat64
		var quotaUsed sql.NullFloat64
		var expiresAt sql.NullTime
		var allowedModelsJSON sql.NullString
		
		err := rows.Scan(&key.Key, &key.MaskedKey, &tokenName, &key.UserID, &key.CreatedAt, &key.UsageCount, 
			&lastUsedAt, &key.IsActive, &quotaLimit, &quotaUsed, &expiresAt, &allowedModelsJSON, &username)
		if err != nil {
			return nil, err
		}
		if username.Valid {
			key.Username = username.String
		}
		if tokenName.Valid {
			key.TokenName = tokenName.String
		}
		if lastUsedAt.Valid {
			key.LastUsedAt = &lastUsedAt.Time
		}
		if quotaLimit.Valid {
			key.QuotaLimit = &quotaLimit.Float64
		}
		if quotaUsed.Valid {
			key.QuotaUsed = quotaUsed.Float64
		}
		if expiresAt.Valid {
			key.ExpiresAt = &expiresAt.Time
		}
		if allowedModelsJSON.Valid && allowedModelsJSON.String != "" {
			var models []string
			if err := json.Unmarshal([]byte(allowedModelsJSON.String), &models); err == nil {
				key.AllowedModels = models
			}
		}
		keys = append(keys, key)
	}
	
	return keys, nil
}

// RemoveAPIKey 删除API密钥
func RemoveAPIKey(key string) error {
	result, err := db.Exec("DELETE FROM api_keys WHERE key_value = ?", key)
	if err != nil {
		return err
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rows == 0 {
		return ErrKeyNotFound
	}
	
	return nil
}

// IncrementKeyUsage 增加密钥使用次数
func IncrementKeyUsage(key string) error {
	_, err := db.Exec(
		"UPDATE api_keys SET usage_count = usage_count + 1 WHERE key_value = ?",
		key,
	)
	return err
}

// maskKey 生成掩码密钥
func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return fmt.Sprintf("%s****%s", key[:4], key[len(key)-4:])
}

// UpdateAPIKeyStatusByUser 更新指定用户的所有API密钥状态
func UpdateAPIKeyStatusByUser(userID int64, isActive bool) error {
	_, err := db.Exec(
		`UPDATE api_keys SET is_active = ? WHERE user_id = ?`,
		isActive, userID,
	)
	return err
}

// ToggleAPIKeyStatus 切换API密钥的启用/禁用状态
func ToggleAPIKeyStatus(key string) error {
	_, err := db.Exec(
		"UPDATE api_keys SET is_active = NOT is_active WHERE key_value = ?",
		key,
	)
	return err
}

// IsKeyActiveWithUser 检查API密钥是否有效（包括用户状态检查）
func IsKeyActiveWithUser(key string) (bool, error) {
	var isActive bool
	var userID *int64
	
	err := db.QueryRow(
		"SELECT k.is_active, k.user_id "+
			"FROM api_keys k "+
			"WHERE k.key_value = ?",
		key,
	).Scan(&isActive, &userID)
	
	if err == sql.ErrNoRows {
		return false, ErrKeyNotFound
	}
	if err != nil {
		return false, err
	}
	
	// 如果密钥本身被禁用，返回false
	if !isActive {
		return false, nil
	}
	
	// 如果密钥关联了用户，检查用户状态
	if userID != nil {
		var userActive bool
		err = db.QueryRow(
			`SELECT is_active FROM users WHERE id = ?`,
			*userID,
		).Scan(&userActive)
		
		if err != nil {
			return false, err
		}
		
		// 如果用户被禁用，密钥也无效
		if !userActive {
			return false, nil
		}
	}
	
	return true, nil
}

// UpdateAPIKeyName 更新API密钥的名称
func UpdateAPIKeyName(key, name string) error {
	result, err := db.Exec(
		"UPDATE api_keys SET token_name = ? WHERE key_value = ?",
		name, key,
	)
	if err != nil {
		return err
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rows == 0 {
		return ErrKeyNotFound
	}
	
	return nil
}

// UpdateAPIKeyLastUsed 更新API密钥的最后使用时间
func UpdateAPIKeyLastUsed(key string, timestamp time.Time) error {
	_, err := db.Exec(
		"UPDATE api_keys SET last_used_at = ? WHERE key_value = ?",
		timestamp, key,
	)
	return err
}

// CheckTokenQuota checks if a token has exceeded its quota limit
// Returns true if the token can be used (quota not exceeded or unlimited)
// Returns false with ErrTokenQuotaExceeded if quota is exceeded
// Requirements: 12.2, 12.3
func CheckTokenQuota(key string) (bool, error) {
	var quotaLimit sql.NullFloat64
	var quotaUsed sql.NullFloat64
	var isActive bool
	
	err := db.QueryRow(
		"SELECT quota_limit, quota_used, is_active FROM api_keys WHERE key_value = ?",
		key,
	).Scan(&quotaLimit, &quotaUsed, &isActive)
	
	if err == sql.ErrNoRows {
		return false, ErrKeyNotFound
	}
	if err != nil {
		return false, err
	}
	
	if !isActive {
		return false, ErrKeyNotFound
	}
	
	// If quota_limit is NULL, the token has unlimited quota
	if !quotaLimit.Valid {
		return true, nil
	}
	
	// Check if quota_used has reached or exceeded quota_limit
	used := 0.0
	if quotaUsed.Valid {
		used = quotaUsed.Float64
	}
	
	if used >= quotaLimit.Float64 {
		return false, ErrTokenQuotaExceeded
	}
	
	return true, nil
}

// CheckTokenQuotaWithInfo checks quota and returns detailed info
// Returns (canUse, quotaLimit, quotaUsed, error)
func CheckTokenQuotaWithInfo(key string) (bool, *float64, float64, error) {
	var quotaLimit sql.NullFloat64
	var quotaUsed sql.NullFloat64
	var isActive bool
	
	err := db.QueryRow(
		"SELECT quota_limit, quota_used, is_active FROM api_keys WHERE key_value = ?",
		key,
	).Scan(&quotaLimit, &quotaUsed, &isActive)
	
	if err == sql.ErrNoRows {
		return false, nil, 0, ErrKeyNotFound
	}
	if err != nil {
		return false, nil, 0, err
	}
	
	if !isActive {
		return false, nil, 0, ErrKeyNotFound
	}
	
	var limit *float64
	if quotaLimit.Valid {
		limit = &quotaLimit.Float64
	}
	
	used := 0.0
	if quotaUsed.Valid {
		used = quotaUsed.Float64
	}
	
	// If quota_limit is NULL, the token has unlimited quota
	if limit == nil {
		return true, nil, used, nil
	}
	
	// Check if quota_used has reached or exceeded quota_limit
	if used >= *limit {
		return false, limit, used, ErrTokenQuotaExceeded
	}
	
	return true, limit, used, nil
}

// CheckTokenExpiration checks if a token has expired
// Returns true if the token can be used (not expired or no expiration set)
// Returns false with ErrTokenExpired if the token has expired
// Requirements: 13.2
func CheckTokenExpiration(key string) (bool, error) {
	var expiresAt sql.NullTime
	var isActive bool
	
	err := db.QueryRow(
		"SELECT expires_at, is_active FROM api_keys WHERE key_value = ?",
		key,
	).Scan(&expiresAt, &isActive)
	
	if err == sql.ErrNoRows {
		return false, ErrKeyNotFound
	}
	if err != nil {
		return false, err
	}
	
	if !isActive {
		return false, ErrKeyNotFound
	}
	
	// If expires_at is NULL, the token never expires
	if !expiresAt.Valid {
		return true, nil
	}
	
	// Check if current time is past expiration
	if time.Now().After(expiresAt.Time) {
		return false, ErrTokenExpired
	}
	
	return true, nil
}

// CheckTokenExpirationWithInfo checks expiration and returns detailed info
// Returns (canUse, expiresAt, error)
func CheckTokenExpirationWithInfo(key string) (bool, *time.Time, error) {
	var expiresAt sql.NullTime
	var isActive bool
	
	err := db.QueryRow(
		"SELECT expires_at, is_active FROM api_keys WHERE key_value = ?",
		key,
	).Scan(&expiresAt, &isActive)
	
	if err == sql.ErrNoRows {
		return false, nil, ErrKeyNotFound
	}
	if err != nil {
		return false, nil, err
	}
	
	if !isActive {
		return false, nil, ErrKeyNotFound
	}
	
	var expTime *time.Time
	if expiresAt.Valid {
		expTime = &expiresAt.Time
	}
	
	// If expires_at is NULL, the token never expires
	if expTime == nil {
		return true, nil, nil
	}
	
	// Check if current time is past expiration
	if time.Now().After(*expTime) {
		return false, expTime, ErrTokenExpired
	}
	
	return true, expTime, nil
}


// CheckTokenModelAccess checks if a token is allowed to access a specific model
// Returns true if the token can access the model (model in allowed list or no restrictions)
// Returns false with ErrModelNotAllowed if the model is not in the allowed list
// Requirements: 14.2
func CheckTokenModelAccess(key string, model string) (bool, error) {
	var allowedModelsJSON sql.NullString
	var isActive bool
	
	err := db.QueryRow(
		"SELECT allowed_models, is_active FROM api_keys WHERE key_value = ?",
		key,
	).Scan(&allowedModelsJSON, &isActive)
	
	if err == sql.ErrNoRows {
		return false, ErrKeyNotFound
	}
	if err != nil {
		return false, err
	}
	
	if !isActive {
		return false, ErrKeyNotFound
	}
	
	// If allowed_models is NULL or empty, all models are allowed
	if !allowedModelsJSON.Valid || allowedModelsJSON.String == "" {
		return true, nil
	}
	
	// Parse the JSON array of allowed models
	var allowedModels []string
	if err := json.Unmarshal([]byte(allowedModelsJSON.String), &allowedModels); err != nil {
		// If parsing fails, treat as no restrictions
		return true, nil
	}
	
	// If the list is empty, all models are allowed
	if len(allowedModels) == 0 {
		return true, nil
	}
	
	// Check if the requested model is in the allowed list
	for _, allowed := range allowedModels {
		if allowed == model {
			return true, nil
		}
	}
	
	return false, ErrModelNotAllowed
}

// CheckTokenModelAccessWithInfo checks model access and returns the allowed models list
// Returns (canUse, allowedModels, error)
func CheckTokenModelAccessWithInfo(key string, model string) (bool, []string, error) {
	var allowedModelsJSON sql.NullString
	var isActive bool
	
	err := db.QueryRow(
		"SELECT allowed_models, is_active FROM api_keys WHERE key_value = ?",
		key,
	).Scan(&allowedModelsJSON, &isActive)
	
	if err == sql.ErrNoRows {
		return false, nil, ErrKeyNotFound
	}
	if err != nil {
		return false, nil, err
	}
	
	if !isActive {
		return false, nil, ErrKeyNotFound
	}
	
	// If allowed_models is NULL or empty, all models are allowed
	if !allowedModelsJSON.Valid || allowedModelsJSON.String == "" {
		return true, nil, nil
	}
	
	// Parse the JSON array of allowed models
	var allowedModels []string
	if err := json.Unmarshal([]byte(allowedModelsJSON.String), &allowedModels); err != nil {
		// If parsing fails, treat as no restrictions
		return true, nil, nil
	}
	
	// If the list is empty, all models are allowed
	if len(allowedModels) == 0 {
		return true, nil, nil
	}
	
	// Check if the requested model is in the allowed list
	for _, allowed := range allowedModels {
		if allowed == model {
			return true, allowedModels, nil
		}
	}
	
	return false, allowedModels, ErrModelNotAllowed
}


// UpdateTokenQuotaUsed increments the quota_used for a token after an API call
// The amount should be the cost in USD for the API call
// Requirements: 12.2
func UpdateTokenQuotaUsed(key string, amount float64) error {
	result, err := db.Exec(
		"UPDATE api_keys SET quota_used = quota_used + ? WHERE key_value = ?",
		amount, key,
	)
	if err != nil {
		return err
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rows == 0 {
		return ErrKeyNotFound
	}
	
	return nil
}

// DisableTokenIfQuotaExceeded checks if a token's quota is exceeded and disables it if so
// Returns true if the token was disabled, false otherwise
// Requirements: 12.3
func DisableTokenIfQuotaExceeded(key string) (bool, error) {
	var quotaLimit sql.NullFloat64
	var quotaUsed sql.NullFloat64
	
	err := db.QueryRow(
		"SELECT quota_limit, quota_used FROM api_keys WHERE key_value = ?",
		key,
	).Scan(&quotaLimit, &quotaUsed)
	
	if err == sql.ErrNoRows {
		return false, ErrKeyNotFound
	}
	if err != nil {
		return false, err
	}
	
	// If no quota limit, nothing to do
	if !quotaLimit.Valid {
		return false, nil
	}
	
	used := 0.0
	if quotaUsed.Valid {
		used = quotaUsed.Float64
	}
	
	// If quota exceeded, disable the token
	if used >= quotaLimit.Float64 {
		_, err := db.Exec(
			"UPDATE api_keys SET is_active = FALSE WHERE key_value = ?",
			key,
		)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	
	return false, nil
}

// GetTokenQuotaInfo returns the quota information for a token
// Returns (quotaLimit, quotaUsed, error) where quotaLimit is nil for unlimited tokens
func GetTokenQuotaInfo(key string) (*float64, float64, error) {
	var quotaLimit sql.NullFloat64
	var quotaUsed sql.NullFloat64
	
	err := db.QueryRow(
		"SELECT quota_limit, quota_used FROM api_keys WHERE key_value = ?",
		key,
	).Scan(&quotaLimit, &quotaUsed)
	
	if err == sql.ErrNoRows {
		return nil, 0, ErrKeyNotFound
	}
	if err != nil {
		return nil, 0, err
	}
	
	var limit *float64
	if quotaLimit.Valid {
		limit = &quotaLimit.Float64
	}
	
	used := 0.0
	if quotaUsed.Valid {
		used = quotaUsed.Float64
	}
	
	return limit, used, nil
}
