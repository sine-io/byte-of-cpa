package access

import "testing"

func TestValidateBearerAPIKey_ValidBearerTokenAccepted(t *testing.T) {
	t.Parallel()

	ok := ValidateBearerAPIKey("Bearer dev-key", []string{"dev-key", "other-key"})
	if !ok {
		t.Fatal("expected bearer token to be accepted")
	}
}

func TestValidateBearerAPIKey_CaseInsensitiveSchemeAccepted(t *testing.T) {
	t.Parallel()

	ok := ValidateBearerAPIKey("bearer dev-key", []string{"dev-key", "other-key"})
	if !ok {
		t.Fatal("expected lowercase bearer scheme to be accepted")
	}
}

func TestValidateBearerAPIKey_WrongSchemeRejected(t *testing.T) {
	t.Parallel()

	ok := ValidateBearerAPIKey("Basic dev-key", []string{"dev-key", "other-key"})
	if ok {
		t.Fatal("expected wrong scheme to be rejected")
	}
}

func TestValidateBearerAPIKey_MissingTokenRejected(t *testing.T) {
	t.Parallel()

	ok := ValidateBearerAPIKey("Bearer", []string{"dev-key", "other-key"})
	if ok {
		t.Fatal("expected missing token to be rejected")
	}
}

func TestValidateBearerAPIKey_WrongTokenRejected(t *testing.T) {
	t.Parallel()

	ok := ValidateBearerAPIKey("Bearer wrong-key", []string{"dev-key", "other-key"})
	if ok {
		t.Fatal("expected wrong token to be rejected")
	}
}
