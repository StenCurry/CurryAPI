package providers

import (
	"reflect"
	"testing"

	"Curry2API-go/models"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"
)

// **Feature: multi-ai-provider, Property 10: Anthropic Message Format Conversion**
// **Validates: Requirements 5.4**
func TestProperty_AnthropicMessageFormatConversion(t *testing.T) {
	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 100

	properties := gopter.NewProperties(params)

	// Generator for OpenAI-format message arrays
	genMessages := gen.SliceOf(
		gen.OneGenOf(
			// System messages
			gen.Struct(reflect.TypeOf(models.Message{}), map[string]gopter.Gen{
				"Role":    gen.Const("system"),
				"Content": gen.AlphaString(),
			}),
			// User messages
			gen.Struct(reflect.TypeOf(models.Message{}), map[string]gopter.Gen{
				"Role":    gen.Const("user"),
				"Content": gen.AlphaString(),
			}),
			// Assistant messages
			gen.Struct(reflect.TypeOf(models.Message{}), map[string]gopter.Gen{
				"Role":    gen.Const("assistant"),
				"Content": gen.AlphaString(),
			}),
		),
	).SuchThat(func(msgs []models.Message) bool {
		// Ensure we have at least one message
		return len(msgs) > 0
	})

	properties.Property("For any OpenAI-format message array, converting to Anthropic format preserves semantic content",
		prop.ForAll(
			func(messages []models.Message) bool {
				provider := NewAnthropicProvider("test-key", "")

				// Convert to Anthropic format
				anthropicMessages, systemPrompt, err := provider.convertToAnthropicFormat(messages)
				if err != nil {
					return false
				}

				// Count original system, user, and assistant messages
				var originalSystemCount, originalUserCount, originalAssistantCount int
				var originalSystemContent, originalUserContent, originalAssistantContent string

				for _, msg := range messages {
					switch msg.Role {
					case "system":
						originalSystemCount++
						if content, ok := msg.Content.(string); ok {
							if originalSystemContent != "" {
								originalSystemContent += "\n"
							}
							originalSystemContent += content
						}
					case "user":
						originalUserCount++
						if content, ok := msg.Content.(string); ok {
							originalUserContent += content
						}
					case "assistant":
						originalAssistantCount++
						if content, ok := msg.Content.(string); ok {
							originalAssistantContent += content
						}
					}
				}

				// Verify system prompt extracted correctly
				if originalSystemCount > 0 && systemPrompt != originalSystemContent {
					return false
				}

				// Verify user/assistant messages preserved
				var convertedUserCount, convertedAssistantCount int
				var convertedUserContent, convertedAssistantContent string

				for _, msg := range anthropicMessages {
					switch msg.Role {
					case "user":
						convertedUserCount++
						convertedUserContent += msg.Content
					case "assistant":
						convertedAssistantCount++
						convertedAssistantContent += msg.Content
					}
				}

				// Check counts match
				if originalUserCount != convertedUserCount {
					return false
				}
				if originalAssistantCount != convertedAssistantCount {
					return false
				}

				// Check content preserved
				if originalUserContent != convertedUserContent {
					return false
				}
				if originalAssistantContent != convertedAssistantContent {
					return false
				}

				return true
			},
			genMessages,
		))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Unit tests for Anthropic provider

func TestNewAnthropicProvider(t *testing.T) {
	tests := []struct {
		name        string
		apiKey      string
		baseURL     string
		expectedURL string
	}{
		{
			name:        "with custom base URL",
			apiKey:      "test-key",
			baseURL:     "https://custom.api.com/v1",
			expectedURL: "https://custom.api.com/v1",
		},
		{
			name:        "with empty base URL",
			apiKey:      "test-key",
			baseURL:     "",
			expectedURL: "https://api.anthropic.com/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewAnthropicProvider(tt.apiKey, tt.baseURL)
			assert.NotNil(t, provider)
			assert.Equal(t, tt.apiKey, provider.apiKey)
			assert.Equal(t, tt.expectedURL, provider.baseURL)
		})
	}
}

func TestAnthropicProvider_IsAvailable(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected bool
	}{
		{
			name:     "with API key",
			apiKey:   "test-key",
			expected: true,
		},
		{
			name:     "without API key",
			apiKey:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewAnthropicProvider(tt.apiKey, "")
			assert.Equal(t, tt.expected, provider.IsAvailable())
		})
	}
}

func TestAnthropicProvider_GetProviderName(t *testing.T) {
	provider := NewAnthropicProvider("test-key", "")
	assert.Equal(t, "anthropic", provider.GetProviderName())
}

func TestAnthropicProvider_GetSupportedModels(t *testing.T) {
	provider := NewAnthropicProvider("test-key", "")
	models := provider.GetSupportedModels()

	// Should have 5 Claude models
	assert.Len(t, models, 5)

	// Check that all models are from anthropic provider
	for _, model := range models {
		assert.Equal(t, "anthropic", model.Provider)
		assert.True(t, model.IsAvailable)
	}

	// Check specific models exist
	modelIDs := make(map[string]bool)
	for _, model := range models {
		modelIDs[model.ID] = true
	}

	assert.True(t, modelIDs["claude-3-5-sonnet-20241022"])
	assert.True(t, modelIDs["claude-3-5-haiku-20241022"])
	assert.True(t, modelIDs["claude-3-opus-20240229"])
	assert.True(t, modelIDs["claude-3-sonnet-20240229"])
	assert.True(t, modelIDs["claude-3-haiku-20240307"])
}

