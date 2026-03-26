package access

import "testing"

func TestValidateBearerAPIKey_Success(t *testing.T) {
	t.Parallel()

	ok := ValidateBearerAPIKey("Bearer dev-key", []string{"dev-key", "other-key"})
	if !ok {
		t.Fatal("expected bearer token to be accepted")
	}
}

func TestValidateBearerAPIKey_SuccessWithLowercaseBearerScheme(t *testing.T) {
	t.Parallel()

	ok := ValidateBearerAPIKey("bearer dev-key", []string{"dev-key", "other-key"})
	if !ok {
		t.Fatal("expected case-insensitive bearer scheme to be accepted")
	}
}
