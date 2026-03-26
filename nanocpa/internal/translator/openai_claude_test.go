package translator

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestOpenAIChatToClaudeRequest_MapsOpenAIChatToClaudeMessages(t *testing.T) {
	t.Parallel()

	openAIRequest := []byte(`{
		"model":"claude-3-7-sonnet",
		"messages":[
			{"role":"user","content":"hello"},
			{"role":"assistant","content":"hi there"}
		]
	}`)

	got, err := OpenAIChatToClaudeRequest(openAIRequest)
	if err != nil {
		t.Fatalf("translate request: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(got, &decoded); err != nil {
		t.Fatalf("decode translated request: %v", err)
	}

	if decoded["model"] != "claude-3-7-sonnet" {
		t.Fatalf("expected model preserved, got %#v", decoded["model"])
	}
	if decoded["max_tokens"] != float64(1024) {
		t.Fatalf("expected narrow default max_tokens, got %#v", decoded["max_tokens"])
	}

	messages, ok := decoded["messages"].([]any)
	if !ok || len(messages) != 2 {
		t.Fatalf("expected two translated messages, got %#v", decoded["messages"])
	}

	first, ok := messages[0].(map[string]any)
	if !ok {
		t.Fatalf("expected object message, got %#v", messages[0])
	}
	if first["role"] != "user" || first["content"] != "hello" {
		t.Fatalf("unexpected first message: %#v", first)
	}

	second, ok := messages[1].(map[string]any)
	if !ok {
		t.Fatalf("expected object message, got %#v", messages[1])
	}
	if second["role"] != "assistant" || second["content"] != "hi there" {
		t.Fatalf("unexpected second message: %#v", second)
	}
}

func TestOpenAIChatToClaudeRequest_SystemRoleMappedToTopLevelSystem(t *testing.T) {
	t.Parallel()

	openAIRequest := []byte(`{
		"model":"claude-3-7-sonnet",
		"messages":[
			{"role":"system","content":"be concise"},
			{"role":"user","content":"hello"}
		]
	}`)

	got, err := OpenAIChatToClaudeRequest(openAIRequest)
	if err != nil {
		t.Fatalf("translate request: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(got, &decoded); err != nil {
		t.Fatalf("decode translated request: %v", err)
	}
	if decoded["system"] != "be concise" {
		t.Fatalf("expected top-level system prompt, got %#v", decoded["system"])
	}

	messages, ok := decoded["messages"].([]any)
	if !ok || len(messages) != 1 {
		t.Fatalf("expected one non-system message, got %#v", decoded["messages"])
	}
}

func TestOpenAIChatToClaudeRequest_RejectsUnsupportedContentShape(t *testing.T) {
	t.Parallel()

	openAIRequest := []byte(`{
		"model":"claude-3-7-sonnet",
		"messages":[{"role":"user","content":[{"type":"text","text":"hello"}]}]
	}`)

	_, err := OpenAIChatToClaudeRequest(openAIRequest)
	if err == nil {
		t.Fatal("expected validation error for unsupported content shape")
	}
	var validationErr *ValidationError
	if !AsValidationError(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if validationErr.StatusCode != 400 {
		t.Fatalf("expected 400 status code, got %d", validationErr.StatusCode)
	}
	if !strings.Contains(validationErr.Message, "unsupported") {
		t.Fatalf("expected unsupported message, got %q", validationErr.Message)
	}
}

func TestOpenAIChatToClaudeRequest_RejectsUnsupportedFeatures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		requestBody string
		wantMessage string
	}{
		{
			name:        "streaming",
			requestBody: `{"model":"claude-3-7-sonnet","stream":true,"messages":[{"role":"user","content":"hello"}]}`,
			wantMessage: "stream",
		},
		{
			name:        "multiple choices",
			requestBody: `{"model":"claude-3-7-sonnet","n":2,"messages":[{"role":"user","content":"hello"}]}`,
			wantMessage: "n",
		},
		{
			name:        "tools",
			requestBody: `{"model":"claude-3-7-sonnet","tools":[{"type":"function","function":{"name":"lookup"}}],"messages":[{"role":"user","content":"hello"}]}`,
			wantMessage: "tools",
		},
		{
			name:        "tool choice",
			requestBody: `{"model":"claude-3-7-sonnet","tool_choice":"auto","messages":[{"role":"user","content":"hello"}]}`,
			wantMessage: "tool_choice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := OpenAIChatToClaudeRequest([]byte(tt.requestBody))
			if err == nil {
				t.Fatal("expected validation error")
			}

			var validationErr *ValidationError
			if !AsValidationError(err, &validationErr) {
				t.Fatalf("expected ValidationError, got %T", err)
			}
			if validationErr.StatusCode != 400 {
				t.Fatalf("expected 400 status code, got %d", validationErr.StatusCode)
			}
			if !strings.Contains(validationErr.Message, tt.wantMessage) {
				t.Fatalf("expected message containing %q, got %q", tt.wantMessage, validationErr.Message)
			}
		})
	}
}

