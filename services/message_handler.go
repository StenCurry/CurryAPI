package services

import (
	"Curry2API-go/models"
	"strings"
)

// messageHandler 处理消息的转换和截断
type messageHandler struct {
	service *CursorService
}

// truncateMessages 截断消息历史以适应长度限制
// 算法：保留系统消息，从最后向前收集消息直到达到限制
func (m *messageHandler) truncateMessages(messages []models.Message) []models.Message {
	if len(messages) == 0 || m.service.config.MaxInputLength <= 0 {
		return messages
	}

	maxLength := m.service.config.MaxInputLength
	total := 0
	for _, msg := range messages {
		total += len(msg.GetStringContent())
	}

	// 如果总长度未超限，直接返回
	if total <= maxLength {
		return messages
	}

	var result []models.Message
	startIdx := 0

	// 保留系统消息（如果存在）
	if strings.EqualFold(messages[0].Role, "system") {
		result = append(result, messages[0])
		maxLength -= len(messages[0].GetStringContent())
		if maxLength < 0 {
			maxLength = 0
		}
		startIdx = 1
	}

	// 从最后一条消息开始，向前收集
	current := 0
	collected := make([]models.Message, 0, len(messages)-startIdx)
	for i := len(messages) - 1; i >= startIdx; i-- {
		msg := messages[i]
		msgLen := len(msg.GetStringContent())
		if msgLen == 0 {
			continue
		}
		if current+msgLen > maxLength {
			continue
		}
		collected = append(collected, msg)
		current += msgLen
	}

	// 反转顺序（因为是从后向前收集的）
	for i, j := 0, len(collected)-1; i < j; i, j = i+1, j-1 {
		collected[i], collected[j] = collected[j], collected[i]
	}

	return append(result, collected...)
}
