package middleware

import (
	"context"
	"Curry2API-go/database"
	"Curry2API-go/models"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// CursorSessionManager Cursor 账号 session 管理器（线程安全）
type CursorSessionManager struct {
	mu            sync.RWMutex
	sessions      map[string]*CursorSessionInfo
	currentIndex  int
	validSessions []*CursorSessionInfo
}

// CursorSessionInfo 与数据库结构保持一致
type CursorSessionInfo = models.CursorSessionInfo

var (
	cursorSessionManager     *CursorSessionManager
	cursorSessionManagerOnce sync.Once
)

// GetCursorSessionManager 获取 Cursor Session 管理器单例
func GetCursorSessionManager() *CursorSessionManager {
	cursorSessionManagerOnce.Do(func() {
		cursorSessionManager = &CursorSessionManager{
			sessions:      make(map[string]*CursorSessionInfo),
			validSessions: make([]*CursorSessionInfo, 0),
		}

		// 优先从数据库加载
		if err := cursorSessionManager.loadSessionsFromDB(); err != nil {
			logrus.Errorf("Failed to load sessions from database: %v", err)
			// 回退到环境变量
			cursorSessionManager.loadSessionsFromEnv()
		} else if len(cursorSessionManager.sessions) == 0 {
			cursorSessionManager.loadSessionsFromEnv()
		}
		go cursorSessionManager.startHealthChecker()
	})
	return cursorSessionManager
}

// loadSessionsFromEnv 从环境变量加载初始 sessions（作为回退方案）
func (csm *CursorSessionManager) loadSessionsFromEnv() {
	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)

	if sessionsStr := os.Getenv("CURSOR_SESSIONS"); sessionsStr != "" {
		sessions := strings.Split(sessionsStr, ",")
		for i, s := range sessions {
			if trimmed := strings.TrimSpace(s); trimmed != "" {
				email := fmt.Sprintf("account-%d@cursor.com", i+1)
				csm.sessions[email] = &CursorSessionInfo{
					Token:        trimmed,
					Email:        email,
					CreatedAt:    now,
					ExpiresAt:    expiresAt,
					IsValid:      true,
					UserAgent:    getDefaultUserAgent(),
					ExtraCookies: nil,
				}
			}
		}
		if len(csm.sessions) > 0 {
			csm.rebuildValidSessions()
			logrus.Infof("Loaded %d Cursor sessions from environment variables", len(csm.sessions))
			return
		}
	}

	if session := os.Getenv("CURSOR_SESSION"); session != "" {
		email := "default@cursor.com"
		csm.sessions[email] = &CursorSessionInfo{
			Token:     session,
			Email:     email,
			CreatedAt: now,
			ExpiresAt: expiresAt,
			IsValid:   true,
			UserAgent: getDefaultUserAgent(),
		}
		csm.rebuildValidSessions()
		logrus.Info("Loaded 1 Cursor session from CURSOR_SESSION")
		return
	}

	logrus.Warn("No Cursor sessions configured. Service will use x-is-human fallback only.")
}

// loadSessionsFromDB 从数据库加载 session 数据
func (csm *CursorSessionManager) loadSessionsFromDB() error {
	sessions, err := database.ListCursorSessions()
	if err != nil {
		return err
	}

	csm.mu.Lock()
	defer csm.mu.Unlock()

	csm.sessions = make(map[string]*CursorSessionInfo, len(sessions))
	csm.validSessions = make([]*CursorSessionInfo, 0, len(sessions))

	for _, session := range sessions {
		if session == nil {
			continue
		}
		if session.UserAgent == "" {
			session.UserAgent = getDefaultUserAgent()
		}

		copySession := &CursorSessionInfo{
			Token:        session.Token,
			Email:        session.Email,
			CreatedAt:    session.CreatedAt,
			LastUsed:     session.LastUsed,
			LastCheck:    session.LastCheck,
			ExpiresAt:    session.ExpiresAt,
			IsValid:      session.IsValid,
			UsageCount:   session.UsageCount,
			FailCount:    session.FailCount,
			UserAgent:    session.UserAgent,
			ExtraCookies: session.ExtraCookies,
		}

		csm.sessions[copySession.Email] = copySession
		if copySession.IsValid {
			csm.validSessions = append(csm.validSessions, copySession)
		}
	}

	logrus.Infof("Loaded %d Cursor sessions from database (%d valid)", len(csm.sessions), len(csm.validSessions))
	return nil
}

