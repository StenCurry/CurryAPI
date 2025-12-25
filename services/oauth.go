package services

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// OAuthConfig OAuth 配置结构
type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURL  string
	StateExpiry        int // State 过期时间（秒）
}

// OAuthService OAuth 服务
type OAuthService struct {
	config *OAuthConfig
}

// OAuthToken OAuth 令牌
type OAuthToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in,omitempty"`
	ExpiresAt    time.Time `json:"-"`
}

// OAuthUserInfo OAuth 用户信息
type OAuthUserInfo struct {
	ProviderUserID string
	Email          string
	Username       string
	AvatarURL      string
	EmailVerified  bool
}

// OAuthError OAuth 错误
type OAuthError struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Provider string `json:"provider"`
}

func (e *OAuthError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Provider, e.Code, e.Message)
}

// NewOAuthService 创建 OAuth 服务
func NewOAuthService(config *OAuthConfig) *OAuthService {
	return &OAuthService{
		config: config,
	}
}

// LoadOAuthConfig 从环境变量加载 OAuth 配置
func LoadOAuthConfig() (*OAuthConfig, error) {
	config := &OAuthConfig{
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", ""),
		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GitHubRedirectURL:  getEnv("GITHUB_REDIRECT_URL", ""),
		StateExpiry:        getEnvAsInt("OAUTH_STATE_EXPIRY", 600), // 默认 10 分钟
	}

	// 验证配置
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("oauth config validation failed: %w", err)
	}

	return config, nil
}

// validate 验证 OAuth 配置
func (c *OAuthConfig) validate() error {
	// 至少需要配置一个 OAuth 提供商
	hasGoogle := c.GoogleClientID != "" && c.GoogleClientSecret != "" && c.GoogleRedirectURL != ""
	hasGitHub := c.GitHubClientID != "" && c.GitHubClientSecret != "" && c.GitHubRedirectURL != ""

	if !hasGoogle && !hasGitHub {
		logrus.Warn("No OAuth providers configured. OAuth login will not be available.")
	}

	if c.StateExpiry <= 0 {
		return fmt.Errorf("state expiry must be positive")
	}

	return nil
}

// GenerateState 生成随机 state 参数
func (s *OAuthService) GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random state: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// StoreState 存储 state 到数据库
func (s *OAuthService) StoreState(state, provider string) error {
	expiresAt := time.Now().Add(time.Duration(s.config.StateExpiry) * time.Second)
	
	// This will be implemented by the database layer
	// For now, we'll define the interface
	return storeOAuthState(state, provider, expiresAt)
}

// VerifyState 验证 state 参数
func (s *OAuthService) VerifyState(state, provider string) (bool, error) {
	// This will be implemented by the database layer
	return verifyOAuthState(state, provider)
}

// DeleteState 删除已使用的 state
func (s *OAuthService) DeleteState(state string) error {
	// This will be implemented by the database layer
	return deleteOAuthState(state)
}

// CleanupExpiredStates 清理过期的 state
func (s *OAuthService) CleanupExpiredStates() error {
	// This will be implemented by the database layer
	return cleanupExpiredOAuthStates()
}

// StartStateCleanupTask 启动定期清理过期 state 的任务
func (s *OAuthService) StartStateCleanupTask() {
	ticker := time.NewTicker(1 * time.Hour) // 每小时清理一次
	go func() {
		for range ticker.C {
			if err := s.CleanupExpiredStates(); err != nil {
				logrus.Errorf("Failed to cleanup expired OAuth states: %v", err)
			}
		}
	}()
	logrus.Info("OAuth state cleanup task started")
}

// GetAuthorizationURL 获取授权 URL
func (s *OAuthService) GetAuthorizationURL(provider, state string) (string, error) {
	switch provider {
	case "google":
		return s.getGoogleAuthURL(state)
	case "github":
		return s.getGitHubAuthURL(state)
	default:
		return "", &OAuthError{
			Code:     "invalid_provider",
			Message:  fmt.Sprintf("unsupported provider: %s", provider),
			Provider: provider,
		}
	}
}

