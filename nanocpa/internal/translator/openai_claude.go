package translator

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type openAIChatRequest struct {
	Model      string               `json:"model"`
	Messages   []openAIInputMessage `json:"messages"`
	Stream     bool                 `json:"stream,omitempty"`
	N          *int                 `json:"n,omitempty"`
	Tools      json.RawMessage      `json:"tools,omitempty"`
	ToolChoice json.RawMessage      `json:"tool_choice,omitempty"`
}

type openAIInputMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type claudeMessagesRequest struct {
	Model     string              `json:"model"`
	System    string              `json:"system,omitempty"`
	Messages  []claudeChatMessage `json:"messages"`
	MaxTokens int                 `json:"max_tokens"`
}

type claudeChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ValidationError struct {
	StatusCode int
	Message    string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func AsValidationError(err error, target **ValidationError) bool {
	return errors.As(err, target)
}

func OpenAIChatToClaudeRequest(openAIRequest []byte) ([]byte, error) {
	var in openAIChatRequest
	if err := json.Unmarshal(openAIRequest, &in); err != nil {
		return nil, &ValidationError{
			StatusCode: 400,
			Message:    fmt.Sprintf("invalid JSON request body: %v", err),
		}
	}
	if in.Model == "" {
		return nil, &ValidationError{
			StatusCode: 400,
			Message:    "model is required",
		}
	}
	if in.Stream {
		return nil, &ValidationError{
			StatusCode: 400,
			Message:    "stream=true is not supported for Claude chat completions in Chapter 7",
		}
	}
	if in.N != nil && *in.N != 1 {
		return nil, &ValidationError{
			StatusCode: 400,
			Message:    fmt.Sprintf("n=%d is not supported; only n=1 is allowed for Claude chat completions in Chapter 7", *in.N),
		}
	}
	if hasJSONValue(in.Tools) {
		return nil, &ValidationError{
			StatusCode: 400,
			Message:    "tools are not supported for Claude chat completions in Chapter 7",
		}
	}
	if hasJSONValue(in.ToolChoice) {
		return nil, &ValidationError{
			StatusCode: 400,
			Message:    "tool_choice is not supported for Claude chat completions in Chapter 7",
		}
	}

	out := claudeMessagesRequest{
		Model:     in.Model,
		Messages:  make([]claudeChatMessage, 0, len(in.Messages)),
		MaxTokens: 1024,
	}

	var systemPrompts []string
	for i, msg := range in.Messages {
		switch msg.Role {
		case "system", "user", "assistant":
		default:
			return nil, &ValidationError{
				StatusCode: 400,
				Message:    fmt.Sprintf("unsupported messages[%d].role %q", i, msg.Role),
			}
		}
		textContent, ok := msg.Content.(string)
		if !ok {
			return nil, &ValidationError{
				StatusCode: 400,
				Message:    fmt.Sprintf("unsupported messages[%d].content shape; only string content is supported", i),
			}
		}
		if msg.Role == "system" {
			systemPrompts = append(systemPrompts, textContent)
			continue
		}
		out.Messages = append(out.Messages, claudeChatMessage{
			Role:    msg.Role,
			Content: textContent,
		})
	}
	out.System = strings.Join(systemPrompts, "\n")
	if len(out.Messages) == 0 {
		return nil, &ValidationError{
			StatusCode: 400,
			Message:    "at least one non-system message is required",
		}
	}

	b, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("encode claude request: %w", err)
	}
	return b, nil
}

func ClaudeResponseToOpenAIResponse(claudeResponse []byte) ([]byte, error) {
	var in struct {
		ID      string `json:"id"`
		Model   string `json:"model"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(claudeResponse, &in); err != nil {
		return nil, fmt.Errorf("decode claude response: %w", err)
	}

	var assistantText string
	hasUsableText := false
	for i, block := range in.Content {
		if block.Type != "text" {
			return nil, fmt.Errorf("unsupported claude response content[%d].type %q", i, block.Type)
		}
		if strings.TrimSpace(block.Text) == "" {
			continue
		}
		if assistantText != "" {
			assistantText += "\n"
		}
		assistantText += block.Text
		hasUsableText = true
	}
	if !hasUsableText {
		return nil, errors.New("claude response content contains no usable text blocks")
	}

	out := map[string]any{
		"id":     in.ID,
		"object": "chat.completion",
		"model":  in.Model,
		"choices": []map[string]any{
			{
				"index": 0,
				"message": map[string]any{
					"role":    "assistant",
					"content": assistantText,
				},
				"finish_reason": "stop",
			},
		},
	}
	return json.Marshal(out)
}

func hasJSONValue(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	return !bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}
