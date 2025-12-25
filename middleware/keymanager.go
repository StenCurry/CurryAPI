package middleware

import (
	"Curry2API-go/database"
	"Curry2API-go/models"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Balance and token validation errors
// Requirements: 3.2, 12.4, 13.3, 14.3
var (
	ErrBalanceExhausted   = errors.New("balance exhausted - insufficient balance to make API calls")
	ErrTokenQuotaExceeded = errors.New("token quota exceeded - this token has reached its spending limit")
	ErrTokenExpired       = errors.New("token expired - this token has passed its expiration date")
	ErrModelNotAllowed    = errors.New("model not allowed - this token does not have access to the requested model")
)

// KeyError 密钥错误类型
type KeyError struct {
	Message string
	Code    string
}

func (e *KeyError) Error() string {
	return e.Message
}

// 错误代码常量
var (
	ErrEmptyKey     = &KeyError{Message: "密钥不能为空", Code: "empty_key"}
	ErrDuplicateKey = &KeyError{Message: "密钥已存在", Code: "duplicate_key"}
	ErrLastKey      = &KeyError{Message: "无法删除最后一个密钥", Code: "last_key"}
	ErrKeyNotFound  = &KeyError{Message: "密钥不存在", Code: "key_not_found"}
)

// KeyManager 密钥管理器（线程安全）
type KeyManager struct {
	mu         sync.RWMutex
	keys       map[string]*KeyInfo
	adminToken string
}

// KeyInfo 密钥信息
type KeyInfo = models.KeyInfo

var (
	keyManager     *KeyManager
	keyManagerOnce sync.Once
)

// GetKeyManager 获取密钥管理器单例
func GetKeyManager() *KeyManager {
	keyManagerOnce.Do(func() {
		keyManager = &KeyManager{
			keys:       make(map[string]*KeyInfo),
			adminToken: getAdminToken(),
		}

		// 优先从数据库加载密钥
		if err := keyManager.loadKeysFromDB(); err != nil {
			logrus.Errorf("Failed to load keys from database: %v", err)
			// 回退到环境变量
			keyManager.loadKeysFromEnv()
		} else if len(keyManager.keys) == 0 {
			// 数据库为空时也回退到环境变量，确保至少一个密钥
			keyManager.loadKeysFromEnv()
		}
	})
	return keyManager
}

// getAdminToken 获取管理员令牌
func getAdminToken() string {
	token := os.Getenv("ADMIN_KEY")
	if token == "" {
		return "admin-0000" // 默认管理员密钥
	}
	return token
}

// loadKeysFromEnv 从环境变量加载初始密钥
func (km *KeyManager) loadKeysFromEnv() {
	now := time.Now()

	// 优先加载 API_KEYS（多个密钥）
	if keysStr := os.Getenv("API_KEYS"); keysStr != "" {
		keys := strings.Split(keysStr, ",")
		for _, k := range keys {
			if trimmed := strings.TrimSpace(k); trimmed != "" {
				km.keys[trimmed] = &KeyInfo{
					Key:       trimmed,
					MaskedKey: maskKey(trimmed),
					CreatedAt: now,
				}
			}
		}
		if len(km.keys) > 0 {
			return
		}
	}

	// 回退到单个 API_KEY
	if key := os.Getenv("API_KEY"); key != "" {
		km.keys[key] = &KeyInfo{
			Key:       key,
			MaskedKey: maskKey(key),
			CreatedAt: now,
		}
		return
	}

	// 默认密钥
	km.keys["0000"] = &KeyInfo{
		Key:       "0000",
		MaskedKey: "0000",
		CreatedAt: now,
	}
}

// loadKeysFromDB 从数据库加载密钥
func (km *KeyManager) loadKeysFromDB() error {
	keys, err := database.ListAPIKeys()
	if err != nil {
		return err
	}

	km.mu.Lock()
	defer km.mu.Unlock()

	for keyStr := range km.keys {
		delete(km.keys, keyStr)
	}

	for _, k := range keys {
		if k == nil {
			continue
		}
		km.keys[k.Key] = &KeyInfo{
			Key:           k.Key,
			MaskedKey:     k.MaskedKey,
			TokenName:     k.TokenName,
			UserID:        k.UserID,
			Username:      k.Username,
			CreatedAt:     k.CreatedAt,
			UsageCount:    k.UsageCount,
			LastUsedAt:    k.LastUsedAt,
			IsActive:      k.IsActive,
			QuotaLimit:    k.QuotaLimit,
			QuotaUsed:     k.QuotaUsed,
			ExpiresAt:     k.ExpiresAt,
			AllowedModels: k.AllowedModels,
		}
	}

	logrus.Infof("Loaded %d API keys from database", len(km.keys))
	return nil
}

// maskKey 掩码密钥（保留前4后4，中间用*代替）
func maskKey(key string) string {
	keyLen := len(key)
	if keyLen <= 8 {
		return key // 太短不掩码
	}
	return key[:4] + strings.Repeat("*", keyLen-8) + key[keyLen-4:]
}

// MaskKey is the exported version of maskKey for use by other packages
func MaskKey(key string) string {
	return maskKey(key)
}

// GetAdminToken 获取管理员令牌（供 handlers 使用）
func (km *KeyManager) GetAdminToken() string {
	return km.adminToken
}

// GetAllKeys 获取所有密钥（用于认证验证）
func (km *KeyManager) GetAllKeys() []string {
	km.mu.RLock()
	defer km.mu.RUnlock()

	keys := make([]string, 0, len(km.keys))
	for k := range km.keys {
		keys = append(keys, k)
	}
	return keys
}

// IsValidKey 检查密钥是否有效（包括用户状态检查）
func (km *KeyManager) IsValidKey(key string) bool {
	// 首先检查内存中是否存在
	km.mu.RLock()
	_, exists := km.keys[key]
	km.mu.RUnlock()
	
	if !exists {
		return false
	}
	
	// 检查数据库中的实时状态（包括用户状态）
	isActive, err := database.IsKeyActiveWithUser(key)
	if err != nil {
		logrus.Warnf("Failed to check key status: %v", err)
		return false
	}
	
	return isActive
}

// IncrementUsage 增加密钥使用次数
func (km *KeyManager) IncrementUsage(key string) {
	km.mu.Lock()
	if info, exists := km.keys[key]; exists {
		info.UsageCount++
		km.mu.Unlock()

		// 异步更新数据库，减少请求阻塞
		go func() {
			if err := database.IncrementKeyUsage(key); err != nil {
				logrus.Warnf("Failed to update key usage in database: %v", err)
			}
		}()
		return
	}
	km.mu.Unlock()
}

// AddKey 添加新密钥（不关联用户）
func (km *KeyManager) AddKey(key string) error {
	return km.AddKeyWithUser(key, 0)
}

// AddKeyWithUser 添加新密钥并关联用户
func (km *KeyManager) AddKeyWithUser(key string, userID int64) error {
	return km.AddKeyWithUserAndName(key, userID, "")
}

// AddKeyWithUserAndName 添加新密钥并关联用户和名称
func (km *KeyManager) AddKeyWithUserAndName(key string, userID int64, tokenName string) error {
	if strings.TrimSpace(key) == "" {
		return ErrEmptyKey
	}

	km.mu.Lock()
	if _, exists := km.keys[key]; exists {
		km.mu.Unlock()
		return ErrDuplicateKey
	}
	km.mu.Unlock()

	maskedKey := maskKey(key)

	// 写数据库，关联用户ID和token名称
	var userIDPtr *int64
	if userID > 0 {
		userIDPtr = &userID
	}
	
	if err := database.AddAPIKeyWithName(key, userIDPtr, tokenName); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return ErrDuplicateKey
		}
		return fmt.Errorf("failed to save key to database: %w", err)
	}

	// 更新内存
	// 获取用户名（如果有用户ID）
	var username string
	if userID > 0 {
		if user, err := database.GetUserByID(userID); err == nil && user != nil {
			username = user.Username
		}
	}

	km.mu.Lock()
	defer km.mu.Unlock()
	km.keys[key] = &KeyInfo{
		Key:        key,
		MaskedKey:  maskedKey,
		TokenName:  tokenName,
		UserID:     userIDPtr,
		Username:   username,
		CreatedAt:  time.Now(),
		UsageCount: 0,
		IsActive:   true,
	}

	logrus.Infof("Added API key: %s (user_id: %d, token_name: %s)", maskedKey, userID, tokenName)
	return nil
}