// ExchangeCode 交换授权码获取访问令牌
func (s *OAuthService) ExchangeCode(provider, code string) (*OAuthToken, error) {
	switch provider {
	case "google":
		return s.exchangeGoogleCode(code)
	case "github":
		return s.exchangeGitHubCode(code)
	default:
		return nil, &OAuthError{
			Code:     "invalid_provider",
			Message:  fmt.Sprintf("unsupported provider: %s", provider),
			Provider: provider,
		}
	}
}

// GetUserInfo 获取用户信息
func (s *OAuthService) GetUserInfo(provider string, token *OAuthToken) (*OAuthUserInfo, error) {
	switch provider {
	case "google":
		return s.getGoogleUserInfo(token)
	case "github":
		return s.getGitHubUserInfo(token)
	default:
		return nil, &OAuthError{
			Code:     "invalid_provider",
			Message:  fmt.Sprintf("unsupported provider: %s", provider),
			Provider: provider,
		}
	}
}

// getGoogleAuthURL 生成 Google 授权 URL
func (s *OAuthService) getGoogleAuthURL(state string) (string, error) {
	if s.config.GoogleClientID == "" {
		return "", &OAuthError{
			Code:     "config_error",
			Message:  "Google OAuth not configured",
			Provider: "google",
		}
	}

	params := url.Values{}
	params.Add("client_id", s.config.GoogleClientID)
	params.Add("redirect_uri", s.config.GoogleRedirectURL)
	params.Add("response_type", "code")
	params.Add("scope", "openid email profile")
	params.Add("state", state)
	params.Add("access_type", "offline")
	params.Add("prompt", "consent")

	authURL := "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
	return authURL, nil
}

// exchangeGoogleCode 交换 Google 授权码
func (s *OAuthService) exchangeGoogleCode(code string) (*OAuthToken, error) {
	tokenURL := "https://oauth2.googleapis.com/token"

	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", s.config.GoogleClientID)
	data.Set("client_secret", s.config.GoogleClientSecret)
	data.Set("redirect_uri", s.config.GoogleRedirectURL)
	data.Set("grant_type", "authorization_code")

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, &OAuthError{
			Code:     "network_error",
			Message:  fmt.Sprintf("failed to exchange code: %v", err),
			Provider: "google",
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &OAuthError{
			Code:     "read_error",
			Message:  fmt.Sprintf("failed to read response: %v", err),
			Provider: "google",
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &OAuthError{
			Code:     "exchange_failed",
			Message:  fmt.Sprintf("token exchange failed: %s", string(body)),
			Provider: "google",
		}
	}

	var token OAuthToken
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, &OAuthError{
			Code:     "parse_error",
			Message:  fmt.Sprintf("failed to parse token response: %v", err),
			Provider: "google",
		}
	}

	// 计算过期时间
	if token.ExpiresIn > 0 {
		token.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	}

	return &token, nil
}

// getGoogleUserInfo 获取 Google 用户信息
func (s *OAuthService) getGoogleUserInfo(token *OAuthToken) (*OAuthUserInfo, error) {
	userInfoURL := "https://www.googleapis.com/oauth2/v2/userinfo"

	req, err := http.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		return nil, &OAuthError{
			Code:     "request_error",
			Message:  fmt.Sprintf("failed to create request: %v", err),
			Provider: "google",
		}
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, &OAuthError{
			Code:     "network_error",
			Message:  fmt.Sprintf("failed to get user info: %v", err),
			Provider: "google",
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &OAuthError{
			Code:     "read_error",
			Message:  fmt.Sprintf("failed to read response: %v", err),
			Provider: "google",
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &OAuthError{
			Code:     "userinfo_failed",
			Message:  fmt.Sprintf("failed to get user info: %s", string(body)),
			Provider: "google",
		}
	}

	var googleUser struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}

	if err := json.Unmarshal(body, &googleUser); err != nil {
		return nil, &OAuthError{
			Code:     "parse_error",
			Message:  fmt.Sprintf("failed to parse user info: %v", err),
			Provider: "google",
		}
	}

	return &OAuthUserInfo{
		ProviderUserID: googleUser.ID,
		Email:          googleUser.Email,
		Username:       googleUser.Name,
		AvatarURL:      googleUser.Picture,
		EmailVerified:  googleUser.VerifiedEmail,
	}, nil
}

