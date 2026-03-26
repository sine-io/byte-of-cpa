package access

import "strings"

func ValidateBearerAPIKey(authorizationHeader string, allowedKeys []string) bool {
	parts := strings.Fields(authorizationHeader)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return false
	}

	key := strings.TrimSpace(parts[1])
	if key == "" {
		return false
	}

	for _, allowed := range allowedKeys {
		if key == allowed {
			return true
		}
	}
	return false
}
