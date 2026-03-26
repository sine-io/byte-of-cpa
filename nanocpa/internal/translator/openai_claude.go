package translator

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type openAIChatRequest struct {
	Model    string               `json:"model"`
	Messages []openAIInputMessage `json:"messages"`
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
	for _, block := range in.Content {
		if block.Type == "text" {
			if assistantText != "" {
				assistantText += "\n"
			}
			assistantText += block.Text
		}
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
