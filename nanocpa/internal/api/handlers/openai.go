package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type OpenAI struct{}

const maxChatCompletionsRequestBodyBytes int64 = 4 * 1024 * 1024

func NewOpenAI() *OpenAI {
	return &OpenAI{}
}

func (h *OpenAI) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/chat/completions", h.ChatCompletions)
	mux.HandleFunc("GET /v1/models", h.Models)
}

func (h *OpenAI) ChatCompletions(w http.ResponseWriter, r *http.Request) {
	limitedBody := http.MaxBytesReader(w, r.Body, maxChatCompletionsRequestBodyBytes)
	requestBody, err := io.ReadAll(limitedBody)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			writeOpenAIError(w, http.StatusRequestEntityTooLarge, fmt.Sprintf("request body too large (max %d bytes)", maxChatCompletionsRequestBodyBytes), "invalid_request_error")
			return
		}
		writeOpenAIError(w, http.StatusBadRequest, "invalid request body", "invalid_request_error")
		return
	}

	var req struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal(requestBody, &req); err != nil {
		writeOpenAIError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON request body: %v", err), "invalid_request_error")
		return
	}

	if req.Model == "" {
		writeOpenAIError(w, http.StatusBadRequest, "model is required", "invalid_request_error")
		return
	}

	writeOpenAIError(w, http.StatusBadGateway, "upstream provider request failed", "api_error")
}

func (h *OpenAI) Models(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	type model struct {
		ID     string `json:"id"`
		Object string `json:"object"`
	}
	response := struct {
		Object string  `json:"object"`
		Data   []model `json:"data"`
	}{
		Object: "list",
		Data:   []model{},
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func writeOpenAIError(w http.ResponseWriter, statusCode int, message, errorType string) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(fmt.Sprintf(`{"error":{"message":%q,"type":%q}}`, message, errorType)))
}
