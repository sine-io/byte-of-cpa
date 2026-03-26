package api

import (
	"fmt"
	"net/http"

	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/access"
)

func APIKeyMiddleware(allowedKeys []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !access.ValidateBearerAPIKey(r.Header.Get("authorization"), allowedKeys) {
			writeAPIError(w, http.StatusUnauthorized, "unauthorized", "invalid_request_error")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeAPIError(w http.ResponseWriter, statusCode int, message, errorType string) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(fmt.Sprintf(`{"error":{"message":%q,"type":%q}}`, message, errorType)))
}
