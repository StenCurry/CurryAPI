package middleware

import (
	"Curry2API-go/config"
	"Curry2API-go/database"
	"Curry2API-go/models"
	"Curry2API-go/utils"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// QuotaManager manages token quotas across all Cursor sessions
type QuotaManager struct {
	sessionMgr *CursorSessionManager
	config     *config.QuotaConfig
	mu         sync.RWMutex
}

var (
	quotaManager     *QuotaManager
	quotaManagerOnce sync.Once
)

// GetQuotaManager returns the singleton QuotaManager instance
func GetQuotaManager(cfg *config.QuotaConfig) *QuotaManager {
	quotaManagerOnce.Do(func() {
		quotaManager = &QuotaManager{
			sessionMgr: GetCursorSessionManager(),
			config:     cfg,
		}
		
		// Start background quota reset scheduler
		if cfg.Enabled {
			go quotaManager.startResetScheduler()
			
			// Check for sessions needing reset on startup
			go quotaManager.checkAndResetStale()
		}
		
		logrus.WithFields(logrus.Fields{
			"enabled":            cfg.Enabled,
			"default_free_quota": cfg.DefaultFreeQuota,
			"default_pro_quota":  cfg.DefaultProQuota,
			"low_threshold":      cfg.LowQuotaThreshold,
		}).Info("QuotaManager initialized")
	})
	return quotaManager
}

// IsEnabled returns whether quota management is enabled
func (qm *QuotaManager) IsEnabled() bool {
	return qm.config.Enabled
}

// startResetScheduler starts the daily quota reset scheduler
func (qm *QuotaManager) startResetScheduler() {
	for {
		now := time.Now().UTC()
		
		// Calculate next reset time (midnight UTC or configured hour)
		nextReset := time.Date(now.Year(), now.Month(), now.Day(), 
			qm.config.ResetHourUTC, 0, 0, 0, time.UTC)
		
		// If we've passed today's reset time, schedule for tomorrow
		if now.After(nextReset) {
			nextReset = nextReset.Add(24 * time.Hour)
		}
		
		// Wait until next reset time
		duration := time.Until(nextReset)
		logrus.WithFields(logrus.Fields{
			"next_reset": nextReset,
			"duration":   duration,
		}).Info("Quota reset scheduled")
		
		time.Sleep(duration)
		
		// Perform reset
		if err := qm.ResetDailyQuotas(); err != nil {
			logrus.WithError(err).Error("Failed to reset daily quotas")
		}
	}
}

// checkAndResetStale checks for sessions that need quota reset on startup
func (qm *QuotaManager) checkAndResetStale() {
	sessions, err := database.GetSessionsNeedingReset()
	if err != nil {
		logrus.WithError(err).Error("Failed to get sessions needing reset")
		return
	}
	
	if len(sessions) == 0 {
		logrus.Info("No sessions need quota reset on startup")
		return
	}
	
	logrus.WithField("count", len(sessions)).Info("Resetting stale session quotas on startup")
	
	for _, session := range sessions {
		if err := database.ResetSessionQuota(session.Email); err != nil {
			logrus.WithError(err).WithField("email", session.Email).Error("Failed to reset session quota")
		}
	}
	
	// Reload sessions from database
	if err := qm.sessionMgr.ReloadFromDB(); err != nil {
		logrus.WithError(err).Error("Failed to reload sessions after reset")
	}
}

// ResetDailyQuotas resets all session quotas (called at midnight UTC)
func (qm *QuotaManager) ResetDailyQuotas() error {
	logrus.Info("Starting daily quota reset for all sessions")
	
	if err := database.ResetAllSessionQuotas(); err != nil {
		return fmt.Errorf("failed to reset quotas in database: %w", err)
	}
	
	// Reload sessions from database to get updated quota values
	if err := qm.sessionMgr.ReloadFromDB(); err != nil {
		return fmt.Errorf("failed to reload sessions after reset: %w", err)
	}
	
	logrus.Info("Daily quota reset completed successfully")
	return nil
}

// TrackUsage updates token usage after API response
func (qm *QuotaManager) TrackUsage(session *CursorSessionInfo, usage models.Usage) error {
	if !qm.config.Enabled || session == nil {
		return nil
	}
	
	qm.mu.Lock()
	
	// Update daily usage counter synchronously
	session.DailyTokenUsed += int64(usage.TotalTokens)
	
	// Update quota status based on usage
	session.UpdateQuotaStatus(qm.config.LowQuotaThreshold)
	
	// Store values for async update
	email := session.Email
	tokensUsed := int64(usage.TotalTokens)
	quotaStatus := session.QuotaStatus
	
	qm.mu.Unlock()
	
	// Schedule async database write with retry
	go qm.persistUsageWithRetry(email, tokensUsed, quotaStatus)
	
	logrus.WithFields(logrus.Fields{
		"email":         email,
		"tokens_used":   tokensUsed,
		"total_used":    session.DailyTokenUsed,
		"remaining":     session.GetRemainingQuota(),
		"quota_status":  quotaStatus,
	}).Debug("Quota usage tracked")
	
	return nil
}

// persistUsageWithRetry persists usage to database with exponential backoff retry
func (qm *QuotaManager) persistUsageWithRetry(email string, tokensUsed int64, quotaStatus string) {
	maxRetries := qm.config.MaxRetries
	backoffMs := qm.config.RetryBackoffMs
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Update token usage
		if err := database.UpdateSessionQuotaUsage(email, tokensUsed); err != nil {
			if attempt < maxRetries-1 {
				// Wait with exponential backoff
				waitTime := time.Duration(backoffMs*(1<<attempt)) * time.Millisecond
				logrus.WithError(err).WithFields(logrus.Fields{
					"email":   email,
					"attempt": attempt + 1,
					"wait":    waitTime,
				}).Warn("Failed to update quota usage, retrying")
				time.Sleep(waitTime)
				continue
			} else {
				logrus.WithError(err).WithField("email", email).Error("Failed to update quota usage after all retries")
				return
			}
		}
		
		// Update quota status
		if err := database.UpdateSessionQuotaStatus(email, quotaStatus); err != nil {
			logrus.WithError(err).WithField("email", email).Warn("Failed to update quota status")
		}
		
		// Success
		return
	}
}

