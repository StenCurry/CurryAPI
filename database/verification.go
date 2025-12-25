package database

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

var (
	ErrCodeNotFound = errors.New("verification code not found")
	ErrCodeExpired  = errors.New("verification code expired")
	ErrCodeInvalid  = errors.New("verification code invalid")
)

const VerificationExpiry = 10 * time.Minute

// VerificationCode 验证码模型
type VerificationCode struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Code      string    `json:"code"`
	CodeType  string    `json:"code_type"`
	IPAddress string    `json:"ip_address"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
}

// CreateVerificationCode 创建验证码
func CreateVerificationCode(email, codeType, ipAddress string) (*VerificationCode, error) {
	// 生成6位数字验证码
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	now := time.Now()
	expiresAt := now.Add(VerificationExpiry)
	
	result, err := db.Exec(
		`INSERT INTO verification_codes (email, code, code_type, ip_address, created_at, expires_at, used) 
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		email, code, codeType, ipAddress, now, expiresAt, false,
	)
	if err != nil {
		return nil, err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	
	return &VerificationCode{
		ID:        id,
		Email:     email,
		Code:      code,
		CodeType:  codeType,
		IPAddress: ipAddress,
		CreatedAt: now,
		ExpiresAt: expiresAt,
		Used:      false,
	}, nil
}

// VerifyCode 验证验证码
func VerifyCode(email, code, codeType string) error {
	var vc VerificationCode
	err := db.QueryRow(
		`SELECT id, email, code, code_type, created_at, expires_at, used 
		 FROM verification_codes 
		 WHERE email = ? AND code_type = ? AND used = FALSE 
		 ORDER BY created_at DESC LIMIT 1`,
		email, codeType,
	).Scan(&vc.ID, &vc.Email, &vc.Code, &vc.CodeType, &vc.CreatedAt, &vc.ExpiresAt, &vc.Used)
	
	if err == sql.ErrNoRows {
		return ErrCodeNotFound
	}
	if err != nil {
		return err
	}
	
	// 检查是否过期
	if time.Now().After(vc.ExpiresAt) {
		return ErrCodeExpired
	}
	
	// 验证码是否匹配
	if vc.Code != code {
		return ErrCodeInvalid
	}
	
	// 标记为已使用
	_, err = db.Exec(`UPDATE verification_codes SET used = TRUE WHERE id = ?`, vc.ID)
	if err != nil {
		return err
	}
	
	return nil
}

// GetRecentCodeSentTime 获取最近发送验证码的时间
func GetRecentCodeSentTime(email, codeType string) (time.Time, error) {
	var createdAt time.Time
	err := db.QueryRow(
		`SELECT created_at FROM verification_codes 
		 WHERE email = ? AND code_type = ? 
		 ORDER BY created_at DESC LIMIT 1`,
		email, codeType,
	).Scan(&createdAt)
	
	if err == sql.ErrNoRows {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	
	return createdAt, nil
}

// InvalidateOldCodes 使旧验证码失效
func InvalidateOldCodes(email, codeType string) error {
	_, err := db.Exec(
		`UPDATE verification_codes SET used = TRUE 
		 WHERE email = ? AND code_type = ? AND used = FALSE`,
		email, codeType,
	)
	return err
}

// CleanExpiredCodes 清理过期验证码
func CleanExpiredCodes() error {
	_, err := db.Exec(`DELETE FROM verification_codes WHERE expires_at < ?`, time.Now())
	return err
}
