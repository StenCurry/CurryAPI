package services

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

// Provider error types (Requirements: 10.1-10.5)
var (
	// ErrProviderNotAvailable indicates the provider is not configured or unavailable
	ErrProviderNotAvailable = errors.New("provider not available")

	// ErrInvalidAPIKey indicates the API key is invalid or expired (HTTP 401)
	ErrInvalidAPIKey = errors.New("API key is invalid or expired")

	// ErrRateLimited indicates rate limit exceeded (HTTP 429)
	ErrRateLimited = errors.New("rate limit exceeded, please try again later")

	// ErrProviderError indicates a server-side error from the provider (HTTP 500-599)
	ErrProviderError = errors.New("AI service temporarily unavailable")

	// ErrTimeout indicates the request timed out
	ErrTimeout = errors.New("request timed out")

	// ErrContextTooLong indicates the context/message is too long for the model
	ErrContextTooLong = errors.New("context too long for this model")
)

// ProviderErrorCode represents standardized error codes
type ProviderErrorCode string

const (
	ErrorCodeProviderNotAvailable ProviderErrorCode = "PROVIDER_NOT_AVAILABLE"
	ErrorCodeInvalidAPIKey        ProviderErrorCode = "INVALID_API_KEY"
	ErrorCodeRateLimited          ProviderErrorCode = "RATE_LIMITED"
	ErrorCodeProviderError        ProviderErrorCode = "PROVIDER_ERROR"
	ErrorCodeTimeout              ProviderErrorCode = "TIMEOUT"
	ErrorCodeContextTooLong       ProviderErrorCode = "CONTEXT_TOO_LONG"
	ErrorCodeBadRequest           ProviderErrorCode = "BAD_REQUEST"
	ErrorCodeUnknown              ProviderErrorCode = "UNKNOWN_ERROR"
)

// ProviderError represents a structured provider error with context
type ProviderError struct {
	Code       ProviderErrorCode
	Message    string
	Provider   string
	Model      string
	RequestID  string
	StatusCode int
	Cause      error
}

// Error implements the error interface
func (e *ProviderError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return string(e.Code)
}

