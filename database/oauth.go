package database

import (
	"Curry2API-go/utils"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Global OAuth crypto instance
var oauthCrypto *utils.OAuthCrypto

// InitOAuthCrypto 初始化 OAuth 加密工具
func InitOAuthCrypto() error {
	crypto, err := utils.NewOAuthCrypto()
	if err != nil {
		return fmt.Errorf("failed to initialize OAuth crypto: %w", err)
	}
	oauthCrypto = crypto
	logrus.Info("OAuth token encryption initialized")
	return nil
}

// OAuthAccount OAuth账号关联
type OAuthAccount struct {
	ID             int64
	UserID         int
	Provider       string
	ProviderUserID string
	Email          string
	Username       string
	AvatarURL      string
	AccessToken    string
	RefreshToken   string
	TokenExpiresAt *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// OAuthState OAuth状态令牌
type OAuthState struct {
	State     string
	Provider  string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// CreateOAuthState 创建OAuth状态令牌
func CreateOAuthState(state, provider string, expiresAt time.Time) error {
	query := `
		INSERT INTO oauth_states (state, provider, expires_at)
		VALUES (?, ?, ?)
	`
	result, err := db.Exec(query, state, provider, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create oauth state: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	logrus.WithFields(logrus.Fields{
		"provider":      provider,
		"state":         state[:10] + "...",
		"expires_at":    expiresAt,
		"rows_affected": rowsAffected,
	}).Debug("OAuth state created in database")
	
	return nil
}

// VerifyOAuthState 验证OAuth状态令牌
func VerifyOAuthState(state, provider string) (bool, error) {
	query := `
		SELECT state, provider, expires_at
		FROM oauth_states
		WHERE state = ? AND provider = ?
	`
	var s OAuthState
	err := db.QueryRow(query, state, provider).Scan(&s.State, &s.Provider, &s.ExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			logrus.WithFields(logrus.Fields{
				"provider": provider,
				"state":    state[:10] + "...",
			}).Warn("OAuth state not found in database")
			return false, nil
		}
		return false, fmt.Errorf("failed to verify oauth state: %w", err)
	}

	// 检查是否过期
	if time.Now().After(s.ExpiresAt) {
		logrus.WithFields(logrus.Fields{
			"provider":   provider,
			"state":      state[:10] + "...",
			"expires_at": s.ExpiresAt,
			"now":        time.Now(),
		}).Warn("OAuth state has expired")
		// 删除过期的state
		_ = DeleteOAuthState(state)
		return false, nil
	}

	logrus.WithFields(logrus.Fields{
		"provider": provider,
		"state":    state[:10] + "...",
	}).Debug("OAuth state verified successfully")
	return true, nil
}

// DeleteOAuthState 删除OAuth状态令牌
func DeleteOAuthState(state string) error {
	query := `DELETE FROM oauth_states WHERE state = ?`
	_, err := db.Exec(query, state)
	if err != nil {
		return fmt.Errorf("failed to delete oauth state: %w", err)
	}
	return nil
}

// CleanupExpiredOAuthStates 清理过期的OAuth状态令牌
func CleanupExpiredOAuthStates() error {
	query := `DELETE FROM oauth_states WHERE expires_at < NOW()`
	result, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired oauth states: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows > 0 {
		logrus.Infof("Cleaned up %d expired OAuth states", rows)
	}
	return nil
}

// ListOAuthStates 列出所有OAuth状态令牌（调试用）
func ListOAuthStates() ([]OAuthState, error) {
	query := `SELECT state, provider, created_at, expires_at FROM oauth_states ORDER BY created_at DESC LIMIT 10`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list oauth states: %w", err)
	}
	defer rows.Close()

	var states []OAuthState
	for rows.Next() {
		var s OAuthState
		if err := rows.Scan(&s.State, &s.Provider, &s.CreatedAt, &s.ExpiresAt); err != nil {
			return nil, fmt.Errorf("failed to scan oauth state: %w", err)
		}
		states = append(states, s)
	}
	return states, nil
}

// GetOAuthAccountByProvider 根据提供商和提供商用户ID获取OAuth账号
func GetOAuthAccountByProvider(provider, providerUserID string) (*OAuthAccount, error) {
	query := `
		SELECT id, user_id, provider, provider_user_id, email, username, avatar_url,
		       access_token, refresh_token, token_expires_at, created_at, updated_at
		FROM oauth_accounts
		WHERE provider = ? AND provider_user_id = ?
	`
	var account OAuthAccount
	var tokenExpiresAt sql.NullTime
	var encryptedAccessToken, encryptedRefreshToken string

	err := db.QueryRow(query, provider, providerUserID).Scan(
		&account.ID,
		&account.UserID,
		&account.Provider,
		&account.ProviderUserID,
		&account.Email,
		&account.Username,
		&account.AvatarURL,
		&encryptedAccessToken,
		&encryptedRefreshToken,
		&tokenExpiresAt,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get oauth account: %w", err)
	}

	if tokenExpiresAt.Valid {
		account.TokenExpiresAt = &tokenExpiresAt.Time
	}

	// 解密 tokens
	if oauthCrypto != nil {
		if encryptedAccessToken != "" {
			decrypted, err := oauthCrypto.DecryptAccessToken(encryptedAccessToken)
			if err != nil {
				logrus.WithError(err).Warn("Failed to decrypt access token")
			} else {
				account.AccessToken = decrypted
			}
		}
		
		if encryptedRefreshToken != "" {
			decrypted, err := oauthCrypto.DecryptRefreshToken(encryptedRefreshToken)
			if err != nil {
				logrus.WithError(err).Warn("Failed to decrypt refresh token")
			} else {
				account.RefreshToken = decrypted
			}
		}
	}

	return &account, nil
}

// GetOAuthAccountsByUserID 根据用户ID获取所有OAuth账号
func GetOAuthAccountsByUserID(userID int) ([]*OAuthAccount, error) {
	query := `
		SELECT id, user_id, provider, provider_user_id, email, username, avatar_url,
		       access_token, refresh_token, token_expires_at, created_at, updated_at
		FROM oauth_accounts
		WHERE user_id = ?
		ORDER BY created_at DESC
	`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*OAuthAccount
	for rows.Next() {
		var account OAuthAccount
		var tokenExpiresAt sql.NullTime
		var encryptedAccessToken, encryptedRefreshToken string

		err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.Provider,
			&account.ProviderUserID,
			&account.Email,
			&account.Username,
			&account.AvatarURL,
			&encryptedAccessToken,
			&encryptedRefreshToken,
			&tokenExpiresAt,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan oauth account: %w", err)
		}

		if tokenExpiresAt.Valid {
			account.TokenExpiresAt = &tokenExpiresAt.Time
		}

		// 解密 tokens
		if oauthCrypto != nil {
			if encryptedAccessToken != "" {
				decrypted, err := oauthCrypto.DecryptAccessToken(encryptedAccessToken)
				if err != nil {
					logrus.WithError(err).Warn("Failed to decrypt access token")
				} else {
					account.AccessToken = decrypted
				}
			}
			
			if encryptedRefreshToken != "" {
				decrypted, err := oauthCrypto.DecryptRefreshToken(encryptedRefreshToken)
				if err != nil {
					logrus.WithError(err).Warn("Failed to decrypt refresh token")
				} else {
					account.RefreshToken = decrypted
				}
			}
		}

		accounts = append(accounts, &account)
	}

	return accounts, nil
}

// CreateOAuthAccount 创建OAuth账号关联
func CreateOAuthAccount(account *OAuthAccount) error {
	// 加密 tokens
	var encryptedAccessToken, encryptedRefreshToken string
	var err error
	
	if oauthCrypto != nil {
		if account.AccessToken != "" {
			encryptedAccessToken, err = oauthCrypto.EncryptAccessToken(account.AccessToken)
			if err != nil {
				return fmt.Errorf("failed to encrypt access token: %w", err)
			}
		}
		
		if account.RefreshToken != "" {
			encryptedRefreshToken, err = oauthCrypto.EncryptRefreshToken(account.RefreshToken)
			if err != nil {
				return fmt.Errorf("failed to encrypt refresh token: %w", err)
			}
		}
	} else {
		// 如果加密未初始化，使用原始值（不推荐）
		logrus.Warn("OAuth crypto not initialized, storing tokens without encryption")
		encryptedAccessToken = account.AccessToken
		encryptedRefreshToken = account.RefreshToken
	}
	
	query := `
		INSERT INTO oauth_accounts (
			user_id, provider, provider_user_id, email, username, avatar_url,
			access_token, refresh_token, token_expires_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := db.Exec(
		query,
		account.UserID,
		account.Provider,
		account.ProviderUserID,
		account.Email,
		account.Username,
		account.AvatarURL,
		encryptedAccessToken,
		encryptedRefreshToken,
		account.TokenExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create oauth account: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	account.ID = id
	return nil
}

// UpdateOAuthAccount 更新OAuth账号信息
func UpdateOAuthAccount(account *OAuthAccount) error {
	// 加密 tokens
	var encryptedAccessToken, encryptedRefreshToken string
	var err error
	
	if oauthCrypto != nil {
		if account.AccessToken != "" {
			encryptedAccessToken, err = oauthCrypto.EncryptAccessToken(account.AccessToken)
			if err != nil {
				return fmt.Errorf("failed to encrypt access token: %w", err)
			}
		}
		
		if account.RefreshToken != "" {
			encryptedRefreshToken, err = oauthCrypto.EncryptRefreshToken(account.RefreshToken)
			if err != nil {
				return fmt.Errorf("failed to encrypt refresh token: %w", err)
			}
		}
	} else {
		// 如果加密未初始化，使用原始值（不推荐）
		logrus.Warn("OAuth crypto not initialized, storing tokens without encryption")
		encryptedAccessToken = account.AccessToken
		encryptedRefreshToken = account.RefreshToken
	}
	
	query := `
		UPDATE oauth_accounts
		SET email = ?, username = ?, avatar_url = ?,
		    access_token = ?, refresh_token = ?, token_expires_at = ?
		WHERE id = ?
	`
	_, err = db.Exec(
		query,
		account.Email,
		account.Username,
		account.AvatarURL,
		encryptedAccessToken,
		encryptedRefreshToken,
		account.TokenExpiresAt,
		account.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update oauth account: %w", err)
	}

	return nil
}

// DeleteOAuthAccount 删除OAuth账号关联
func DeleteOAuthAccount(id int64) error {
	query := `DELETE FROM oauth_accounts WHERE id = ?`
	_, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete oauth account: %w", err)
	}
	return nil
}

// DeleteOAuthAccountsByUserID 删除用户的所有OAuth账号关联
func DeleteOAuthAccountsByUserID(userID int) error {
	query := `DELETE FROM oauth_accounts WHERE user_id = ?`
	_, err := db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete oauth accounts: %w", err)
	}
	return nil
}

// OAuthUserInfo OAuth用户信息
type OAuthUserInfo struct {
	ProviderUserID string
	Email          string
	Username       string
	AvatarURL      string
	EmailVerified  bool
}

// FindOrCreateUserFromOAuth 从OAuth信息查找或创建用户
// 实现邮箱匹配查找现有用户，如果不存在则创建新用户
func FindOrCreateUserFromOAuth(oauthInfo *OAuthUserInfo, provider string) (*User, *OAuthAccount, error) {
	// 1. 首先检查是否已经存在OAuth账号关联
	existingOAuth, err := GetOAuthAccountByProvider(provider, oauthInfo.ProviderUserID)
	if err == nil && existingOAuth != nil {
		// OAuth账号已存在，获取关联的用户
		user, err := GetUserByID(int64(existingOAuth.UserID))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get user by oauth account: %w", err)
		}
		logrus.WithFields(logrus.Fields{
			"provider":        provider,
			"provider_userid": oauthInfo.ProviderUserID,
			"user_id":         user.ID,
		}).Info("Existing OAuth account found")
		return user, existingOAuth, nil
	}

	// 2. OAuth账号不存在，尝试通过邮箱查找现有用户
	var user *User
	if oauthInfo.Email != "" {
		existingUser, err := GetUserByEmail(oauthInfo.Email)
		if err == nil && existingUser != nil {
			// 找到现有用户，关联OAuth账号
			user = existingUser
			logrus.WithFields(logrus.Fields{
				"provider":        provider,
				"provider_userid": oauthInfo.ProviderUserID,
				"user_id":         user.ID,
				"email":           oauthInfo.Email,
			}).Info("Linking OAuth account to existing user by email")
		} else if err != nil && err != ErrUserNotFound {
			return nil, nil, fmt.Errorf("failed to check existing user by email: %w", err)
		}
	}

	// 3. 如果没有找到现有用户，创建新用户
	if user == nil {
		newUser, err := CreateUserFromOAuth(oauthInfo)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create user from oauth: %w", err)
		}
		user = newUser
		logrus.WithFields(logrus.Fields{
			"provider":        provider,
			"provider_userid": oauthInfo.ProviderUserID,
			"user_id":         user.ID,
			"username":        user.Username,
		}).Info("New user created from OAuth")
	}

	// 4. 创建OAuth账号关联
	oauthAccount := &OAuthAccount{
		UserID:         int(user.ID),
		Provider:       provider,
		ProviderUserID: oauthInfo.ProviderUserID,
		Email:          oauthInfo.Email,
		Username:       oauthInfo.Username,
		AvatarURL:      oauthInfo.AvatarURL,
	}

	err = CreateOAuthAccount(oauthAccount)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create oauth account: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"provider":        provider,
		"provider_userid": oauthInfo.ProviderUserID,
		"user_id":         user.ID,
	}).Info("OAuth account linked successfully")

	return user, oauthAccount, nil
}

// generateUniqueUsername 生成唯一的用户名
func generateUniqueUsername(oauthInfo *OAuthUserInfo) string {
	// 优先使用OAuth提供的用户名
	if oauthInfo.Username != "" {
		return oauthInfo.Username
	}
	
	// 如果没有用户名，尝试从邮箱提取
	if oauthInfo.Email != "" {
		// 提取邮箱@符号前的部分作为用户名
		if atIndex := strings.Index(oauthInfo.Email, "@"); atIndex > 0 {
			username := oauthInfo.Email[:atIndex]
			// 清理用户名，只保留字母数字和下划线
			username = strings.ReplaceAll(username, ".", "_")
			username = strings.ReplaceAll(username, "-", "_")
			username = strings.ReplaceAll(username, "+", "_")
			if len(username) >= 3 {
				return username
			}
		}
	}
	
	// 如果都不可用，生成基于时间戳的用户名
	return fmt.Sprintf("user_%d", time.Now().Unix())
}

// CreateUserFromOAuth 从OAuth信息创建新用户
func CreateUserFromOAuth(oauthInfo *OAuthUserInfo) (*User, error) {
	// 生成唯一的用户名
	username := generateUniqueUsername(oauthInfo)

	// 检查用户名是否已存在，如果存在则添加随机后缀
	originalUsername := username
	suffix := 1
	for {
		_, err := GetUserByUsername(username)
		if err == ErrUserNotFound {
			// 用户名可用
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to check username availability: %w", err)
		}
		// 用户名已存在，尝试添加后缀
		username = fmt.Sprintf("%s_%d", originalUsername, suffix)
		suffix++
		
		// 防止无限循环，最多尝试100次
		if suffix > 100 {
			return nil, fmt.Errorf("failed to generate unique username after 100 attempts")
		}
	}

	// 创建用户（OAuth用户不需要密码）
	// 使用随机密码哈希，因为OAuth用户不会使用密码登录
	randomPassword := fmt.Sprintf("oauth_%s_%d", oauthInfo.ProviderUserID, time.Now().Unix())
	user, err := CreateUser(username, oauthInfo.Email, randomPassword, "user")
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// LinkOAuthAccountToUser 将OAuth账号关联到现有用户
func LinkOAuthAccountToUser(userID int, oauthInfo *OAuthUserInfo, provider string) (*OAuthAccount, error) {
	// 检查是否已经存在该OAuth账号的关联
	existingOAuth, err := GetOAuthAccountByProvider(provider, oauthInfo.ProviderUserID)
	if err == nil && existingOAuth != nil {
		// OAuth账号已经关联到其他用户
		if existingOAuth.UserID != userID {
			return nil, fmt.Errorf("oauth account already linked to another user")
		}
		// 已经关联到当前用户，返回现有关联
		return existingOAuth, nil
	}

	// 创建新的OAuth账号关联
	oauthAccount := &OAuthAccount{
		UserID:         userID,
		Provider:       provider,
		ProviderUserID: oauthInfo.ProviderUserID,
		Email:          oauthInfo.Email,
		Username:       oauthInfo.Username,
		AvatarURL:      oauthInfo.AvatarURL,
	}

	err = CreateOAuthAccount(oauthAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to create oauth account: %w", err)
	}

	return oauthAccount, nil
}

// CheckEmailConflict 检查邮箱是否已被其他用户使用
func CheckEmailConflict(email string, excludeUserID int) (bool, error) {
	if email == "" {
		return false, nil
	}

	user, err := GetUserByEmail(email)
	if err == ErrUserNotFound {
		// 邮箱未被使用
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check email: %w", err)
	}

	// 如果邮箱属于当前用户，不算冲突
	if int(user.ID) == excludeUserID {
		return false, nil
	}

	// 邮箱已被其他用户使用
	return true, nil
}
