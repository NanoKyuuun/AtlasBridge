package security

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
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
	return subtle.ConstantTimeCompare([]byte(HashToken(token)), []byte(hash)) == 1
}

// HashPassword hashes a plain-text password using bcrypt with default cost.
func HashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		// bcrypt failure is extremely rare; fall back to SHA-256 to avoid data loss
		h := sha256.Sum256([]byte(password))
		return "sha256:" + hex.EncodeToString(h[:])
	}
	return string(hash)
}

// VerifyPassword checks a plain-text password against a stored hash.
// Supports both bcrypt hashes and legacy SHA-256 hashes (prefixed "sha256:").
func VerifyPassword(password, hash string) bool {
	if strings.HasPrefix(hash, "sha256:") {
		h := sha256.Sum256([]byte(password))
		return subtle.ConstantTimeCompare([]byte(hash), []byte("sha256:"+hex.EncodeToString(h[:]))) == 1
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
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
// If expiresAt is non-zero and in the past, the session is considered expired.
type AuthGetter func() (enabled bool, tokenHash string, expiresAt int64)

func AdminAuth(getter AuthGetter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			enabled, tokenHash, expiresAt := getter()
			if !enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Check session expiry
			if expiresAt > 0 && time.Now().Unix() > expiresAt {
				http.Error(w, `{"error":{"message":"session expired, please login again","type":"auth_error"}}`, http.StatusUnauthorized)
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
