package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/auth"
	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/registry"
	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/translator"
)

type ChatRuntime interface {
	SupportsModel(model string) bool
	Execute(ctx context.Context, model string, openAIRequest []byte) (*auth.Result, error)
}

type OpenAI struct {
	modelRegistry *registry.ModelRegistry
	runtime       ChatRuntime
}

const maxChatCompletionsRequestBodyBytes int64 = 4 * 1024 * 1024

func NewOpenAI(modelRegistry *registry.ModelRegistry, runtime ChatRuntime) *OpenAI {
	return &OpenAI{
		modelRegistry: modelRegistry,
		runtime:       runtime,
	}
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

	if h.runtime != nil {
		if !h.runtime.SupportsModel(req.Model) {
			writeOpenAIError(w, http.StatusBadRequest, fmt.Sprintf("model %q is not available", req.Model), "invalid_request_error")
			return
		}

		result, err := h.runtime.Execute(r.Context(), req.Model, requestBody)
		if err != nil {
			var validationErr *translator.ValidationError
			if translator.AsValidationError(err, &validationErr) {
				writeOpenAIError(w, validationErr.StatusCode, validationErr.Message, "invalid_request_error")
				return
			}
			writeOpenAIError(w, http.StatusBadGateway, "upstream provider request failed", "api_error")
			return
		}
		if result == nil {
			writeOpenAIError(w, http.StatusBadGateway, "upstream provider request failed", "api_error")
			return
		}

		writeRuntimeResult(w, result)
		return
	}

	if h.modelRegistry == nil || len(h.modelRegistry.GetModelProviders(req.Model)) == 0 {
		writeOpenAIError(w, http.StatusBadRequest, fmt.Sprintf("model %q is not available", req.Model), "invalid_request_error")
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
	if h.modelRegistry != nil {
		models := h.modelRegistry.ListModels()
		response.Data = make([]model, 0, len(models))
		for _, info := range models {
			response.Data = append(response.Data, model{
				ID:     info.ID,
				Object: "model",
			})
		}
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func writeRuntimeResult(w http.ResponseWriter, result *auth.Result) {
	for key, values := range result.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}

	statusCode := result.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	w.WriteHeader(statusCode)
	if len(result.Body) == 0 {
		return
	}
	_, _ = w.Write(result.Body)
}

func writeOpenAIError(w http.ResponseWriter, statusCode int, message, errorType string) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(fmt.Sprintf(`{"error":{"message":%q,"type":%q}}`, message, errorType)))
}