// SelectSessionForRequest chooses the best session based on quota and estimated usage
func (qm *QuotaManager) SelectSessionForRequest(estimatedTokens int) (*CursorSessionInfo, error) {
	if !qm.config.Enabled {
		// Quota management disabled, use original selection
		return qm.sessionMgr.GetValidSession()
	}
	
	qm.mu.RLock()
	sessions := qm.sessionMgr.ListSessions()
	qm.mu.RUnlock()
	
	if len(sessions) == 0 {
		return nil, fmt.Errorf("no sessions available")
	}
	
	// Filter and sort sessions by quota availability
	var availableSessions []*CursorSessionInfo
	var lowQuotaSessions []*CursorSessionInfo
	
	for _, session := range sessions {
		// Skip invalid or expired sessions
		if !session.IsValid || time.Now().After(session.ExpiresAt) {
			continue
		}
		
		// Check if session needs quota reset
		if session.NeedsQuotaReset() {
			// Reset this session's quota
			if err := database.ResetSessionQuota(session.Email); err != nil {
				logrus.WithError(err).WithField("email", session.Email).Warn("Failed to reset session quota")
				continue
			}
			// Reload session data
			updatedSession, err := database.GetCursorSession(session.Email)
			if err != nil {
				logrus.WithError(err).WithField("email", session.Email).Warn("Failed to reload session after reset")
				continue
			}
			session = updatedSession
		}
		
		// Categorize by quota status
		if session.IsSuitableForRequest(estimatedTokens) {
			availableSessions = append(availableSessions, session)
		} else if session.QuotaStatus != "exhausted" {
			lowQuotaSessions = append(lowQuotaSessions, session)
		}
	}
	
	// Priority 1: Sessions with sufficient quota
	if len(availableSessions) > 0 {
		return qm.selectBestSession(availableSessions), nil
	}
	
	// Priority 2: Sessions with low quota (best effort)
	if len(lowQuotaSessions) > 0 {
		logrus.Warn("All sessions have low quota, selecting best available")
		return qm.selectBestSession(lowQuotaSessions), nil
	}
	
	// All sessions exhausted, return error to trigger fallback
	return nil, fmt.Errorf("all sessions have exhausted their daily quota")
}

// selectBestSession selects the session with highest remaining quota percentage
func (qm *QuotaManager) selectBestSession(sessions []*CursorSessionInfo) *CursorSessionInfo {
	if len(sessions) == 0 {
		return nil
	}
	
	if len(sessions) == 1 {
		return sessions[0]
	}
	
	// Find session with highest remaining quota percentage
	bestSession := sessions[0]
	bestPercentage := 100.0 - bestSession.GetQuotaPercentageUsed()
	
	for _, session := range sessions[1:] {
		remainingPercentage := 100.0 - session.GetQuotaPercentageUsed()
		
		// If percentages are similar (within 5%), use round-robin
		if abs(remainingPercentage-bestPercentage) < 5.0 {
			// Simple round-robin: alternate between similar sessions
			continue
		}
		
		if remainingPercentage > bestPercentage {
			bestSession = session
			bestPercentage = remainingPercentage
		}
	}
	
	return bestSession
}

// abs returns absolute value of float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// EstimateRequestTokens estimates total tokens needed for a request
func (qm *QuotaManager) EstimateRequestTokens(messages []models.Message) int {
	return utils.EstimateTotalRequestTokens(messages, qm.config.EstimationMultiplier)
}

