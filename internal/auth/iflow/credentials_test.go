package iflow

import "testing"

func TestOAuthClientCredentialsFromEnvironment(t *testing.T) {
	t.Setenv("IFLOW_OAUTH_CLIENT_ID", "  test-client-id  ")
	t.Setenv("IFLOW_OAUTH_CLIENT_SECRET", "  test-client-secret  ")

	if got := oauthClientID(); got != "test-client-id" {
		t.Fatalf("oauthClientID() = %q, want %q", got, "test-client-id")
	}
	if got := oauthClientSecret(); got != "test-client-secret" {
		t.Fatalf("oauthClientSecret() = %q, want %q", got, "test-client-secret")
	}
}

func TestOAuthClientCredentialsUnset(t *testing.T) {
	t.Setenv("IFLOW_OAUTH_CLIENT_ID", "")
	t.Setenv("IFLOW_OAUTH_CLIENT_SECRET", "")

	if got := oauthClientID(); got != "" {
		t.Fatalf("oauthClientID() = %q, want empty string", got)
	}
	if got := oauthClientSecret(); got != "" {
		t.Fatalf("oauthClientSecret() = %q, want empty string", got)
	}
}