func TestOpenAIChatToClaudeRequest_RejectsInvalidRole(t *testing.T) {
	t.Parallel()

	openAIRequest := []byte(`{
		"model":"claude-3-7-sonnet",
		"messages":[{"role":"tool","content":"invalid"}]
	}`)

	_, err := OpenAIChatToClaudeRequest(openAIRequest)
	if err == nil {
		t.Fatal("expected validation error for invalid role")
	}
	var validationErr *ValidationError
	if !AsValidationError(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if validationErr.StatusCode != 400 {
		t.Fatalf("expected 400 status code, got %d", validationErr.StatusCode)
	}
	if !strings.Contains(validationErr.Message, "role") {
		t.Fatalf("expected role message, got %q", validationErr.Message)
	}
}

func TestOpenAIChatToClaudeRequest_RejectsOnlySystemMessages(t *testing.T) {
	t.Parallel()

	openAIRequest := []byte(`{
		"model":"claude-3-7-sonnet",
		"messages":[{"role":"system","content":"instructions only"}]
	}`)

	_, err := OpenAIChatToClaudeRequest(openAIRequest)
	if err == nil {
		t.Fatal("expected validation error for empty effective messages")
	}
	var validationErr *ValidationError
	if !AsValidationError(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if validationErr.StatusCode != 400 {
		t.Fatalf("expected 400 status code, got %d", validationErr.StatusCode)
	}
	if !strings.Contains(validationErr.Message, "at least one") {
		t.Fatalf("expected at least one message hint, got %q", validationErr.Message)
	}
}

func TestClaudeResponseToOpenAIResponse_MapsToOpenAIShape(t *testing.T) {
	t.Parallel()

	input := []byte(`{
		"id":"msg_1",
		"model":"claude-3-7-sonnet",
		"content":[{"type":"text","text":"hello from claude"}]
	}`)
	got, err := ClaudeResponseToOpenAIResponse(input)
	if err != nil {
		t.Fatalf("map response: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(got, &decoded); err != nil {
		t.Fatalf("decode openai response: %v", err)
	}
	if decoded["object"] != "chat.completion" {
		t.Fatalf("expected chat.completion object, got %#v", decoded["object"])
	}

	choices, ok := decoded["choices"].([]any)
	if !ok || len(choices) != 1 {
		t.Fatalf("expected one choice, got %#v", decoded["choices"])
	}
	choice, ok := choices[0].(map[string]any)
	if !ok {
		t.Fatalf("expected choice object, got %#v", choices[0])
	}
	message, ok := choice["message"].(map[string]any)
	if !ok {
		t.Fatalf("expected message object, got %#v", choice["message"])
	}
	if message["role"] != "assistant" || message["content"] != "hello from claude" {
		t.Fatalf("unexpected assistant message: %#v", message)
	}
}

func TestClaudeResponseToOpenAIResponse_RejectsUnsupportedOrEmptyContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       []byte
		wantMessage string
	}{
		{
			name: "unsupported block type",
			input: []byte(`{
				"id":"msg_1",
				"model":"claude-3-7-sonnet",
				"content":[{"type":"tool_use","text":"ignored"}]
			}`),
			wantMessage: "unsupported",
		},
		{
			name: "empty text content",
			input: []byte(`{
				"id":"msg_1",
				"model":"claude-3-7-sonnet",
				"content":[{"type":"text","text":"   "}]
			}`),
			wantMessage: "no usable text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ClaudeResponseToOpenAIResponse(tt.input)
			if err == nil {
				t.Fatal("expected response translation error")
			}
			if !strings.Contains(err.Error(), tt.wantMessage) {
				t.Fatalf("expected error containing %q, got %q", tt.wantMessage, err.Error())
			}
		})
	}
}
