package database

import (
	"database/sql"
	"errors"
	"time"

	"Curry2API-go/models"
)

// Chat system errors
var (
	ErrConversationNotFound = errors.New("conversation not found")
	ErrMessageNotFound      = errors.New("message not found")
)

// CreateConversation creates a new chat conversation for a user
// Requirements: 1.1
func CreateConversation(userID int64, title, model string) (*models.Conversation, error) {
	now := time.Now()

	result, err := db.Exec(
		`INSERT INTO chat_conversations (user_id, title, model, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?)`,
		userID, title, model, now, now,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.Conversation{
		ID:        id,
		UserID:    userID,
		Title:     title,
		Model:     model,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// GetConversations retrieves paginated conversations for a user, sorted by updated_at DESC
// Requirements: 1.2, 7.3
func GetConversations(userID int64, page, limit int) ([]models.Conversation, int, error) {
	// Calculate offset
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	// Get total count
	var total int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM chat_conversations WHERE user_id = ?`,
		userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get conversations sorted by updated_at DESC
	rows, err := db.Query(
		`SELECT id, user_id, title, model, COALESCE(system_prompt, ''), created_at, updated_at
		 FROM chat_conversations 
		 WHERE user_id = ? 
		 ORDER BY updated_at DESC 
		 LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// Initialize as empty slice to ensure JSON serializes to [] instead of null
	conversations := make([]models.Conversation, 0)
	for rows.Next() {
		var conv models.Conversation
		err := rows.Scan(&conv.ID, &conv.UserID, &conv.Title, &conv.Model,
			&conv.SystemPrompt, &conv.CreatedAt, &conv.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		conversations = append(conversations, conv)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return conversations, total, nil
}

// GetConversation retrieves a single conversation by ID for a specific user
// Requirements: 1.3
func GetConversation(id, userID int64) (*models.Conversation, error) {
	conv := &models.Conversation{}

	err := db.QueryRow(
		`SELECT id, user_id, title, model, COALESCE(system_prompt, ''), created_at, updated_at
		 FROM chat_conversations 
		 WHERE id = ? AND user_id = ?`,
		id, userID,
	).Scan(&conv.ID, &conv.UserID, &conv.Title, &conv.Model,
		&conv.SystemPrompt, &conv.CreatedAt, &conv.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrConversationNotFound
	}
	if err != nil {
		return nil, err
	}

	return conv, nil
}

// UpdateConversation updates a conversation's title and/or model
// Requirements: 1.5
func UpdateConversation(id, userID int64, title, model string) error {
	result, err := db.Exec(
		`UPDATE chat_conversations 
		 SET title = ?, model = ?, updated_at = ?
		 WHERE id = ? AND user_id = ?`,
		title, model, time.Now(), id, userID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrConversationNotFound
	}

	return nil
}

// DeleteConversation deletes a conversation and all its messages (cascade)
// Requirements: 1.4
func DeleteConversation(id, userID int64) error {
	// The foreign key constraint with ON DELETE CASCADE will automatically
	// delete all associated messages when the conversation is deleted
	result, err := db.Exec(
		`DELETE FROM chat_conversations WHERE id = ? AND user_id = ?`,
		id, userID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrConversationNotFound
	}

	return nil
}

// CreateMessage creates a new message in a conversation
// Requirements: 2.1
func CreateMessage(conversationID int64, role, content string, tokens int, cost float64) (*models.ChatMessage, error) {
	now := time.Now()

	// Start transaction to update conversation's updated_at as well
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Insert message
	result, err := tx.Exec(
		`INSERT INTO chat_messages (conversation_id, role, content, tokens, cost, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		conversationID, role, content, tokens, cost, now,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Update conversation's updated_at
	_, err = tx.Exec(
		`UPDATE chat_conversations SET updated_at = ? WHERE id = ?`,
		now, conversationID,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &models.ChatMessage{
		ID:             id,
		ConversationID: conversationID,
		Role:           role,
		Content:        content,
		Tokens:         tokens,
		Cost:           cost,
		CreatedAt:      now,
	}, nil
}

// GetMessages retrieves paginated messages for a conversation, sorted by created_at ASC
// Requirements: 1.3, 7.2
func GetMessages(conversationID int64, page, limit int) ([]models.ChatMessage, int, error) {
	// Calculate offset
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	// Get total count
	var total int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM chat_messages WHERE conversation_id = ?`,
		conversationID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get messages sorted by created_at ASC (chronological order)
	rows, err := db.Query(
		`SELECT id, conversation_id, role, content, tokens, cost, created_at
		 FROM chat_messages 
		 WHERE conversation_id = ? 
		 ORDER BY created_at ASC 
		 LIMIT ? OFFSET ?`,
		conversationID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// Initialize as empty slice to ensure JSON serializes to [] instead of null
	messages := make([]models.ChatMessage, 0)
	for rows.Next() {
		var msg models.ChatMessage
		err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content,
			&msg.Tokens, &msg.Cost, &msg.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// GetAllMessages retrieves all messages for a conversation (for context building)
// Requirements: 2.3
func GetAllMessages(conversationID int64) ([]models.ChatMessage, error) {
	rows, err := db.Query(
		`SELECT id, conversation_id, role, content, tokens, cost, created_at
		 FROM chat_messages 
		 WHERE conversation_id = ? 
		 ORDER BY created_at ASC`,
		conversationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Initialize as empty slice to ensure JSON serializes to [] instead of null
	messages := make([]models.ChatMessage, 0)
	for rows.Next() {
		var msg models.ChatMessage
		err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content,
			&msg.Tokens, &msg.Cost, &msg.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

// UpdateConversationTimestamp updates only the updated_at timestamp of a conversation
func UpdateConversationTimestamp(conversationID int64) error {
	_, err := db.Exec(
		`UPDATE chat_conversations SET updated_at = ? WHERE id = ?`,
		time.Now(), conversationID,
	)
	return err
}

// ConversationBelongsToUser checks if a conversation belongs to a specific user
func ConversationBelongsToUser(conversationID, userID int64) (bool, error) {
	var exists bool
	err := db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM chat_conversations WHERE id = ? AND user_id = ?)`,
		conversationID, userID,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
