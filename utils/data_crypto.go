package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// DataCrypto 通用数据加密工具
// 使用 AES-256-GCM 加密敏感数据
type DataCrypto struct {
	key []byte
}

var (
	dataCrypto     *DataCrypto
	dataCryptoOnce sync.Once
	dataCryptoErr  error
)

// InitDataCrypto 初始化数据加密工具
// 从环境变量 DATA_ENCRYPTION_KEY 读取加密密钥
func InitDataCrypto() error {
	dataCryptoOnce.Do(func() {
		keyStr := os.Getenv("DATA_ENCRYPTION_KEY")

		var key []byte

		if keyStr == "" {
			logrus.Warn("DATA_ENCRYPTION_KEY not set, generating a temporary key (NOT for production)")
			// 生成一个临时密钥（仅用于开发）
			key = make([]byte, 32) // AES-256
			if _, err := rand.Read(key); err != nil {
				dataCryptoErr = fmt.Errorf("failed to generate encryption key: %w", err)
				return
			}
			// 输出生成的密钥，方便开发者设置
			logrus.Warnf("Generated temporary DATA_ENCRYPTION_KEY: %s", base64.StdEncoding.EncodeToString(key))
		} else {
			// 从 base64 解码密钥
			var err error
			key, err = base64.StdEncoding.DecodeString(keyStr)
			if err != nil {
				dataCryptoErr = fmt.Errorf("failed to decode encryption key: %w", err)
				return
			}

			// 验证密钥长度（AES-256 需要 32 字节）
			if len(key) != 32 {
				dataCryptoErr = fmt.Errorf("invalid encryption key length: expected 32 bytes, got %d", len(key))
				return
			}
		}

		dataCrypto = &DataCrypto{key: key}
		logrus.Info("Data encryption initialized successfully")
	})

	return dataCryptoErr
}

// GetDataCrypto 获取数据加密工具实例
func GetDataCrypto() *DataCrypto {
	return dataCrypto
}

// Encrypt 加密数据
func (c *DataCrypto) Encrypt(plaintext string) (string, error) {
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

	// 返回带前缀的 base64 编码密文，用于识别加密数据
	return "ENC:" + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密数据
func (c *DataCrypto) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// 检查是否是加密数据（带 ENC: 前缀）
	if !strings.HasPrefix(ciphertext, "ENC:") {
		// 不是加密数据，直接返回原文（兼容旧数据）
		return ciphertext, nil
	}

	// 移除前缀
	ciphertext = strings.TrimPrefix(ciphertext, "ENC:")

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

// IsEncrypted 检查数据是否已加密
func IsEncrypted(data string) bool {
	return strings.HasPrefix(data, "ENC:")
}

// EncryptSensitiveData 加密敏感数据（便捷函数）
func EncryptSensitiveData(plaintext string) (string, error) {
	if dataCrypto == nil {
		return plaintext, fmt.Errorf("data crypto not initialized")
	}
	return dataCrypto.Encrypt(plaintext)
}

// DecryptSensitiveData 解密敏感数据（便捷函数）
func DecryptSensitiveData(ciphertext string) (string, error) {
	if dataCrypto == nil {
		// 如果加密未初始化，返回原文（兼容模式）
		return ciphertext, nil
	}
	return dataCrypto.Decrypt(ciphertext)
}

// GenerateDataEncryptionKey 生成新的数据加密密钥
// 返回 base64 编码的密钥，可以设置为 DATA_ENCRYPTION_KEY 环境变量
func GenerateDataEncryptionKey() (string, error) {
	key := make([]byte, 32) // AES-256
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
