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
	execpkg "github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/runtime/executor"
	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/translator"
)

type ChatCompletionsRunner interface {
	Execute(ctx context.Context, model string, openAIRequest []byte) (*auth.Result, error)
}

type ModelSupportChecker interface {
	SupportsModel(model string) bool
}

type OpenAI struct {
	runner        ChatCompletionsRunner
	modelRegistry *registry.ModelRegistry
	supportsModel ModelSupportChecker
}

const maxChatCompletionsRequestBodyBytes int64 = 4 * 1024 * 1024

func NewOpenAI(runner ChatCompletionsRunner, modelRegistry *registry.ModelRegistry) *OpenAI {
	openAI := &OpenAI{
		runner:        runner,
		modelRegistry: modelRegistry,
	}
	if checker, ok := runner.(ModelSupportChecker); ok {
		openAI.supportsModel = checker
	}
	return openAI
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
	if h.runner == nil {
		writeOpenAIError(w, http.StatusBadGateway, "upstream provider request failed", "api_error")
		return
	}
	if !h.isAllowedModel(req.Model) {
		writeOpenAIError(w, http.StatusBadRequest, fmt.Sprintf("model %q is not allowed", req.Model), "invalid_request_error")
		return
	}

	result, err := h.runner.Execute(r.Context(), req.Model, requestBody)
	if err != nil {
		var validationErr *translator.ValidationError
		if errors.As(err, &validationErr) {
			writeOpenAIError(w, validationErr.StatusCode, validationErr.Message, "invalid_request_error")
			return
		}
		var upstreamErr *execpkg.UpstreamError
		if errors.As(err, &upstreamErr) {
			statusCode := upstreamErr.StatusCode
			if statusCode < 400 || statusCode > 599 {
				statusCode = http.StatusBadGateway
			}
			writeOpenAIError(w, statusCode, fmt.Sprintf("upstream provider error (%d)", upstreamErr.StatusCode), "api_error")
			return
		}
		writeOpenAIError(w, http.StatusBadGateway, "upstream provider request failed", "api_error")
		return
	}

	for key, values := range result.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	if w.Header().Get("content-type") == "" {
		w.Header().Set("content-type", "application/json")
	}
	w.WriteHeader(result.StatusCode)
	_, _ = w.Write(result.Body)
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
		for _, m := range models {
			if h.supportsModel != nil && !h.supportsModel.SupportsModel(m.ID) {
				continue
			}
			response.Data = append(response.Data, model{ID: m.ID, Object: "model"})
		}
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func writeOpenAIError(w http.ResponseWriter, statusCode int, message, errorType string) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(fmt.Sprintf(`{"error":{"message":%q,"type":%q}}`, message, errorType)))
}

func (h *OpenAI) isAllowedModel(model string) bool {
	if h.supportsModel != nil {
		return h.supportsModel.SupportsModel(model)
	}
	if h.modelRegistry == nil {
		return false
	}
	return len(h.modelRegistry.GetModelProviders(model)) > 0
}