// RemoveKey 删除密钥
func (km *KeyManager) RemoveKey(key string) error {
	km.mu.Lock()
	if _, exists := km.keys[key]; !exists {
		km.mu.Unlock()
		return ErrKeyNotFound
	}
	km.mu.Unlock()

	// 从数据库删除（软删除）
	if err := database.RemoveAPIKey(key); err != nil {
		return fmt.Errorf("failed to remove key from database: %w", err)
	}

	km.mu.Lock()
	defer km.mu.Unlock()

	if _, exists := km.keys[key]; !exists {
		return ErrKeyNotFound
	}

	delete(km.keys, key)
	logrus.Infof("Removed API key: %s", maskKey(key))
	return nil
}

// ListKeys 列出所有密钥信息（掩码后）
func (km *KeyManager) ListKeys() []*KeyInfo {
	km.mu.RLock()
	defer km.mu.RUnlock()

	result := make([]*KeyInfo, 0, len(km.keys))
	for _, info := range km.keys {
		// 创建副本避免暴露内部结构
		result = append(result, &KeyInfo{
			Key:           info.Key,
			MaskedKey:     info.MaskedKey,
			TokenName:     info.TokenName,
			UserID:        info.UserID,
			Username:      info.Username,
			CreatedAt:     info.CreatedAt,
			UsageCount:    info.UsageCount,
			LastUsedAt:    info.LastUsedAt,
			IsActive:      info.IsActive,
			QuotaLimit:    info.QuotaLimit,
			QuotaUsed:     info.QuotaUsed,
			ExpiresAt:     info.ExpiresAt,
			AllowedModels: info.AllowedModels,
		})
	}
	return result
}

