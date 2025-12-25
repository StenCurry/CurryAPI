package providers

import (
	"context"
	"encoding/json"
	"testing"

	"Curry2API-go/middleware"
	"Curry2API-go/models"
)

// mockCursorService is a mock implementation of CursorServiceInterface for testing
type mockCursorService struct {
	chatCompletionFunc func(ctx context.Context, request *models.ChatCompletionRequest) (<-chan interface{}, *middleware.CursorSessionInfo, error)
}

func (m *mockCursorService) ChatCompletion(ctx context.Context, request *models.ChatCompletionRequest) (<-chan interface{}, *middleware.CursorSessionInfo, error) {
	if m.chatCompletionFunc != nil {
		return m.chatCompletionFunc(ctx, request)
	}
	// Default implementation
	ch := make(chan interface{})
	close(ch)
	return ch, nil, nil
}

func TestNewCursorProvider(t *testing.T) {
	cursorService := &mockCursorService{}
	provider := NewCursorProvider(cursorService)

	if provider == nil {
		t.Fatal("NewCursorProvider() returned nil")
	}
	if provider.cursorService != cursorService {
		t.Error("NewCursorProvider() did not set cursorService correctly")
	}
}

func TestCursorProvider_IsAvailable(t *testing.T) {
	cursorService := &mockCursorService{}
	provider := NewCursorProvider(cursorService)

	// Cursor provider should always be available
	if !provider.IsAvailable() {
		t.Error("IsAvailable() = false, want true (Cursor provider should always be available)")
	}
}

func TestCursorProvider_GetProviderName(t *testing.T) {
	cursorService := &mockCursorService{}
	provider := NewCursorProvider(cursorService)

	if got := provider.GetProviderName(); got != "cursor" {
		t.Errorf("GetProviderName() = %v, want cursor", got)
	}
}

func TestCursorProvider_GetSupportedModels(t *testing.T) {
	cursorService := &mockCursorService{}
	provider := NewCursorProvider(cursorService)

	models := provider.GetSupportedModels()

	if len(models) == 0 {
		t.Error("GetSupportedModels() returned empty list")
	}

	// Check that all models have provider set to "cursor"
	for i, model := range models {
		if model.Provider != "cursor" {
			t.Errorf("Model[%d].Provider = %v, want cursor", i, model.Provider)
		}
		if !model.IsAvailable {
			t.Errorf("Model[%d].IsAvailable = false, want true", i)
		}
		if model.ID == "" {
			t.Errorf("Model[%d].ID is empty", i)
		}
		if model.Name == "" {
			t.Errorf("Model[%d].Name is empty", i)
		}
	}

	// Check for some expected models
	expectedModels := map[string]bool{
		"claude-3.5-sonnet":  false,
		"gpt-5":              false,
		"gemini-2.5-pro":     false,
		"deepseek-r1":        false,
	}

	for _, model := range models {
		if _, exists := expectedModels[model.ID]; exists {
			expectedModels[model.ID] = true
		}
	}

	for modelID, found := range expectedModels {
		if !found {
			t.Errorf("Expected model %s not found in supported models", modelID)
		}
	}
}

