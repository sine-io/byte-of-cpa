package antigravity

import "testing"

func TestOAuthClientCredentialsFromEnvironment(t *testing.T) {
	t.Setenv("ANTIGRAVITY_OAUTH_CLIENT_ID", "  test-client-id  ")
	t.Setenv("ANTIGRAVITY_OAUTH_CLIENT_SECRET", "  test-client-secret  ")

	if got := OAuthClientID(); got != "test-client-id" {
		t.Fatalf("OAuthClientID() = %q, want %q", got, "test-client-id")
	}
	if got := OAuthClientSecret(); got != "test-client-secret" {
		t.Fatalf("OAuthClientSecret() = %q, want %q", got, "test-client-secret")
	}
}

func TestOAuthClientCredentialsUnset(t *testing.T) {
	t.Setenv("ANTIGRAVITY_OAUTH_CLIENT_ID", "")
	t.Setenv("ANTIGRAVITY_OAUTH_CLIENT_SECRET", "")

	if got := OAuthClientID(); got != "" {
		t.Fatalf("OAuthClientID() = %q, want empty string", got)
	}
	if got := OAuthClientSecret(); got != "" {
		t.Fatalf("OAuthClientSecret() = %q, want empty string", got)
	}
}