// ToggleKeyStatus 切换密钥的启用/禁用状态
func (km *KeyManager) ToggleKeyStatus(key string) error {
	km.mu.Lock()
	info, exists := km.keys[key]
	if !exists {
		km.mu.Unlock()
		return ErrKeyNotFound
	}
	km.mu.Unlock()

	// 更新数据库
	if err := database.ToggleAPIKeyStatus(key); err != nil {
		return fmt.Errorf("failed to toggle key status in database: %w", err)
	}

	// 更新内存
	km.mu.Lock()
	defer km.mu.Unlock()
	info.IsActive = !info.IsActive
	
	logrus.Infof("Toggled API key status: %s (active: %v)", maskKey(key), info.IsActive)
	return nil
}

// ListKeysByUser 列出指定用户的密钥信息
func (km *KeyManager) ListKeysByUser(userID int64) []*KeyInfo {
	km.mu.RLock()
	defer km.mu.RUnlock()

	result := make([]*KeyInfo, 0)
	logrus.Debugf("ListKeysByUser: Looking for keys for userID=%d, total keys in memory=%d", userID, len(km.keys))
	
	for _, info := range km.keys {
		if info.UserID != nil {
			logrus.Debugf("ListKeysByUser: Key %s has userID=%d", info.MaskedKey, *info.UserID)
		} else {
			logrus.Debugf("ListKeysByUser: Key %s has no userID", info.MaskedKey)
		}
		
		// 只返回该用户创建的密钥
		if info.UserID != nil && *info.UserID == userID {
			result = append(result, &KeyInfo{
				Key:           info.Key,
				MaskedKey:     info.MaskedKey,
				TokenName:     info.TokenName,
				UserID:        info.UserID,
				Username:      info.Username,
				CreatedAt:     info.CreatedAt,
				UsageCount:    info.UsageCount,
				LastUsedAt:    info.LastUsedAt,
				IsActive:      info.IsActive,
				QuotaLimit:    info.QuotaLimit,
				QuotaUsed:     info.QuotaUsed,
				ExpiresAt:     info.ExpiresAt,
				AllowedModels: info.AllowedModels,
			})
		}
	}
	
	logrus.Debugf("ListKeysByUser: Found %d keys for userID=%d", len(result), userID)
	return result
}

// UpdateKeyName 更新密钥名称
func (km *KeyManager) UpdateKeyName(key, name string) error {
	km.mu.Lock()
	info, exists := km.keys[key]
	if !exists {
		km.mu.Unlock()
		return ErrKeyNotFound
	}
	km.mu.Unlock()

	// 更新数据库
	if err := database.UpdateAPIKeyName(key, name); err != nil {
		if err == database.ErrKeyNotFound {
			return ErrKeyNotFound
		}
		return fmt.Errorf("failed to update key name in database: %w", err)
	}

	// 更新内存
	km.mu.Lock()
	defer km.mu.Unlock()
	info.TokenName = name
	
	logrus.Infof("Updated API key name: %s (name: %s)", maskKey(key), name)
	return nil
}

// ============================================
// Balance and Token Validation Functions
// Requirements: 3.2, 12.4, 13.3, 14.3
// ============================================

