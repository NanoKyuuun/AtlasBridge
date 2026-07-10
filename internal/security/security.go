package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
)

func GenerateToken() (string, string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("failed to generate token: %w", err)
	}
	token := hex.EncodeToString(b)
	hash := HashToken(token)
	return token, hash, nil
}

func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func VerifyToken(token, hash string) bool {
	return HashToken(token) == hash
}

// EnsureToken checks whether a token hash exists. If not, it generates a
// new secure random token, writes its hash into *hashPtr, and returns the
// raw token. If a hash already exists, returns empty string (no-op).
func EnsureToken(hashPtr *string) (string, error) {
	if *hashPtr != "" {
		return "", nil
	}
	token, hash, err := GenerateToken()
	if err != nil {
		return "", err
	}
	*hashPtr = hash
	return token, nil
}

// AdminAuth returns HTTP middleware that guards routes using Bearer token
// authentication. It accepts a getter function that is called on every request,
// so the enabled flag and hash can be updated dynamically at runtime.
type AuthGetter func() (enabled bool, tokenHash string)

func AdminAuth(getter AuthGetter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			enabled, tokenHash := getter()
			if !enabled {
				next.ServeHTTP(w, r)
				return
			}

			token := r.Header.Get("Authorization")
			token = strings.TrimPrefix(token, "Bearer ")

			if token == "" {
				http.Error(w, `{"error":{"message":"unauthorized","type":"auth_error"}}`, http.StatusUnauthorized)
				return
			}

			if !VerifyToken(token, tokenHash) {
				http.Error(w, `{"error":{"message":"invalid token","type":"auth_error"}}`, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