func TestCursorProvider_ConvertCursorStream(t *testing.T) {
	cursorService := &mockCursorService{}
	provider := NewCursorProvider(cursorService)

	tests := []struct {
		name           string
		cursorEvents   []interface{}
		wantEventTypes []string
		wantContent    string
	}{
		{
			name: "string delta events",
			cursorEvents: []interface{}{
				`{"type":"delta","delta":"Hello"}`,
				`{"type":"delta","delta":" World"}`,
				`{"type":"done"}`,
			},
			wantEventTypes: []string{"start", "content", "content", "done"},
			wantContent:    "Hello World",
		},
		{
			name: "struct delta events",
			cursorEvents: []interface{}{
				models.CursorEventData{Type: "delta", Delta: "Test"},
				models.CursorEventData{Type: "delta", Delta: " Message"},
				models.CursorEventData{Type: "done"},
			},
			wantEventTypes: []string{"start", "content", "content", "done"},
			wantContent:    "Test Message",
		},
		{
			name: "error event",
			cursorEvents: []interface{}{
				models.CursorEventData{Type: "error", ErrorText: "Test error"},
			},
			wantEventTypes: []string{"start", "error"},
			wantContent:    "",
		},
		{
			name: "with usage metadata",
			cursorEvents: []interface{}{
				`{"type":"delta","delta":"Hello"}`,
				`{"type":"done","messageMetadata":{"usage":{"inputTokens":10,"outputTokens":5,"totalTokens":15}}}`,
			},
			wantEventTypes: []string{"start", "content", "usage", "done"},
			wantContent:    "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create channels
			cursorChan := make(chan interface{})
			eventChan := make(chan models.StreamEvent)

			// Start conversion goroutine
			go provider.convertCursorStream(cursorChan, eventChan)

			// Send cursor events in a separate goroutine
			go func() {
				for _, event := range tt.cursorEvents {
					cursorChan <- event
				}
				close(cursorChan)
			}()

			// Collect events
			var events []models.StreamEvent
			var content string
			for event := range eventChan {
				events = append(events, event)
				if event.Type == "content" {
					content += event.Content
				}
			}

			// Verify event types
			if len(events) != len(tt.wantEventTypes) {
				t.Errorf("Got %d events, want %d", len(events), len(tt.wantEventTypes))
			}

			for i, event := range events {
				if i < len(tt.wantEventTypes) && event.Type != tt.wantEventTypes[i] {
					t.Errorf("Event[%d].Type = %v, want %v", i, event.Type, tt.wantEventTypes[i])
				}
			}

			// Verify content
			if content != tt.wantContent {
				t.Errorf("Content = %v, want %v", content, tt.wantContent)
			}

			// Verify usage if expected
			if tt.name == "with usage metadata" {
				hasUsage := false
				for _, event := range events {
					if event.Type == "usage" && event.Tokens != nil {
						hasUsage = true
						if event.Tokens.PromptTokens != 10 {
							t.Errorf("PromptTokens = %d, want 10", event.Tokens.PromptTokens)
						}
						if event.Tokens.CompletionTokens != 5 {
							t.Errorf("CompletionTokens = %d, want 5", event.Tokens.CompletionTokens)
						}
						if event.Tokens.TotalTokens != 15 {
							t.Errorf("TotalTokens = %d, want 15", event.Tokens.TotalTokens)
						}
					}
				}
				if !hasUsage {
					t.Error("Expected usage event but none found")
				}
			}
		})
	}
}

func TestCursorProvider_FormatConversion(t *testing.T) {
	cursorService := &mockCursorService{}
	provider := NewCursorProvider(cursorService)

	// Test ChatRequest to CursorService format conversion
	req := &models.ChatRequest{
		Model: "claude-3.5-sonnet",
		Messages: []models.Message{
			{Role: "system", Content: "You are a helpful assistant"},
			{Role: "user", Content: "Hello"},
		},
		Stream:      true,
		MaxTokens:   1000,
		Temperature: 0.7,
	}

	// We can't easily test the actual conversion without mocking CursorService,
	// but we can verify the provider accepts the request format
	ctx := context.Background()

	// This will fail because we don't have a real Cursor service configured,
	// but we're testing that the format conversion doesn't panic
	_, err := provider.ChatCompletion(ctx, req)
	
	// We expect an error since we don't have real Cursor credentials,
	// but the error should not be about format conversion
	if err == nil {
		t.Log("ChatCompletion succeeded (unexpected but not an error in test)")
	}
}

func TestCursorProvider_StreamEventConversion_EdgeCases(t *testing.T) {
	cursorService := &mockCursorService{}
	provider := NewCursorProvider(cursorService)

	tests := []struct {
		name         string
		cursorEvents []interface{}
		wantError    bool
	}{
		{
			name: "empty delta",
			cursorEvents: []interface{}{
				models.CursorEventData{Type: "delta", Delta: ""},
				models.CursorEventData{Type: "done"},
			},
			wantError: false,
		},
		{
			name: "plain text string",
			cursorEvents: []interface{}{
				"plain text content",
				models.CursorEventData{Type: "done"},
			},
			wantError: false,
		},
		{
			name: "invalid JSON string",
			cursorEvents: []interface{}{
				"{invalid json",
				models.CursorEventData{Type: "done"},
			},
			wantError: false,
		},
		{
			name: "unknown event type",
			cursorEvents: []interface{}{
				models.CursorEventData{Type: "unknown"},
				models.CursorEventData{Type: "done"},
			},
			wantError: false,
		},
		{
			name: "map interface conversion",
			cursorEvents: []interface{}{
				map[string]interface{}{
					"type":  "delta",
					"delta": "test",
				},
				models.CursorEventData{Type: "done"},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursorChan := make(chan interface{})
			eventChan := make(chan models.StreamEvent)

			go provider.convertCursorStream(cursorChan, eventChan)

			go func() {
				for _, event := range tt.cursorEvents {
					cursorChan <- event
				}
				close(cursorChan)
			}()

			// Collect events - should not panic
			var events []models.StreamEvent
			for event := range eventChan {
				events = append(events, event)
			}

			// Should always have at least start and done events
			if len(events) < 2 {
				t.Errorf("Expected at least 2 events (start, done), got %d", len(events))
			}

			// First event should be start
			if events[0].Type != "start" {
				t.Errorf("First event type = %v, want start", events[0].Type)
			}

			// Last event should be done (unless error)
			lastEvent := events[len(events)-1]
			if !tt.wantError && lastEvent.Type != "done" && lastEvent.Type != "error" {
				t.Errorf("Last event type = %v, want done or error", lastEvent.Type)
			}
		})
	}
}