// getGitHubAuthURL 生成 GitHub 授权 URL
func (s *OAuthService) getGitHubAuthURL(state string) (string, error) {
	if s.config.GitHubClientID == "" {
		return "", &OAuthError{
			Code:     "config_error",
			Message:  "GitHub OAuth not configured",
			Provider: "github",
		}
	}

	params := url.Values{}
	params.Add("client_id", s.config.GitHubClientID)
	params.Add("redirect_uri", s.config.GitHubRedirectURL)
	params.Add("scope", "user:email")
	params.Add("state", state)

	authURL := "https://github.com/login/oauth/authorize?" + params.Encode()
	return authURL, nil
}

// exchangeGitHubCode 交换 GitHub 授权码
// 增加重试机制以应对网络不稳定的情况
func (s *OAuthService) exchangeGitHubCode(code string) (*OAuthToken, error) {
	tokenURL := "https://github.com/login/oauth/access_token"

	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", s.config.GitHubClientID)
	data.Set("client_secret", s.config.GitHubClientSecret)
	data.Set("redirect_uri", s.config.GitHubRedirectURL)

	// 创建带有更长超时时间的HTTP客户端
	client := &http.Client{
		Timeout: 60 * time.Second, // 增加到60秒
		Transport: &http.Transport{
			TLSHandshakeTimeout:   30 * time.Second, // TLS握手超时30秒
			ResponseHeaderTimeout: 30 * time.Second,
			IdleConnTimeout:       90 * time.Second,
		},
	}

	var lastErr error
	maxRetries := 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
		if err != nil {
			return nil, &OAuthError{
				Code:     "request_error",
				Message:  fmt.Sprintf("failed to create request: %v", err),
				Provider: "github",
			}
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")

		logrus.Debugf("GitHub OAuth code exchange attempt %d/%d", attempt, maxRetries)

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			logrus.Warnf("GitHub OAuth code exchange attempt %d failed: %v", attempt, err)
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * 2 * time.Second) // 指数退避
				continue
			}
			return nil, &OAuthError{
				Code:     "network_error",
				Message:  fmt.Sprintf("failed to exchange code after %d attempts: %v", maxRetries, err),
				Provider: "github",
			}
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, &OAuthError{
				Code:     "read_error",
				Message:  fmt.Sprintf("failed to read response: %v", err),
				Provider: "github",
			}
		}

		if resp.StatusCode != http.StatusOK {
			return nil, &OAuthError{
				Code:     "exchange_failed",
				Message:  fmt.Sprintf("token exchange failed: %s", string(body)),
				Provider: "github",
			}
		}

		var token OAuthToken
		if err := json.Unmarshal(body, &token); err != nil {
			return nil, &OAuthError{
				Code:     "parse_error",
				Message:  fmt.Sprintf("failed to parse token response: %v", err),
				Provider: "github",
			}
		}

		// GitHub 的 token 通常不会过期，但如果有 expires_in，计算过期时间
		if token.ExpiresIn > 0 {
			token.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
		}

		logrus.Infof("GitHub OAuth code exchange successful on attempt %d", attempt)
		return &token, nil
	}

	return nil, &OAuthError{
		Code:     "network_error",
		Message:  fmt.Sprintf("failed to exchange code after %d attempts: %v", maxRetries, lastErr),
		Provider: "github",
	}
}

