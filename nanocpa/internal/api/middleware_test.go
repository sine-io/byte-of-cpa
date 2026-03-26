package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIKeyMiddleware_AllowsAuthorizedRequest(t *testing.T) {
	t.Parallel()

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusNoContent)
	})

	handler := APIKeyMiddleware([]string{"dev-key"}, next)
	req := httptest.NewRequest(http.MethodGet, "/v1/chat/completions", nil)
	req.Header.Set("Authorization", "bearer dev-key")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if !nextCalled {
		t.Fatal("expected middleware to call next handler")
	}
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
}

func TestAPIKeyMiddleware_UnauthorizedRequests(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		authorization string
	}{
		{
			name:          "wrong scheme",
			authorization: "Basic dev-key",
		},
		{
			name:          "missing token",
			authorization: "Bearer",
		},
		{
			name:          "wrong token",
			authorization: "Bearer wrong-key",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusNoContent)
			})

			handler := APIKeyMiddleware([]string{"dev-key"}, next)
			req := httptest.NewRequest(http.MethodGet, "/v1/chat/completions", nil)
			if tc.authorization != "" {
				req.Header.Set("Authorization", tc.authorization)
			}
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, req)

			if nextCalled {
				t.Fatal("expected middleware to block unauthorized request")
			}
			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
			}
			if got := recorder.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("expected content-type application/json, got %q", got)
			}
			if got := recorder.Header().Get("WWW-Authenticate"); got != "Bearer" {
				t.Fatalf("expected WWW-Authenticate Bearer, got %q", got)
			}
			const wantBody = `{"error":{"message":"unauthorized","type":"invalid_request_error"}}`
			if got := recorder.Body.String(); got != wantBody {
				t.Fatalf("expected body %q, got %q", wantBody, got)
			}
		})
	}
}

func TestAPIKeyMiddleware_MissingAuthorizationHeaderUnauthorized(t *testing.T) {
	t.Parallel()

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusNoContent)
	})

	handler := APIKeyMiddleware([]string{"dev-key"}, next)
	req := httptest.NewRequest(http.MethodGet, "/v1/chat/completions", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if nextCalled {
		t.Fatal("expected middleware to block request without authorization header")
	}
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected content-type application/json, got %q", got)
	}
	if got := recorder.Header().Get("WWW-Authenticate"); got != "Bearer" {
		t.Fatalf("expected WWW-Authenticate Bearer, got %q", got)
	}
	const wantBody = `{"error":{"message":"unauthorized","type":"invalid_request_error"}}`
	if got := recorder.Body.String(); got != wantBody {
		t.Fatalf("expected body %q, got %q", wantBody, got)
	}
}