// CheckBalanceStatus checks if the user associated with the token has sufficient balance
// Returns nil if balance is OK, ErrBalanceExhausted if balance is exhausted
// Requirements: 3.2
func (km *KeyManager) CheckBalanceStatus(key string) error {
	// Get the key info to find the user ID
	km.mu.RLock()
	keyInfo, exists := km.keys[key]
	km.mu.RUnlock()
	
	if !exists {
		return ErrKeyNotFound
	}
	
	// If no user is associated with this key, skip balance check
	if keyInfo.UserID == nil {
		return nil
	}
	
	// Check user's balance status from database
	balance, err := database.GetUserBalance(*keyInfo.UserID)
	if err != nil {
		if err == database.ErrBalanceNotFound {
			// No balance record means user hasn't been set up with balance system
			// Allow the request to proceed
			return nil
		}
		logrus.Warnf("Failed to check balance status for user %d: %v", *keyInfo.UserID, err)
		return nil // Don't block on database errors
	}
	
	// Check if balance is exhausted
	if balance.Status == database.BalanceStatusExhausted {
		logrus.Warnf("Balance exhausted for user %d, token %s", *keyInfo.UserID, maskKey(key))
		return ErrBalanceExhausted
	}
	
	return nil
}

// CheckTokenQuota checks if the token has exceeded its quota limit
// Returns nil if quota is OK or unlimited, ErrTokenQuotaExceeded if quota is exceeded
// Requirements: 12.4
func (km *KeyManager) CheckTokenQuota(key string) error {
	canUse, err := database.CheckTokenQuota(key)
	if err != nil {
		if err == database.ErrKeyNotFound {
			return ErrKeyNotFound
		}
		if err == database.ErrTokenQuotaExceeded {
			logrus.Warnf("Token quota exceeded for key %s", maskKey(key))
			return ErrTokenQuotaExceeded
		}
		logrus.Warnf("Failed to check token quota for key %s: %v", maskKey(key), err)
		return nil // Don't block on database errors
	}
	
	if !canUse {
		return ErrTokenQuotaExceeded
	}
	
	return nil
}

// CheckTokenExpiration checks if the token has expired
// Returns nil if token is valid or has no expiration, ErrTokenExpired if expired
// Requirements: 13.3
func (km *KeyManager) CheckTokenExpiration(key string) error {
	canUse, err := database.CheckTokenExpiration(key)
	if err != nil {
		if err == database.ErrKeyNotFound {
			return ErrKeyNotFound
		}
		if err == database.ErrTokenExpired {
			logrus.Warnf("Token expired for key %s", maskKey(key))
			return ErrTokenExpired
		}
		logrus.Warnf("Failed to check token expiration for key %s: %v", maskKey(key), err)
		return nil // Don't block on database errors
	}
	
	if !canUse {
		return ErrTokenExpired
	}
	
	return nil
}

// CheckTokenModelAccess checks if the token is allowed to access the specified model
// Returns nil if model is allowed or no restrictions, ErrModelNotAllowed if not allowed
// Requirements: 14.3
func (km *KeyManager) CheckTokenModelAccess(key, model string) error {
	canUse, err := database.CheckTokenModelAccess(key, model)
	if err != nil {
		if err == database.ErrKeyNotFound {
			return ErrKeyNotFound
		}
		if err == database.ErrModelNotAllowed {
			logrus.Warnf("Model %s not allowed for key %s", model, maskKey(key))
			return ErrModelNotAllowed
		}
		logrus.Warnf("Failed to check model access for key %s: %v", maskKey(key), err)
		return nil // Don't block on database errors
	}
	
	if !canUse {
		return ErrModelNotAllowed
	}
	
	return nil
}

// ValidateTokenForRequest performs all validation checks for a token before an API request
// This includes: balance status, token quota, token expiration, and model access
// Returns nil if all checks pass, or the first error encountered
// Requirements: 3.2, 12.4, 13.3, 14.3
func (km *KeyManager) ValidateTokenForRequest(key, model string) error {
	// 1. Check balance status
	if err := km.CheckBalanceStatus(key); err != nil {
		return err
	}
	
	// 2. Check token quota
	if err := km.CheckTokenQuota(key); err != nil {
		return err
	}
	
	// 3. Check token expiration
	if err := km.CheckTokenExpiration(key); err != nil {
		return err
	}
	
	// 4. Check model access (only if model is specified)
	if model != "" {
		if err := km.CheckTokenModelAccess(key, model); err != nil {
			return err
		}
	}
	
	return nil
}

// GetUserIDForKey returns the user ID associated with a key, or nil if not found
func (km *KeyManager) GetUserIDForKey(key string) *int64 {
	km.mu.RLock()
	defer km.mu.RUnlock()
	
	if keyInfo, exists := km.keys[key]; exists {
		return keyInfo.UserID
	}
	return nil
}

// ReloadKeys reloads all keys from the database
func (km *KeyManager) ReloadKeys() error {
	return km.loadKeysFromDB()
}
