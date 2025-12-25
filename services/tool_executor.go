package services

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"Curry2API-go/models"
	"github.com/sirupsen/logrus"
)

// ToolExecutor 工具执行器 - 处理 Claude Code CLI 的工具调用
// 由于 Cursor API 不支持原生工具调用，我们通过以下方式实现：
// 1. 将工具定义注入到系统提示中
// 2. 让 Claude 以特定格式输出工具调用意图
// 3. 解析输出并转换为标准的 tool_use 响应格式
type ToolExecutor struct {
	// 工具定义缓存
	toolDefinitions map[string]*models.ClaudeTool
}

// NewToolExecutor 创建新的工具执行器
func NewToolExecutor() *ToolExecutor {
	return &ToolExecutor{
		toolDefinitions: make(map[string]*models.ClaudeTool),
	}
}

// HasToolUse 检查请求是否包含工具调用
func (te *ToolExecutor) HasToolUse(request *models.ClaudeMessageRequest) bool {
	return len(request.Tools) > 0
}

// IsAnthropicBuiltinTool 检查是否是 Anthropic 内置工具
func (te *ToolExecutor) IsAnthropicBuiltinTool(tool *models.ClaudeTool) bool {
	builtinTypes := []string{
		"text_editor_20250728",
		"text_editor_20250429",
		"text_editor_20250124",
		"text_editor_20241022",
		"bash_20250124",
		"bash_20241022",
		"computer_20250124",
		"computer_20241022",
	}
	
	for _, t := range builtinTypes {
		if tool.Type == t {
			return true
		}
	}
	return false
}

// BuildToolSystemPrompt 构建工具系统提示
// 注意：Claude Code CLI 已经在系统提示中包含了完整的工具说明
// 我们只需要添加输出格式说明，让模型知道如何输出工具调用
func (te *ToolExecutor) BuildToolSystemPrompt(tools []models.ClaudeTool) string {
	// 缓存工具定义
	for _, tool := range tools {
		te.toolDefinitions[te.getToolName(tool)] = &tool
	}
	
	// 构建工具列表
	var toolList strings.Builder
	toolList.WriteString("\n\n# Tool Calling Instructions\n\n")
	toolList.WriteString("You have access to the following tools:\n\n")
	
	for _, tool := range tools {
		name := te.getToolName(tool)
		toolList.WriteString(fmt.Sprintf("- **%s**", name))
		if tool.Description != "" {
			toolList.WriteString(fmt.Sprintf(": %s", tool.Description))
		}
		toolList.WriteString("\n")
	}
	
	toolList.WriteString(`
## How to use tools

When you need to use a tool, you MUST output it in this EXACT format:

<tool_call>
<tool_name>TOOL_NAME_HERE</tool_name>
<tool_input>
{"param1": "value1", "param2": "value2"}
</tool_input>
</tool_call>

IMPORTANT RULES:
1. Replace TOOL_NAME_HERE with the exact tool name from the list above
2. The tool_input MUST be valid JSON
3. Only output ONE tool call at a time
4. After outputting a tool call, STOP and wait for the result
5. Do NOT include any text after the </tool_call> tag
`)
	
	return toolList.String()
}

// getToolName 获取工具名称
func (te *ToolExecutor) getToolName(tool models.ClaudeTool) string {
	if tool.Name != "" {
		return tool.Name
	}
	return tool.Type
}