func TestAnthropicProvider_ConvertToAnthropicFormat(t *testing.T) {
	provider := NewAnthropicProvider("test-key", "")

	tests := []struct {
		name                    string
		messages                []models.Message
		expectedSystemPrompt    string
		expectedMessageCount    int
		expectedFirstRole       string
		expectedFirstContent    string
	}{
		{
			name: "system and user messages",
			messages: []models.Message{
				{Role: "system", Content: "You are a helpful assistant"},
				{Role: "user", Content: "Hello"},
			},
			expectedSystemPrompt: "You are a helpful assistant",
			expectedMessageCount: 1,
			expectedFirstRole:    "user",
			expectedFirstContent: "Hello",
		},
		{
			name: "multiple system messages",
			messages: []models.Message{
				{Role: "system", Content: "First instruction"},
				{Role: "system", Content: "Second instruction"},
				{Role: "user", Content: "Hello"},
			},
			expectedSystemPrompt: "First instruction\nSecond instruction",
			expectedMessageCount: 1,
			expectedFirstRole:    "user",
			expectedFirstContent: "Hello",
		},
		{
			name: "user and assistant messages",
			messages: []models.Message{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi there!"},
				{Role: "user", Content: "How are you?"},
			},
			expectedSystemPrompt: "",
			expectedMessageCount: 3,
			expectedFirstRole:    "user",
			expectedFirstContent: "Hello",
		},
		{
			name: "no system messages",
			messages: []models.Message{
				{Role: "user", Content: "Hello"},
			},
			expectedSystemPrompt: "",
			expectedMessageCount: 1,
			expectedFirstRole:    "user",
			expectedFirstContent: "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anthropicMessages, systemPrompt, err := provider.convertToAnthropicFormat(tt.messages)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedSystemPrompt, systemPrompt)
			assert.Len(t, anthropicMessages, tt.expectedMessageCount)

			if tt.expectedMessageCount > 0 {
				assert.Equal(t, tt.expectedFirstRole, anthropicMessages[0].Role)
				assert.Equal(t, tt.expectedFirstContent, anthropicMessages[0].Content)
			}
		})
	}
}

func TestAnthropicProvider_MapErrorCode(t *testing.T) {
	provider := NewAnthropicProvider("test-key", "")

	tests := []struct {
		name           string
		statusCode     int
		message        string
		expectedPrefix string
	}{
		{
			name:           "unauthorized",
			statusCode:     401,
			message:        "Invalid API key",
			expectedPrefix: "INVALID_API_KEY",
		},
		{
			name:           "rate limited",
			statusCode:     429,
			message:        "Too many requests",
			expectedPrefix: "RATE_LIMITED",
		},
		{
			name:           "context too long",
			statusCode:     400,
			message:        "context length exceeded",
			expectedPrefix: "CONTEXT_TOO_LONG",
		},
		{
			name:           "bad request",
			statusCode:     400,
			message:        "Invalid request",
			expectedPrefix: "BAD_REQUEST",
		},
		{
			name:           "server error",
			statusCode:     500,
			message:        "Internal server error",
			expectedPrefix: "PROVIDER_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.mapErrorCode(tt.statusCode, tt.message)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedPrefix)
		})
	}
}

func TestAnthropicProvider_ProcessAnthropicEvent(t *testing.T) {
	provider := NewAnthropicProvider("test-key", "")
	
	tests := []struct {
		name          string
		eventType     string
		data          string
		expectContent bool
		expectError   bool
	}{
		{
			name:      "message_start event",
			eventType: "message_start",
			data:      `{"type":"message_start","message":{"id":"msg_123","type":"message","role":"assistant","content":[],"model":"claude-3-5-sonnet-20241022","usage":{"input_tokens":10,"output_tokens":0}}}`,
			expectContent: false,
			expectError:   false,
		},
		{
			name:      "content_block_delta event",
			eventType: "content_block_delta",
			data:      `{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}`,
			expectContent: true,
			expectError:   false,
		},
		{
			name:      "message_delta event",
			eventType: "message_delta",
			data:      `{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":15}}`,
			expectContent: false,
			expectError:   false,
		},
		{
			name:      "message_stop event",
			eventType: "message_stop",
			data:      `{"type":"message_stop"}`,
			expectContent: false,
			expectError:   false,
		},
		{
			name:      "error event",
			eventType: "error",
			data:      `{"type":"error","error":{"type":"invalid_request_error","message":"Invalid request"}}`,
			expectContent: false,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventChan := make(chan models.StreamEvent, 10)
			var totalUsage *models.TokenUsage

			provider.processAnthropicEvent(tt.eventType, tt.data, eventChan, &totalUsage)
			close(eventChan)

			// Collect events
			var events []models.StreamEvent
			for event := range eventChan {
				events = append(events, event)
			}

			if tt.expectContent {
				// Should have at least one content event
				hasContent := false
				for _, event := range events {
					if event.Type == "content" && event.Content != "" {
						hasContent = true
						break
					}
				}
				assert.True(t, hasContent, "Expected content event")
			}

			if tt.expectError {
				// Should have an error event
				hasError := false
				for _, event := range events {
					if event.Type == "error" {
						hasError = true
						break
					}
				}
				assert.True(t, hasError, "Expected error event")
			}
		})
	}
}
