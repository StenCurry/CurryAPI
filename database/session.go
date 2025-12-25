package database

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
)

// Session 会话模型
type Session struct {
	ID        string    `json:"id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// CreateSession 创建新会话
func CreateSession(userID int64, username, role, ipAddress, userAgent string, duration time.Duration) (*Session, error) {
	sessionID := uuid.New().String()
	now := time.Now()
	expiresAt := now.Add(duration)
	
	_, err := db.Exec(
		`INSERT INTO sessions (id, user_id, username, role, ip_address, user_agent, created_at, expires_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		sessionID, userID, username, role, ipAddress, userAgent, now, expiresAt,
	)
	if err != nil {
		return nil, err
	}
	
	return &Session{
		ID:        sessionID,
		UserID:    userID,
		Username:  username,
		Role:      role,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}, nil
}

// GetSession 获取会话
func GetSession(sessionID string) (*Session, error) {
	session := &Session{}
	err := db.QueryRow(
		`SELECT id, user_id, username, role, ip_address, user_agent, created_at, expires_at 
		 FROM sessions WHERE id = ?`,
		sessionID,
	).Scan(&session.ID, &session.UserID, &session.Username, &session.Role, 
		&session.IPAddress, &session.UserAgent, &session.CreatedAt, &session.ExpiresAt)
	
	if err == sql.ErrNoRows {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	
	// 检查是否过期
	if time.Now().After(session.ExpiresAt) {
		// 删除过期会话
		_ = DeleteSession(sessionID)
		return nil, ErrSessionExpired
	}
	
	// 自动续期：如果会话剩余时间少于12小时，自动延长到24小时
	remainingTime := time.Until(session.ExpiresAt)
	if remainingTime < 12*time.Hour {
		newExpiresAt := time.Now().Add(24 * time.Hour)
		_ = ExtendSession(sessionID, newExpiresAt)
		session.ExpiresAt = newExpiresAt
		logrus.Debugf("Session %s auto-extended to %v", sessionID[:8]+"...", newExpiresAt)
	}
	
	return session, nil
}

// ExtendSession 延长会话有效期
func ExtendSession(sessionID string, newExpiresAt time.Time) error {
	_, err := db.Exec(`UPDATE sessions SET expires_at = ? WHERE id = ?`, newExpiresAt, sessionID)
	return err
}

// DeleteSession 删除会话
func DeleteSession(sessionID string) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE id = ?`, sessionID)
	return err
}

// DeleteUserSessions 删除用户的所有会话
func DeleteUserSessions(userID int64) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID)
	return err
}

// DeleteUserOldSessions 删除用户的旧会话（保留最新的N个）
func DeleteUserOldSessions(userID int64, keepCount int) error {
	// 获取用户的所有会话，按创建时间降序
	rows, err := db.Query(`
		SELECT id FROM sessions 
		WHERE user_id = ? AND expires_at > ? 
		ORDER BY created_at DESC
	`, userID, time.Now())
	if err != nil {
		return err
	}
	defer rows.Close()
	
	var sessionIDs []string
	for rows.Next() {
		var sessionID string
		if err := rows.Scan(&sessionID); err != nil {
			return err
		}
		sessionIDs = append(sessionIDs, sessionID)
	}
	
	// 如果会话数量超过保留数量，删除旧的
	if len(sessionIDs) > keepCount {
		for i := keepCount; i < len(sessionIDs); i++ {
			_ = DeleteSession(sessionIDs[i])
		}
	}
	
	return nil
}

// CleanExpiredSessions 清理过期会话
func CleanExpiredSessions() error {
	result, err := db.Exec(`DELETE FROM sessions WHERE expires_at < ?`, time.Now())
	if err != nil {
		return err
	}
	
	rows, _ := result.RowsAffected()
	if rows > 0 {
		logrus.Infof("Cleaned up %d expired sessions", rows)
	}
	return nil
}