// Unwrap returns the underlying cause
func (e *ProviderError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches a target error
func (e *ProviderError) Is(target error) bool {
	switch target {
	case ErrProviderNotAvailable:
		return e.Code == ErrorCodeProviderNotAvailable
	case ErrInvalidAPIKey:
		return e.Code == ErrorCodeInvalidAPIKey
	case ErrRateLimited:
		return e.Code == ErrorCodeRateLimited
	case ErrProviderError:
		return e.Code == ErrorCodeProviderError
	case ErrTimeout:
		return e.Code == ErrorCodeTimeout
	case ErrContextTooLong:
		return e.Code == ErrorCodeContextTooLong
	}
	return false
}

// NewProviderError creates a new ProviderError
func NewProviderError(code ProviderErrorCode, message, provider, model, requestID string) *ProviderError {
	return &ProviderError{
		Code:      code,
		Message:   message,
		Provider:  provider,
		Model:     model,
		RequestID: requestID,
	}
}

// MapHTTPStatusToError maps HTTP status codes to appropriate ProviderError
// Requirements: 10.1, 10.2, 10.3, 10.4, 10.5
func MapHTTPStatusToError(statusCode int, responseBody string, provider, model, requestID string) *ProviderError {
	err := &ProviderError{
		Provider:   provider,
		Model:      model,
		RequestID:  requestID,
		StatusCode: statusCode,
	}

	switch statusCode {
	case http.StatusUnauthorized: // 401
		// Requirements: 10.1 - INVALID_API_KEY
		err.Code = ErrorCodeInvalidAPIKey
		err.Message = "API key is invalid or expired"

	case http.StatusTooManyRequests: // 429
		// Requirements: 10.2 - RATE_LIMITED
		err.Code = ErrorCodeRateLimited
		err.Message = "Rate limit exceeded, please try again later"

	case http.StatusBadRequest: // 400
		// Check if it's a context length error
		lowerBody := strings.ToLower(responseBody)
		if strings.Contains(lowerBody, "context") || 
		   strings.Contains(lowerBody, "token") ||
		   strings.Contains(lowerBody, "maximum") ||
		   strings.Contains(lowerBody, "length") {
			// Requirements: 10.5 - CONTEXT_TOO_LONG
			err.Code = ErrorCodeContextTooLong
			err.Message = extractContextLengthMessage(responseBody)
		} else {
			err.Code = ErrorCodeBadRequest
			err.Message = responseBody
		}

	default:
		if statusCode >= 500 && statusCode < 600 {
			// Requirements: 10.3 - PROVIDER_ERROR for 500-599
			err.Code = ErrorCodeProviderError
			err.Message = "AI service temporarily unavailable"
		} else {
			err.Code = ErrorCodeUnknown
			err.Message = responseBody
		}
	}

	return err
}

// MapTimeoutError creates a timeout error
// Requirements: 10.4
func MapTimeoutError(provider, model, requestID string) *ProviderError {
	return &ProviderError{
		Code:      ErrorCodeTimeout,
		Message:   "Request timed out",
		Provider:  provider,
		Model:     model,
		RequestID: requestID,
	}
}

// MapProviderNotAvailableError creates a provider not available error
// Requirements: 2.6
func MapProviderNotAvailableError(provider, model, requestID string) *ProviderError {
	return &ProviderError{
		Code:      ErrorCodeProviderNotAvailable,
		Message:   fmt.Sprintf("%s provider requires API key configuration", provider),
		Provider:  provider,
		Model:     model,
		RequestID: requestID,
	}
}

// extractContextLengthMessage extracts a user-friendly message for context length errors
func extractContextLengthMessage(responseBody string) string {
	// Try to extract max context length from error message
	lowerBody := strings.ToLower(responseBody)
	
	// Common patterns in error messages
	if strings.Contains(lowerBody, "maximum context length") {
		return responseBody
	}
	if strings.Contains(lowerBody, "token limit") {
		return responseBody
	}
	
	return "Message too long for this model"
}

// LogProviderError logs a provider error with structured fields
// Requirements: 10.6
func LogProviderError(err *ProviderError) {
	logrus.WithFields(logrus.Fields{
		"request_id":   err.RequestID,
		"provider":     err.Provider,
		"model":        err.Model,
		"error_code":   err.Code,
		"error_message": err.Message,
		"status_code":  err.StatusCode,
	}).Error("Provider error occurred")
}

// LogProviderErrorWithContext logs a provider error with additional context
// Requirements: 10.6
func LogProviderErrorWithContext(requestID, provider, model string, errorCode ProviderErrorCode, errorMessage string) {
	logrus.WithFields(logrus.Fields{
		"request_id":    requestID,
		"provider":      provider,
		"model":         model,
		"error_code":    errorCode,
		"error_message": errorMessage,
	}).Error("Provider error occurred")
}

// ParseErrorFromString parses an error string to determine the error type
// This is useful for parsing errors from provider responses
func ParseErrorFromString(errStr string) ProviderErrorCode {
	lowerErr := strings.ToLower(errStr)

	if strings.Contains(lowerErr, "invalid_api_key") || 
	   strings.Contains(lowerErr, "invalid api key") ||
	   strings.Contains(lowerErr, "authentication") ||
	   strings.Contains(lowerErr, "unauthorized") {
		return ErrorCodeInvalidAPIKey
	}

	if strings.Contains(lowerErr, "rate_limited") || 
	   strings.Contains(lowerErr, "rate limit") ||
	   strings.Contains(lowerErr, "too many requests") {
		return ErrorCodeRateLimited
	}

	if strings.Contains(lowerErr, "timeout") ||
	   strings.Contains(lowerErr, "timed out") {
		return ErrorCodeTimeout
	}

	if strings.Contains(lowerErr, "context_too_long") ||
	   strings.Contains(lowerErr, "context too long") ||
	   strings.Contains(lowerErr, "maximum context") ||
	   strings.Contains(lowerErr, "token limit") {
		return ErrorCodeContextTooLong
	}

	if strings.Contains(lowerErr, "provider_not_available") ||
	   strings.Contains(lowerErr, "not available") ||
	   strings.Contains(lowerErr, "not configured") {
		return ErrorCodeProviderNotAvailable
	}

	if strings.Contains(lowerErr, "provider_error") ||
	   strings.Contains(lowerErr, "service unavailable") ||
	   strings.Contains(lowerErr, "internal server error") {
		return ErrorCodeProviderError
	}

	return ErrorCodeUnknown
}

// WrapError wraps an existing error with provider context
func WrapError(err error, provider, model, requestID string) *ProviderError {
	if err == nil {
		return nil
	}

	// Check if it's already a ProviderError
	var providerErr *ProviderError
	if errors.As(err, &providerErr) {
		// Update context if not set
		if providerErr.Provider == "" {
			providerErr.Provider = provider
		}
		if providerErr.Model == "" {
			providerErr.Model = model
		}
		if providerErr.RequestID == "" {
			providerErr.RequestID = requestID
		}
		return providerErr
	}

	// Parse error string to determine type
	errCode := ParseErrorFromString(err.Error())

	return &ProviderError{
		Code:      errCode,
		Message:   err.Error(),
		Provider:  provider,
		Model:     model,
		RequestID: requestID,
		Cause:     err,
	}
}

// GetUserFriendlyMessage returns a user-friendly error message
func (e *ProviderError) GetUserFriendlyMessage() string {
	switch e.Code {
	case ErrorCodeInvalidAPIKey:
		return "API key is invalid or expired"
	case ErrorCodeRateLimited:
		return "Rate limit exceeded, please try again later"
	case ErrorCodeProviderError:
		return "AI service temporarily unavailable"
	case ErrorCodeTimeout:
		return "Request timed out"
	case ErrorCodeContextTooLong:
		if e.Message != "" && e.Message != "context too long for this model" {
			return e.Message
		}
		return "Message too long for this model"
	case ErrorCodeProviderNotAvailable:
		return fmt.Sprintf("%s provider is not configured", strings.Title(e.Provider))
	default:
		if e.Message != "" {
			return e.Message
		}
		return "An unexpected error occurred"
	}
}
