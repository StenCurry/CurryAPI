package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// OAuthCrypto OAuth 加密工具
type OAuthCrypto struct {
	key []byte
}

// NewOAuthCrypto 创建 OAuth 加密工具
// 从环境变量 OAUTH_ENCRYPTION_KEY 读取加密密钥
// 如果未设置，将生成一个新密钥（仅用于开发环境）
func NewOAuthCrypto() (*OAuthCrypto, error) {
	keyStr := os.Getenv("OAUTH_ENCRYPTION_KEY")
	
	var key []byte
	var err error
	
	if keyStr == "" {
		logrus.Warn("OAUTH_ENCRYPTION_KEY not set, generating a temporary key (NOT for production)")
		// 生成一个临时密钥（仅用于开发）
		key = make([]byte, 32) // AES-256
		if _, err := rand.Read(key); err != nil {
			return nil, fmt.Errorf("failed to generate encryption key: %w", err)
		}
	} else {
		// 从 base64 解码密钥
		key, err = base64.StdEncoding.DecodeString(keyStr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode encryption key: %w", err)
		}
		
		// 验证密钥长度（AES-256 需要 32 字节）
		if len(key) != 32 {
			return nil, fmt.Errorf("invalid encryption key length: expected 32 bytes, got %d", len(key))
		}
	}
	
	return &OAuthCrypto{key: key}, nil
}

// EncryptToken 加密 token
func (c *OAuthCrypto) EncryptToken(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	
	// 创建 AES cipher
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}
	
	// 创建 GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}
	
	// 生成随机 nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	
	// 加密数据
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	
	// 返回 base64 编码的密文
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptToken 解密 token
func (c *OAuthCrypto) DecryptToken(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	
	// 解码 base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}
	
	// 创建 AES cipher
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}
	
	// 创建 GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}
	
	// 验证数据长度
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	
	// 提取 nonce 和密文
	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	
	// 解密数据
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}
	
	return string(plaintext), nil
}

// EncryptAccessToken 加密 access token
func (c *OAuthCrypto) EncryptAccessToken(token string) (string, error) {
	return c.EncryptToken(token)
}

// DecryptAccessToken 解密 access token
func (c *OAuthCrypto) DecryptAccessToken(encryptedToken string) (string, error) {
	return c.DecryptToken(encryptedToken)
}

// EncryptRefreshToken 加密 refresh token
func (c *OAuthCrypto) EncryptRefreshToken(token string) (string, error) {
	return c.EncryptToken(token)
}

// DecryptRefreshToken 解密 refresh token
func (c *OAuthCrypto) DecryptRefreshToken(encryptedToken string) (string, error) {
	return c.DecryptToken(encryptedToken)
}

// GenerateEncryptionKey 生成新的加密密钥（用于初始化）
// 返回 base64 编码的密钥，可以设置为 OAUTH_ENCRYPTION_KEY 环境变量
func GenerateEncryptionKey() (string, error) {
	key := make([]byte, 32) // AES-256
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