// GetQuotaStats returns current quota statistics for all sessions
func (qm *QuotaManager) GetQuotaStats() (*QuotaStatistics, error) {
	qm.mu.RLock()
	sessions := qm.sessionMgr.ListSessions()
	qm.mu.RUnlock()
	
	stats := &QuotaStatistics{
		SessionDetails: make([]SessionQuotaDetail, 0, len(sessions)),
	}
	
	var totalQuota int64
	var totalUsed int64
	
	for _, session := range sessions {
		stats.TotalSessions++
		
		totalQuota += session.DailyTokenLimit
		totalUsed += session.DailyTokenUsed
		
		switch session.QuotaStatus {
		case "available":
			stats.AvailableSessions++
		case "low":
			stats.LowQuotaSessions++
		case "exhausted":
			stats.ExhaustedSessions++
		}
		
		// Calculate estimated exhaustion time
		var estimatedExhaustion *time.Time
		if session.DailyTokenUsed > 0 && session.GetRemainingQuota() > 0 {
			// Calculate tokens per hour
			hoursSinceReset := time.Since(session.LastResetDate).Hours()
			if hoursSinceReset > 0 {
				tokensPerHour := float64(session.DailyTokenUsed) / hoursSinceReset
				if tokensPerHour > 0 {
					hoursUntilExhaustion := float64(session.GetRemainingQuota()) / tokensPerHour
					exhaustionTime := time.Now().Add(time.Duration(hoursUntilExhaustion * float64(time.Hour)))
					estimatedExhaustion = &exhaustionTime
				}
			}
		}
		
		detail := SessionQuotaDetail{
			Email:               session.Email,
			DailyLimit:          session.DailyTokenLimit,
			Used:                session.DailyTokenUsed,
			Remaining:           session.GetRemainingQuota(),
			PercentageUsed:      session.GetQuotaPercentageUsed(),
			Status:              session.QuotaStatus,
			EstimatedExhaustion: estimatedExhaustion,
			AccountType:         session.AccountType,
		}
		
		stats.SessionDetails = append(stats.SessionDetails, detail)
	}
	
	stats.TotalQuota = totalQuota
	stats.TotalUsed = totalUsed
	stats.TotalRemaining = totalQuota - totalUsed
	
	if stats.TotalSessions > 0 {
		stats.AverageUsagePercent = float64(totalUsed) / float64(totalQuota) * 100
	}
	
	// Calculate next reset time
	now := time.Now().UTC()
	nextReset := time.Date(now.Year(), now.Month(), now.Day(), 
		qm.config.ResetHourUTC, 0, 0, 0, time.UTC)
	if now.After(nextReset) {
		nextReset = nextReset.Add(24 * time.Hour)
	}
	stats.NextResetTime = nextReset
	
	return stats, nil
}

// UpdateSessionQuota allows manual quota limit adjustment
func (qm *QuotaManager) UpdateSessionQuota(email string, newLimit int64) error {
	if newLimit <= 0 {
		return fmt.Errorf("quota limit must be positive, got: %d", newLimit)
	}
	
	// Update in database
	if err := database.UpdateSessionQuota(email, newLimit); err != nil {
		return fmt.Errorf("failed to update quota in database: %w", err)
	}
	
	// Reload session from database
	session, err := database.GetCursorSession(email)
	if err != nil {
		return fmt.Errorf("failed to reload session: %w", err)
	}
	
	// Update quota status based on new limit
	session.UpdateQuotaStatus(qm.config.LowQuotaThreshold)
	
	// Persist status update
	if err := database.UpdateSessionQuotaStatus(email, session.QuotaStatus); err != nil {
		logrus.WithError(err).Warn("Failed to update quota status after limit change")
	}
	
	// If new limit is below current usage, mark as exhausted
	if session.DailyTokenUsed >= newLimit {
		if err := database.UpdateSessionStatus(email, false, 0); err != nil {
			logrus.WithError(err).Warn("Failed to mark session as invalid")
		}
	}
	
	// Reload sessions in manager
	if err := qm.sessionMgr.ReloadFromDB(); err != nil {
		return fmt.Errorf("failed to reload sessions: %w", err)
	}
	
	logrus.WithFields(logrus.Fields{
		"email":     email,
		"new_limit": newLimit,
		"status":    session.QuotaStatus,
	}).Info("Session quota limit updated")
	
	return nil
}

// QuotaStatistics represents quota statistics for all sessions
type QuotaStatistics struct {
	TotalSessions       int                  `json:"total_sessions"`
	AvailableSessions   int                  `json:"available_sessions"`
	LowQuotaSessions    int                  `json:"low_quota_sessions"`
	ExhaustedSessions   int                  `json:"exhausted_sessions"`
	TotalQuota          int64                `json:"total_quota"`
	TotalUsed           int64                `json:"total_used"`
	TotalRemaining      int64                `json:"total_remaining"`
	AverageUsagePercent float64              `json:"average_usage_percent"`
	NextResetTime       time.Time            `json:"next_reset_time"`
	SessionDetails      []SessionQuotaDetail `json:"session_details"`
}

// SessionQuotaDetail represents quota details for a single session
type SessionQuotaDetail struct {
	Email               string     `json:"email"`
	DailyLimit          int64      `json:"daily_limit"`
	Used                int64      `json:"used"`
	Remaining           int64      `json:"remaining"`
	PercentageUsed      float64    `json:"percentage_used"`
	Status              string     `json:"status"`
	EstimatedExhaustion *time.Time `json:"estimated_exhaustion,omitempty"`
	AccountType         string     `json:"account_type"`
}
