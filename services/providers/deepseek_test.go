package providers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"Curry2API-go/models"
)

func TestNewDeepSeekProvider(t *testing.T) {
	tests := []struct {
		name        string
		apiKey      string
		baseURL     string
		wantBaseURL string
	}{
		{
			name:        "with custom base URL",
			apiKey:      "test-key",
			baseURL:     "https://custom.deepseek.com/v1",
			wantBaseURL: "https://custom.deepseek.com/v1",
		},
		{
			name:        "with empty base URL uses default",
			apiKey:      "test-key",
			baseURL:     "",
			wantBaseURL: "https://api.deepseek.com/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewDeepSeekProvider(tt.apiKey, tt.baseURL)
			if provider.apiKey != tt.apiKey {
				t.Errorf("apiKey = %v, want %v", provider.apiKey, tt.apiKey)
			}
			if provider.baseURL != tt.wantBaseURL {
				t.Errorf("baseURL = %v, want %v", provider.baseURL, tt.wantBaseURL)
			}
		})
	}
}

func TestDeepSeekProvider_IsAvailable(t *testing.T) {
	tests := []struct {
		name   string
		apiKey string
		want   bool
	}{
		{
			name:   "available with API key",
			apiKey: "test-key",
			want:   true,
		},
		{
			name:   "not available without API key",
			apiKey: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewDeepSeekProvider(tt.apiKey, "")
			if got := provider.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeepSeekProvider_GetProviderName(t *testing.T) {
	provider := NewDeepSeekProvider("test-key", "")
	if got := provider.GetProviderName(); got != "deepseek" {
		t.Errorf("GetProviderName() = %v, want %v", got, "deepseek")
	}
}

func TestDeepSeekProvider_GetSupportedModels(t *testing.T) {
	tests := []struct {
		name       string
		apiKey     string
		wantModels []string
		wantAvail  bool
	}{
		{
			name:   "with API key",
			apiKey: "test-key",
			wantModels: []string{
				"deepseek-chat", "deepseek-coder", "deepseek-reasoner",
			},
			wantAvail: true,
		},
		{
			name:   "without API key",
			apiKey: "",
			wantModels: []string{
				"deepseek-chat", "deepseek-coder", "deepseek-reasoner",
			},
			wantAvail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewDeepSeekProvider(tt.apiKey, "")
			models := provider.GetSupportedModels()

			if len(models) != len(tt.wantModels) {
				t.Errorf("GetSupportedModels() returned %d models, want %d", len(models), len(tt.wantModels))
			}

			for i, model := range models {
				if model.ID != tt.wantModels[i] {
					t.Errorf("Model[%d].ID = %v, want %v", i, model.ID, tt.wantModels[i])
				}
				if model.Provider != "deepseek" {
					t.Errorf("Model[%d].Provider = %v, want deepseek", i, model.Provider)
				}
				if model.IsAvailable != tt.wantAvail {
					t.Errorf("Model[%d].IsAvailable = %v, want %v", i, model.IsAvailable, tt.wantAvail)
				}
			}
		})
	}
}

func TestDeepSeekProvider_ChatCompletion_NotAvailable(t *testing.T) {
	provider := NewDeepSeekProvider("", "")
	ctx := context.Background()

	req := &models.ChatRequest{
		Model: "deepseek-chat",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
		Stream: true,
	}

	_, err := provider.ChatCompletion(ctx, req)
	if err == nil {
		t.Error("ChatCompletion() should return error when provider not available")
	}
	if !strings.Contains(err.Error(), "not available") {
		t.Errorf("ChatCompletion() error = %v, want error containing 'not available'", err)
	}
}

func TestDeepSeekProvider_ChatCompletion_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Expected Authorization header with Bearer token")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json")
		}

		// Send streaming response
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("Expected http.ResponseWriter to be an http.Flusher")
		}

		// Send start chunk
		w.Write([]byte(`data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1234567890,"model":"deepseek-chat","choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}` + "\n\n"))
		flusher.Flush()

		// Send content chunks
		w.Write([]byte(`data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1234567890,"model":"deepseek-chat","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}` + "\n\n"))
		flusher.Flush()

		w.Write([]byte(`data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1234567890,"model":"deepseek-chat","choices":[{"index":0,"delta":{"content":" World"},"finish_reason":null}]}` + "\n\n"))
		flusher.Flush()

		// Send finish chunk
		finishReason := "stop"
		w.Write([]byte(`data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1234567890,"model":"deepseek-chat","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}` + "\n\n"))
		flusher.Flush()

		// Send [DONE]
		w.Write([]byte("data: [DONE]\n\n"))
		flusher.Flush()

		_ = finishReason // Use the variable
	}))
	defer server.Close()

	provider := NewDeepSeekProvider("test-key", server.URL)
	ctx := context.Background()

	req := &models.ChatRequest{
		Model: "deepseek-chat",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
		Stream: true,
	}

	eventChan, err := provider.ChatCompletion(ctx, req)
	if err != nil {
		t.Fatalf("ChatCompletion() error = %v", err)
	}

	// Collect events
	var events []models.StreamEvent
	for event := range eventChan {
		events = append(events, event)
	}

	// Verify events
	if len(events) < 3 {
		t.Errorf("Expected at least 3 events (start, content, done), got %d", len(events))
	}

	// Check start event
	if events[0].Type != "start" {
		t.Errorf("First event type = %v, want start", events[0].Type)
	}

	// Check for content events
	hasContent := false
	for _, event := range events {
		if event.Type == "content" && event.Content != "" {
			hasContent = true
			break
		}
	}
	if !hasContent {
		t.Error("Expected at least one content event with non-empty content")
	}

	// Check done event
	lastEvent := events[len(events)-1]
	if lastEvent.Type != "done" {
		t.Errorf("Last event type = %v, want done", lastEvent.Type)
	}
}

func TestDeepSeekProvider_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		responseBody  string
		wantErrorCode string
	}{
		{
			name:          "401 unauthorized",
			statusCode:    http.StatusUnauthorized,
			responseBody:  `{"error":{"message":"Invalid API key","type":"invalid_request_error"}}`,
			wantErrorCode: "INVALID_API_KEY",
		},
		{
			name:          "429 rate limited",
			statusCode:    http.StatusTooManyRequests,
			responseBody:  `{"error":{"message":"Rate limit exceeded","type":"rate_limit_error"}}`,
			wantErrorCode: "RATE_LIMITED",
		},
		{
			name:          "500 server error",
			statusCode:    http.StatusInternalServerError,
			responseBody:  `{"error":{"message":"Internal server error","type":"server_error"}}`,
			wantErrorCode: "PROVIDER_ERROR",
		},
		{
			name:          "400 context too long",
			statusCode:    http.StatusBadRequest,
			responseBody:  `{"error":{"message":"Maximum context length exceeded","type":"invalid_request_error"}}`,
			wantErrorCode: "CONTEXT_TOO_LONG",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			provider := NewDeepSeekProvider("test-key", server.URL)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			req := &models.ChatRequest{
				Model: "deepseek-chat",
				Messages: []models.Message{
					{Role: "user", Content: "Hello"},
				},
				Stream: true,
			}

			_, err := provider.ChatCompletion(ctx, req)
			if err == nil {
				t.Error("ChatCompletion() should return error")
			}
			if !strings.Contains(err.Error(), tt.wantErrorCode) {
				t.Errorf("ChatCompletion() error = %v, want error containing %v", err, tt.wantErrorCode)
			}
		})
	}
}