// getDefaultUserAgent 获取默认 User-Agent
func getDefaultUserAgent() string {
	return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36"
}

// rebuildValidSessions 重建有效 session 缓存
func (csm *CursorSessionManager) rebuildValidSessions() {
	csm.validSessions = make([]*CursorSessionInfo, 0, len(csm.sessions))
	now := time.Now()
	for _, session := range csm.sessions {
		if session.IsValid && now.Before(session.ExpiresAt) {
			csm.validSessions = append(csm.validSessions, session)
		}
	}
}

// HasValidSessions 是否存在有效 session
func (csm *CursorSessionManager) HasValidSessions() bool {
	csm.mu.RLock()
	defer csm.mu.RUnlock()
	return len(csm.validSessions) > 0
}

// GetValidSession 获取一个有效 session（轮询负载均衡）
func (csm *CursorSessionManager) GetValidSession() (*CursorSessionInfo, error) {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	if len(csm.validSessions) == 0 {
		return nil, fmt.Errorf("no valid Cursor sessions available")
	}

	session := csm.validSessions[csm.currentIndex%len(csm.validSessions)]
	csm.currentIndex = (csm.currentIndex + 1) % len(csm.validSessions)
	return session, nil
}

// MarkSessionFailed 标记 session 失败，并持久化状态
func (csm *CursorSessionManager) MarkSessionFailed(session *CursorSessionInfo) {
	if session == nil {
		return
	}

	csm.mu.Lock()
	session.FailCount++
	if session.FailCount >= 3 {
		session.IsValid = false
		csm.rebuildValidSessions()
		logrus.Errorf("Session %s marked as invalid after %d consecutive failures", session.Email, session.FailCount)
	} else {
		logrus.Warnf("Session %s failed (count: %d)", session.Email, session.FailCount)
	}
	failCount := session.FailCount
	isValid := session.IsValid
	email := session.Email
	csm.mu.Unlock()

	// 异步更新数据库
	go func() {
		if err := database.UpdateSessionStatus(email, isValid, failCount); err != nil {
			logrus.Warnf("Failed to update session status in database: %v", err)
		}
	}()
}

// MarkSessionSuccess 标记 session 成功，并更新统计
func (csm *CursorSessionManager) MarkSessionSuccess(session *CursorSessionInfo) {
	if session == nil {
		return
	}

	csm.mu.Lock()
	session.LastUsed = time.Now()
	session.FailCount = 0
	session.IsValid = true
	session.UsageCount++
	email := session.Email
	csm.rebuildValidSessions()
	csm.mu.Unlock()

	// 异步更新数据库
	go func() {
		if err := database.UpdateSessionUsage(email); err != nil {
			logrus.Warnf("Failed to update session usage in database: %v", err)
		}
	}()

	go func() {
		if err := database.UpdateSessionStatus(email, true, 0); err != nil {
			logrus.Warnf("Failed to reset session status in database: %v", err)
		}
	}()
}

