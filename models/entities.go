package models

import "time"

// KeyInfo 表示存儲於資料層的 API 密鑰資訊
// 單獨放置在 models 包內，方便中間層、資料層與處理器共享，避免循環依賴。
type KeyInfo struct {
    Key           string     `json:"key"`
    MaskedKey     string     `json:"masked_key"`
    TokenName     string     `json:"token_name,omitempty"`
    UserID        *int64     `json:"user_id,omitempty"`
    Username      string     `json:"username,omitempty"`
    CreatedAt     time.Time  `json:"created_at"`
    UsageCount    int64      `json:"usage_count"`
    LastUsedAt    *time.Time `json:"last_used_at,omitempty"`
    IsActive      bool       `json:"is_active"`
    // Balance system extension fields
    QuotaLimit    *float64   `json:"quota_limit,omitempty"`    // Quota limit in USD, nil means unlimited
    QuotaUsed     float64    `json:"quota_used"`               // Quota used in USD
    ExpiresAt     *time.Time `json:"expires_at,omitempty"`     // Expiration time, nil means never expires
    AllowedModels []string   `json:"allowed_models,omitempty"` // Allowed models, nil/empty means all models
}

// CursorSessionInfo 表示 Cursor session 的持久化結構
// 注意：ExtraCookies 序列化為 JSON 字串保存於資料庫
//       讀取時再反序列化為 map。
type CursorSessionInfo struct {
    Token        string            `json:"token"`
    Email        string            `json:"email"`
    CreatedAt    time.Time         `json:"created_at"`
    LastUsed     time.Time         `json:"last_used"`
    LastCheck    time.Time         `json:"last_check"`
    ExpiresAt    time.Time         `json:"expires_at"`
    IsValid      bool              `json:"is_valid"`
    UsageCount   int64             `json:"usage_count"`
    FailCount    int               `json:"fail_count"`
    UserAgent    string            `json:"user_agent"`
    ExtraCookies map[string]string `json:"extra_cookies,omitempty"`
    
    // Quota management fields
    DailyTokenLimit int64     `json:"daily_token_limit"` // Maximum tokens per day
    DailyTokenUsed  int64     `json:"daily_token_used"`  // Tokens used today
    LastResetDate   time.Time `json:"last_reset_date"`   // Last quota reset
    QuotaStatus     string    `json:"quota_status"`      // "available", "low", "exhausted"
    AccountType     string    `json:"account_type"`      // "free", "pro"
}


// GetRemainingQuota calculates tokens remaining for the session
func (s *CursorSessionInfo) GetRemainingQuota() int64 {
	remaining := s.DailyTokenLimit - s.DailyTokenUsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetQuotaPercentageUsed returns percentage of quota consumed
func (s *CursorSessionInfo) GetQuotaPercentageUsed() float64 {
	if s.DailyTokenLimit == 0 {
		return 0
	}
	percentage := float64(s.DailyTokenUsed) / float64(s.DailyTokenLimit) * 100
	if percentage > 100 {
		return 100
	}
	return percentage
}

// IsSuitableForRequest checks if session has enough quota for estimated usage
func (s *CursorSessionInfo) IsSuitableForRequest(estimatedTokens int) bool {
	if !s.IsValid {
		return false
	}
	
	remaining := s.GetRemainingQuota()
	// Add 20% buffer for estimation errors
	required := int64(float64(estimatedTokens) * 1.2)
	return remaining >= required
}

// NeedsQuotaReset checks if session needs quota reset (>24 hours since last reset)
func (s *CursorSessionInfo) NeedsQuotaReset() bool {
	return time.Since(s.LastResetDate) > 24*time.Hour
}

// UpdateQuotaStatus updates the quota status based on remaining quota and threshold
func (s *CursorSessionInfo) UpdateQuotaStatus(lowThreshold float64) {
	percentageUsed := s.GetQuotaPercentageUsed()
	
	if percentageUsed >= 100 {
		s.QuotaStatus = "exhausted"
	} else if percentageUsed >= (lowThreshold * 100) {
		s.QuotaStatus = "low"
	} else {
		s.QuotaStatus = "available"
	}
}


// ============================================================================
// Chat Data Models
// ============================================================================

// Conversation 会话模型 - represents a chat conversation stored in the database
type Conversation struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	Title        string    `json:"title"`
	Model        string    `json:"model"`
	SystemPrompt string    `json:"system_prompt,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ChatMessage 聊天消息模型 - represents a message in a chat conversation stored in the database
// Note: Named ChatMessage to distinguish from the API Message type in models.go
type ChatMessage struct {
	ID             int64     `json:"id"`
	ConversationID int64     `json:"conversation_id"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	Tokens         int       `json:"tokens"`
	Cost           float64   `json:"cost"`
	CreatedAt      time.Time `json:"created_at"`
}

// ChatTokenUsage represents token usage information for AI responses in chat
type ChatTokenUsage struct {
	Prompt     int `json:"prompt"`
	Completion int `json:"completion"`
}

// ChatStreamEvent SSE 事件 - represents a Server-Sent Event for chat streaming
type ChatStreamEvent struct {
	Type      string          `json:"type"`
	MessageID int64           `json:"message_id,omitempty"`
	Delta     string          `json:"delta,omitempty"`
	Tokens    *ChatTokenUsage `json:"tokens,omitempty"`
	Cost      float64         `json:"cost,omitempty"`
	Error     string          `json:"error,omitempty"`
}
