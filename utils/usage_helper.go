package utils

import (
	"errors"
	"github.com/gin-gonic/gin"
)

// UsageContextInfo contains user and token information extracted from context
type UsageContextInfo struct {
	UserID    int64
	Username  string
	APIToken  string
	TokenName string
}

// ExtractUsageFromContext extracts usage tracking information from gin context
// Returns error if required fields (api_key) are missing
func ExtractUsageFromContext(c *gin.Context) (*UsageContextInfo, error) {
	info := &UsageContextInfo{}
	
	// Extract API token (required)
	apiKey, exists := c.Get("api_key")
	if !exists {
		return nil, errors.New("api_key not found in context")
	}
	apiKeyStr, ok := apiKey.(string)
	if !ok || apiKeyStr == "" {
		return nil, errors.New("invalid api_key in context")
	}
	info.APIToken = apiKeyStr
	
	// Extract user ID (optional - may not exist for legacy keys)
	if userID, exists := c.Get("user_id"); exists {
		if userIDInt64, ok := userID.(int64); ok {
			info.UserID = userIDInt64
		}
	}
	
	// Extract username (optional)
	if username, exists := c.Get("username"); exists {
		if usernameStr, ok := username.(string); ok {
			info.Username = usernameStr
		}
	}
	
	// Extract token name (optional)
	if tokenName, exists := c.Get("token_name"); exists {
		if tokenNameStr, ok := tokenName.(string); ok {
			info.TokenName = tokenNameStr
		}
	}
	
	return info, nil
}