// ValidateSession 验证 session 是否有效
func (csm *CursorSessionManager) ValidateSession(ctx context.Context, session *CursorSessionInfo) bool {
	if session == nil {
		return false
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://cursor.com/api/user", nil)
	if err != nil {
		logrus.Debugf("Failed to create validation request: %v", err)
		return csm.updateCheckResult(session, false)
	}

	req.Header.Set("User-Agent", session.UserAgent)
	req.Header.Set("Cookie", fmt.Sprintf("cursor_session=%s", session.Token))

	if len(session.ExtraCookies) > 0 {
		var cookies []string
		for name, value := range session.ExtraCookies {
			cookies = append(cookies, fmt.Sprintf("%s=%s", name, value))
		}
		req.Header.Add("Cookie", strings.Join(cookies, "; "))
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logrus.Debugf("Session validation request failed: %v", err)
		return csm.updateCheckResult(session, false)
	}
	defer resp.Body.Close()

	var result bool
	switch resp.StatusCode {
	case http.StatusOK, http.StatusNotFound:
		result = csm.updateCheckResult(session, true)
	case http.StatusUnauthorized, http.StatusForbidden:
		logrus.Warnf("Session %s validation failed with status %d", session.Email, resp.StatusCode)
		result = csm.updateCheckResult(session, false)
	default:
		// 其他状态暂视为有效，可能是临时错误
		result = csm.updateCheckResult(session, true)
	}

	// 异步更新数据库
	lastCheck := session.LastCheck
	isValid := session.IsValid
	email := session.Email
	go func() {
		if err := database.UpdateSessionCheck(email, lastCheck, isValid); err != nil {
			logrus.Warnf("Failed to update session check in database: %v", err)
		}
	}()

	return result
}

// updateCheckResult 更新最后一次检查结果并同步数据库
func (csm *CursorSessionManager) updateCheckResult(session *CursorSessionInfo, isValid bool) bool {
	now := time.Now()
	session.LastCheck = now
	session.IsValid = isValid
	return isValid
}

// startHealthChecker 启动后台健康检查
func (csm *CursorSessionManager) startHealthChecker() {
	healthTicker := time.NewTicker(30 * time.Minute)
	cleanupTicker := time.NewTicker(24 * time.Hour) // 每24小时清理一次过期 session
	defer healthTicker.Stop()
	defer cleanupTicker.Stop()

	logrus.Info("Cursor session health checker started")
	
	// 注意：不在启动时执行清理，避免误删数据
	
	for {
		select {
		case <-healthTicker.C:
			csm.performHealthCheck()
		case <-cleanupTicker.C:
			csm.cleanupExpiredSessions()
		}
	}
}

// cleanupExpiredSessions 清理过期的 sessions
func (csm *CursorSessionManager) cleanupExpiredSessions() {
	// 先获取过期的 sessions 用于日志记录
	expiredSessions, err := database.GetExpiredSessions()
	if err != nil {
		logrus.Warnf("Failed to get expired sessions: %v", err)
	} else if len(expiredSessions) > 0 {
		for _, session := range expiredSessions {
			logrus.Infof("Cleaning up expired session: %s (expired at: %s)", session.Email, session.ExpiresAt.Format(time.RFC3339))
		}
	}
	
	// 从数据库删除过期 sessions
	deleted, err := database.CleanupExpiredSessions()
	if err != nil {
		logrus.Errorf("Failed to cleanup expired sessions from database: %v", err)
		return
	}
	
	if deleted > 0 {
		logrus.Infof("Cleaned up %d expired Cursor sessions from database", deleted)
		
		// 从内存中移除过期 sessions
		csm.mu.Lock()
		now := time.Now()
		for email, session := range csm.sessions {
			if now.After(session.ExpiresAt) {
				delete(csm.sessions, email)
				logrus.Debugf("Removed expired session from memory: %s", email)
			}
		}
		csm.rebuildValidSessions()
		csm.mu.Unlock()
		
		logrus.Infof("Cursor session cleanup completed: %d sessions removed", deleted)
	}
}

// performHealthCheck 执行健康检查
func (csm *CursorSessionManager) performHealthCheck() {
	csm.mu.Lock()
	sessionsCopy := make([]*CursorSessionInfo, 0, len(csm.sessions))
	for _, session := range csm.sessions {
		sessionsCopy = append(sessionsCopy, session)
	}
	csm.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	validCount := 0
	for _, session := range sessionsCopy {
		if time.Since(session.LastCheck) < 15*time.Minute {
			if session.IsValid {
				validCount++
			}
			continue
		}

		isValid := csm.ValidateSession(ctx, session)

		csm.mu.Lock()
		session.IsValid = isValid
		if !isValid {
			session.FailCount++
		} else {
			session.FailCount = 0
			validCount++
		}
		csm.mu.Unlock()

		if err := database.UpdateSessionStatus(session.Email, session.IsValid, session.FailCount); err != nil {
			logrus.Debugf("Failed to update session status for %s: %v", session.Email, err)
		}

		logrus.Debugf("Session %s health check: valid=%v", session.Email, isValid)
	}

	csm.mu.Lock()
	csm.rebuildValidSessions()
	csm.mu.Unlock()

	logrus.Infof("Health check completed: %d/%d sessions valid", validCount, len(sessionsCopy))
}

// AddSession 添加新的 session
func (csm *CursorSessionManager) AddSession(email, token string, expiresAt time.Time, extraCookies map[string]string) error {
	if email == "" || token == "" {
		return fmt.Errorf("email and token cannot be empty")
	}

	csm.mu.Lock()
	if _, exists := csm.sessions[email]; exists {
		csm.mu.Unlock()
		return fmt.Errorf("session already exists for email: %s", email)
	}
	csm.mu.Unlock()

	var cookiesCopy map[string]string
	if len(extraCookies) > 0 {
		cookiesCopy = make(map[string]string, len(extraCookies))
		for k, v := range extraCookies {
			cookiesCopy[k] = v
		}
	}

	// 写数据库
	if err := database.AddCursorSession(email, token, "", expiresAt, cookiesCopy); err != nil {
		return fmt.Errorf("failed to save session to database: %w", err)
	}

	// 更新内存
	csm.mu.Lock()
	defer csm.mu.Unlock()

	session := &CursorSessionInfo{
		Token:        token,
		Email:        email,
		CreatedAt:    time.Now(),
		ExpiresAt:    expiresAt,
		IsValid:      true,
		ExtraCookies: cookiesCopy,
		UserAgent:    getDefaultUserAgent(),
		UsageCount:   0,
		FailCount:    0,
	}
	csm.sessions[email] = session
	csm.rebuildValidSessions()

	logrus.Infof("Added Cursor session: %s", email)
	return nil
}

// RemoveSession 删除 session
func (csm *CursorSessionManager) RemoveSession(email string) error {
	// 从数据库删除
	if err := database.RemoveCursorSession(email); err != nil {
		return fmt.Errorf("failed to remove session from database: %w", err)
	}

	// 从内存删除
	csm.mu.Lock()
	defer csm.mu.Unlock()

	if _, exists := csm.sessions[email]; !exists {
		return fmt.Errorf("session not found: %s", email)
	}

	delete(csm.sessions, email)
	csm.rebuildValidSessions()

	logrus.Infof("Removed Cursor session: %s", email)
	return nil
}

// ListSessions 列出所有 session（提供安全副本）
func (csm *CursorSessionManager) ListSessions() []*CursorSessionInfo {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	result := make([]*CursorSessionInfo, 0, len(csm.sessions))
	for _, session := range csm.sessions {
		copySession := *session
		copySession.Token = maskToken(session.Token)
		copySession.ExtraCookies = nil
		result = append(result, &copySession)
	}

	return result
}

// ReloadFromDB 从数据库重新加载所有 sessions
func (csm *CursorSessionManager) ReloadFromDB() error {
	logrus.Info("Reloading Cursor sessions from database...")
	
	if err := csm.loadSessionsFromDB(); err != nil {
		logrus.Errorf("Failed to reload sessions: %v", err)
		return err
	}
	
	logrus.Infof("Successfully reloaded %d sessions from database", len(csm.sessions))
	return nil
}

// GetStats 获取统计信息
func (csm *CursorSessionManager) GetStats() map[string]interface{} {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	totalUsage := int64(0)
	for _, session := range csm.sessions {
		totalUsage += session.UsageCount
	}

	return map[string]interface{}{
		"total_sessions":  len(csm.sessions),
		"valid_sessions":  len(csm.validSessions),
		"total_usage":     totalUsage,
		"current_index":   csm.currentIndex,
		"fallback_active": len(csm.validSessions) == 0,
	}
}

// maskToken 掩码 token（保留前8后4）
func maskToken(token string) string {
	tokenLen := len(token)
	if tokenLen <= 12 {
		return strings.Repeat("*", tokenLen)
	}
	return token[:8] + strings.Repeat("*", tokenLen-12) + token[tokenLen-4:]
}


// MigrateEncryptSessions 迁移现有明文数据到加密格式
func (csm *CursorSessionManager) MigrateEncryptSessions() (int, error) {
	logrus.Info("Starting cursor session encryption migration...")
	
	migratedCount, err := database.MigrateEncryptCursorSessions()
	if err != nil {
		return 0, err
	}
	
	// 重新加载数据以确保内存中的数据是最新的
	if migratedCount > 0 {
		if err := csm.loadSessionsFromDB(); err != nil {
			logrus.Warnf("Failed to reload sessions after migration: %v", err)
		}
	}
	
	logrus.Infof("Cursor session encryption migration completed: %d sessions migrated", migratedCount)
	return migratedCount, nil
}