// getBuiltinToolDescription 获取内置工具的描述
func (te *ToolExecutor) getBuiltinToolDescription(toolType string) string {
	switch {
	case strings.HasPrefix(toolType, "text_editor"):
		return `Text editor tool for file operations.

To VIEW a file:
<tool_call>
<tool_name>` + toolType + `</tool_name>
<tool_input>
{"command": "view", "path": "/path/to/file"}
</tool_input>
</tool_call>

To CREATE a file:
<tool_call>
<tool_name>` + toolType + `</tool_name>
<tool_input>
{"command": "create", "path": "/path/to/file", "file_text": "file content here"}
</tool_input>
</tool_call>

To REPLACE text in a file:
<tool_call>
<tool_name>` + toolType + `</tool_name>
<tool_input>
{"command": "str_replace", "path": "/path/to/file", "old_str": "text to find", "new_str": "replacement text"}
</tool_input>
</tool_call>

To INSERT text at a line:
<tool_call>
<tool_name>` + toolType + `</tool_name>
<tool_input>
{"command": "insert", "path": "/path/to/file", "insert_line": 10, "new_str": "text to insert"}
</tool_input>
</tool_call>
`
	case strings.HasPrefix(toolType, "bash"):
		return `Bash tool for executing shell commands.

To run a command:
<tool_call>
<tool_name>` + toolType + `</tool_name>
<tool_input>
{"command": "ls -la", "timeout": 30}
</tool_input>
</tool_call>
`
	case strings.HasPrefix(toolType, "computer"):
		return `Computer use tool for GUI automation.
Input: {"action": "action_type", ...action_specific_params}
`
	default:
		return ""
	}
}

// ParseToolCallFromResponse 从 Claude 响应中解析工具调用
func (te *ToolExecutor) ParseToolCallFromResponse(content string) (*models.ClaudeToolUse, string, bool) {
	// 匹配 <tool_call>...</tool_call> 格式
	toolCallRegex := regexp.MustCompile(`(?s)<tool_call>\s*<tool_name>([^<]+)</tool_name>\s*<tool_input>\s*(.*?)\s*</tool_input>\s*</tool_call>`)
	
	matches := toolCallRegex.FindStringSubmatch(content)
	if len(matches) < 3 {
		return nil, content, false
	}
	
	toolName := strings.TrimSpace(matches[1])
	toolInputStr := strings.TrimSpace(matches[2])
	
	// 解析 JSON 输入
	var toolInput map[string]interface{}
	if err := json.Unmarshal([]byte(toolInputStr), &toolInput); err != nil {
		logrus.WithError(err).Warn("Failed to parse tool input JSON")
		// 尝试清理 JSON
		toolInputStr = strings.ReplaceAll(toolInputStr, "\n", "")
		if err := json.Unmarshal([]byte(toolInputStr), &toolInput); err != nil {
			return nil, content, false
		}
	}
	
	// 生成工具调用 ID
	toolUseID := fmt.Sprintf("toolu_%s", generateRandomID(24))
	
	toolUse := &models.ClaudeToolUse{
		Type:  "tool_use",
		ID:    toolUseID,
		Name:  toolName,
		Input: toolInput,
	}
	
	// 提取工具调用之前的文本
	beforeText := strings.TrimSpace(content[:strings.Index(content, "<tool_call>")])
	
	return toolUse, beforeText, true
}

// generateRandomID 生成随机 ID
func generateRandomID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	// 使用时间戳作为种子来生成伪随机数
	seed := uint64(time.Now().UnixNano())
	for i := range result {
		seed = seed*1103515245 + 12345
		idx := int(seed>>16) % len(charset)
		result[i] = charset[idx]
	}
	return string(result)
}

// getToolChoicePrompt 根据 tool_choice 生成额外的提示
func (te *ToolExecutor) getToolChoicePrompt(toolChoice interface{}) string {
	if toolChoice == nil {
		return ""
	}
	
	// 处理不同的 tool_choice 格式
	switch tc := toolChoice.(type) {
	case string:
		switch tc {
		case "any", "required":
			return "\n\n**MANDATORY**: You MUST use a tool in your response. Do not respond with just text - you MUST output a <tool_call> block.\n"
		case "auto":
			return "\n\n**IMPORTANT**: Use tools when appropriate to complete the task.\n"
		case "none":
			return "" // 不使用工具
		}
	case map[string]interface{}:
		if typeVal, ok := tc["type"].(string); ok {
			switch typeVal {
			case "any", "required":
				return "\n\n**MANDATORY**: You MUST use a tool in your response. Do not respond with just text - you MUST output a <tool_call> block.\n"
			case "auto":
				return "\n\n**IMPORTANT**: Use tools when appropriate to complete the task.\n"
			case "tool":
				// 指定特定工具
				if name, ok := tc["name"].(string); ok {
					return fmt.Sprintf("\n\n**MANDATORY**: You MUST use the tool '%s' in your response. Output a <tool_call> block with tool_name='%s'.\n", name, name)
				}
			}
		}
	}
	
	return ""
}