// getGitHubUserInfo 获取 GitHub 用户信息
func (s *OAuthService) getGitHubUserInfo(token *OAuthToken) (*OAuthUserInfo, error) {
	// 获取用户基本信息
	userInfoURL := "https://api.github.com/user"

	req, err := http.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		return nil, &OAuthError{
			Code:     "request_error",
			Message:  fmt.Sprintf("failed to create request: %v", err),
			Provider: "github",
		}
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// 创建带有更长超时时间的HTTP客户端
	client := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			TLSHandshakeTimeout:   30 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			IdleConnTimeout:       90 * time.Second,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, &OAuthError{
			Code:     "network_error",
			Message:  fmt.Sprintf("failed to get user info: %v", err),
			Provider: "github",
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &OAuthError{
			Code:     "read_error",
			Message:  fmt.Sprintf("failed to read response: %v", err),
			Provider: "github",
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &OAuthError{
			Code:     "userinfo_failed",
			Message:  fmt.Sprintf("failed to get user info: %s", string(body)),
			Provider: "github",
		}
	}

	var githubUser struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.Unmarshal(body, &githubUser); err != nil {
		return nil, &OAuthError{
			Code:     "parse_error",
			Message:  fmt.Sprintf("failed to parse user info: %v", err),
			Provider: "github",
		}
	}

	// 如果基本信息中没有邮箱，获取邮箱列表
	email := githubUser.Email
	emailVerified := false
	if email == "" {
		email, emailVerified, err = s.getGitHubPrimaryEmail(token)
		if err != nil {
			logrus.Warnf("Failed to get GitHub email: %v", err)
		}
	} else {
		emailVerified = true // 如果在基本信息中有邮箱，认为已验证
	}

	username := githubUser.Name
	if username == "" {
		username = githubUser.Login
	}

	return &OAuthUserInfo{
		ProviderUserID: fmt.Sprintf("%d", githubUser.ID),
		Email:          email,
		Username:       username,
		AvatarURL:      githubUser.AvatarURL,
		EmailVerified:  emailVerified,
	}, nil
}

// getGitHubPrimaryEmail 获取 GitHub 主邮箱
func (s *OAuthService) getGitHubPrimaryEmail(token *OAuthToken) (string, bool, error) {
	emailURL := "https://api.github.com/user/emails"

	req, err := http.NewRequest("GET", emailURL, nil)
	if err != nil {
		return "", false, err
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// 创建带有更长超时时间的HTTP客户端
	client := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			TLSHandshakeTimeout:   30 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			IdleConnTimeout:       90 * time.Second,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", false, err
	}

	if resp.StatusCode != http.StatusOK {
		return "", false, fmt.Errorf("failed to get emails: %s", string(body))
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.Unmarshal(body, &emails); err != nil {
		return "", false, err
	}

	// 查找主邮箱
	for _, e := range emails {
		if e.Primary {
			return e.Email, e.Verified, nil
		}
	}

	// 如果没有主邮箱，返回第一个验证过的邮箱
	for _, e := range emails {
		if e.Verified {
			return e.Email, true, nil
		}
	}

	// 如果都没有，返回第一个邮箱
	if len(emails) > 0 {
		return emails[0].Email, emails[0].Verified, nil
	}

	return "", false, fmt.Errorf("no email found")
}

// Helper functions for environment variables

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为int
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		logrus.Warnf("Invalid integer value for %s: %s, using default: %d", key, valueStr, defaultValue)
		return defaultValue
	}

	return value
}

// Database interface functions
// These functions will be implemented by importing the database package

var (
	storeOAuthState          func(state, provider string, expiresAt time.Time) error
	verifyOAuthState         func(state, provider string) (bool, error)
	deleteOAuthState         func(state string) error
	cleanupExpiredOAuthStates func() error
)

// SetDatabaseFunctions 设置数据库函数（由 main 包调用）
func SetDatabaseFunctions(
	store func(state, provider string, expiresAt time.Time) error,
	verify func(state, provider string) (bool, error),
	delete func(state string) error,
	cleanup func() error,
) {
	storeOAuthState = store
	verifyOAuthState = verify
	deleteOAuthState = delete
	cleanupExpiredOAuthStates = cleanup
}