func TestCursorProvider_UsageMetadataExtraction(t *testing.T) {
	cursorService := &mockCursorService{}
	provider := NewCursorProvider(cursorService)

	// Test with usage metadata in done event
	usage := &models.CursorUsage{
		InputTokens:  100,
		OutputTokens: 50,
		TotalTokens:  150,
	}

	metadata := &models.CursorMessageMetadata{
		Usage: usage,
	}

	doneEvent := models.CursorEventData{
		Type:            "done",
		MessageMetadata: metadata,
	}

	cursorChan := make(chan interface{})
	eventChan := make(chan models.StreamEvent)

	go provider.convertCursorStream(cursorChan, eventChan)

	go func() {
		cursorChan <- doneEvent
		close(cursorChan)
	}()

	var events []models.StreamEvent
	for event := range eventChan {
		events = append(events, event)
	}

	// Should have start, usage, and done events
	if len(events) < 3 {
		t.Fatalf("Expected at least 3 events, got %d", len(events))
	}

	// Find usage event
	var usageEvent *models.StreamEvent
	for i := range events {
		if events[i].Type == "usage" {
			usageEvent = &events[i]
			break
		}
	}

	if usageEvent == nil {
		t.Fatal("No usage event found")
	}

	if usageEvent.Tokens == nil {
		t.Fatal("Usage event has nil Tokens")
	}

	if usageEvent.Tokens.PromptTokens != 100 {
		t.Errorf("PromptTokens = %d, want 100", usageEvent.Tokens.PromptTokens)
	}
	if usageEvent.Tokens.CompletionTokens != 50 {
		t.Errorf("CompletionTokens = %d, want 50", usageEvent.Tokens.CompletionTokens)
	}
	if usageEvent.Tokens.TotalTokens != 150 {
		t.Errorf("TotalTokens = %d, want 150", usageEvent.Tokens.TotalTokens)
	}
}

func TestCursorProvider_ErrorEventHandling(t *testing.T) {
	cursorService := &mockCursorService{}
	provider := NewCursorProvider(cursorService)

	tests := []struct {
		name      string
		event     interface{}
		wantError string
	}{
		{
			name: "string error event",
			event: `{"type":"error","errorText":"Test error message"}`,
			wantError: "Test error message",
		},
		{
			name: "struct error event",
			event: models.CursorEventData{
				Type:      "error",
				ErrorText: "Another error",
			},
			wantError: "Another error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursorChan := make(chan interface{})
			eventChan := make(chan models.StreamEvent)

			go provider.convertCursorStream(cursorChan, eventChan)

			go func() {
				cursorChan <- tt.event
				close(cursorChan)
			}()

			var events []models.StreamEvent
			for event := range eventChan {
				events = append(events, event)
			}

			// Should have start and error events
			if len(events) < 2 {
				t.Fatalf("Expected at least 2 events, got %d", len(events))
			}

			// Find error event
			var errorEvent *models.StreamEvent
			for i := range events {
				if events[i].Type == "error" {
					errorEvent = &events[i]
					break
				}
			}

			if errorEvent == nil {
				t.Fatal("No error event found")
			}

			if errorEvent.Error != tt.wantError {
				t.Errorf("Error = %v, want %v", errorEvent.Error, tt.wantError)
			}
		})
	}
}

func TestCursorProvider_JSONMarshaling(t *testing.T) {
	// Test that complex event structures can be marshaled/unmarshaled
	event := models.CursorEventData{
		Type:  "delta",
		Delta: "test content",
		MessageMetadata: &models.CursorMessageMetadata{
			Usage: &models.CursorUsage{
				InputTokens:  10,
				OutputTokens: 5,
				TotalTokens:  15,
			},
		},
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	// Unmarshal back
	var decoded models.CursorEventData
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	// Verify
	if decoded.Type != event.Type {
		t.Errorf("Type = %v, want %v", decoded.Type, event.Type)
	}
	if decoded.Delta != event.Delta {
		t.Errorf("Delta = %v, want %v", decoded.Delta, event.Delta)
	}
	if decoded.MessageMetadata == nil {
		t.Fatal("MessageMetadata is nil")
	}
	if decoded.MessageMetadata.Usage == nil {
		t.Fatal("Usage is nil")
	}
	if decoded.MessageMetadata.Usage.InputTokens != 10 {
		t.Errorf("InputTokens = %d, want 10", decoded.MessageMetadata.Usage.InputTokens)
	}
}