// ConvertToolResultToMessage 将工具结果转换为消息格式
func (te *ToolExecutor) ConvertToolResultToMessage(toolResult *models.ClaudeToolResult) string {
	var content string
	
	switch c := toolResult.Content.(type) {
	case string:
		content = c
	case []interface{}:
		// 处理内容块数组
		var texts []string
		for _, item := range c {
			if block, ok := item.(map[string]interface{}); ok {
				if text, exists := block["text"].(string); exists {
					texts = append(texts, text)
				}
			}
		}
		content = strings.Join(texts, "\n")
	default:
		contentBytes, _ := json.Marshal(c)
		content = string(contentBytes)
	}
	
	if toolResult.IsError {
		return fmt.Sprintf("<tool_result tool_use_id=\"%s\" is_error=\"true\">\n%s\n</tool_result>", 
			toolResult.ToolUseID, content)
	}
	
	return fmt.Sprintf("<tool_result tool_use_id=\"%s\">\n%s\n</tool_result>", 
		toolResult.ToolUseID, content)
}

// InjectToolPrompt 将工具提示注入到请求中
func (te *ToolExecutor) InjectToolPrompt(request *models.ClaudeMessageRequest) *models.ClaudeMessageRequest {
	if !te.HasToolUse(request) {
		return request
	}
	
	// 记录工具信息用于调试
	toolNames := make([]string, 0, len(request.Tools))
	for _, tool := range request.Tools {
		toolNames = append(toolNames, te.getToolName(tool))
	}
	logrus.WithFields(logrus.Fields{
		"tool_count": len(request.Tools),
		"tool_names": toolNames,
	}).Debug("Processing tools from request")
	
	// 构建工具系统提示
	toolPrompt := te.BuildToolSystemPrompt(request.Tools)
	
	// 检查 tool_choice 是否强制使用工具
	toolChoicePrompt := te.getToolChoicePrompt(request.ToolChoice)
	if toolChoicePrompt != "" {
		toolPrompt = toolChoicePrompt + toolPrompt
	}
	
	// 注入到系统提示中
	switch sys := request.System.(type) {
	case string:
		// 记录原始系统提示长度
		logrus.WithField("original_system_length", len(sys)).Debug("Injecting tool prompt into string system")
		request.System = sys + toolPrompt
	case nil:
		request.System = toolPrompt
	case []interface{}:
		// 处理数组格式的系统提示 - Claude Code CLI 使用这种格式
		// 记录系统提示块数量
		logrus.WithFields(logrus.Fields{
			"system_blocks": len(sys),
			"tool_count":    len(request.Tools),
		}).Debug("Injecting tool prompt into system array")
		
		// 添加一个新的文本块包含工具提示
		newBlock := map[string]interface{}{
			"type": "text",
			"text": toolPrompt,
		}
		request.System = append(sys, newBlock)
	case []map[string]interface{}:
		// 处理已解析的数组格式
		newBlock := map[string]interface{}{
			"type": "text",
			"text": toolPrompt,
		}
		newSys := make([]interface{}, len(sys)+1)
		for i, block := range sys {
			newSys[i] = block
		}
		newSys[len(sys)] = newBlock
		request.System = newSys
	default:
		// 如果是其他类型，转换为字符串
		request.System = fmt.Sprintf("%v%s", sys, toolPrompt)
	}
	
	// tool_result 的处理已经移到 models/claude.go 的 ToOpenAIRequest 中
	// 这里不再重复处理，避免格式冲突
	
	return request
}

// appendToolReminder 在最后一条用户消息后添加工具使用提醒
// 注意：这个提醒可能会干扰正常对话，所以我们暂时禁用它
func (te *ToolExecutor) appendToolReminder(request *models.ClaudeMessageRequest) {
	// 暂时禁用工具提醒，因为它可能导致重复回答
	// 工具指令已经在系统提示中了
	return
}
