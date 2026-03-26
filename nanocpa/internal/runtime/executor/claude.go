package executor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/auth"
	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/translator"
)

type Claude struct {
	client *http.Client
}

const defaultUpstreamTimeout = 30 * time.Second

func NewClaude(client *http.Client) *Claude {
	if client == nil {
		client = &http.Client{Timeout: defaultUpstreamTimeout}
	} else if client.Timeout <= 0 {
		cloned := *client
		cloned.Timeout = defaultUpstreamTimeout
		client = &cloned
	}
	return &Claude{
		client: client,
	}
}

func (c *Claude) Execute(ctx context.Context, openAIRequest []byte, runtimeAuth *auth.Auth) (*Result, error) {
	if runtimeAuth == nil {
		return nil, errors.New("runtime auth is required")
	}

	translatedRequest, err := translator.OpenAIChatToClaudeRequest(openAIRequest)
	if err != nil {
		return nil, fmt.Errorf("translate openai request: %w", err)
	}

	baseURL := strings.TrimSuffix(strings.TrimSpace(runtimeAuth.Attributes["base_url"]), "/")
	if baseURL == "" {
		return nil, errors.New("runtime auth missing base_url")
	}
	apiKey := strings.TrimSpace(runtimeAuth.Attributes["api_key"])
	if apiKey == "" {
		return nil, errors.New("runtime auth missing api_key")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/messages", bytes.NewReader(translatedRequest))
	if err != nil {
		return nil, fmt.Errorf("create upstream request: %w", err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send upstream request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read upstream response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &UpstreamError{
			StatusCode: resp.StatusCode,
			Body:       responseBody,
			Headers:    resp.Header.Clone(),
		}
	}

	payload, err := translator.ClaudeResponseToOpenAIResponse(responseBody)
	if err != nil {
		return nil, fmt.Errorf("translate claude response: %w", err)
	}

	return &Result{
		StatusCode: resp.StatusCode,
		Body:       payload,
		Headers:    translatedResponseHeaders(resp.Header),
	}, nil
}

func translatedResponseHeaders(upstream http.Header) http.Header {
	headers := make(http.Header)
	headers.Set("content-type", "application/json")
	if requestID := strings.TrimSpace(upstream.Get("x-request-id")); requestID != "" {
		headers.Set("x-request-id", requestID)
	}
	return headers
}
