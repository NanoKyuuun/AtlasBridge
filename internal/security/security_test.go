package security

import (
	"testing"
)

func TestGenerateToken(t *testing.T) {
	token, hash, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
	if !VerifyToken(token, hash) {
		t.Fatal("token should verify against its own hash")
	}
}

func TestGenerateTokenUnique(t *testing.T) {
	t1, h1, _ := GenerateToken()
	t2, h2, _ := GenerateToken()
	if t1 == t2 {
		t.Error("expected different tokens")
	}
	if h1 == h2 {
		t.Error("expected different hashes")
	}
}

func TestHashTokenConsistent(t *testing.T) {
	token := "test-token-123"
	h1 := HashToken(token)
	h2 := HashToken(token)
	if h1 != h2 {
		t.Error("HashToken should be deterministic")
	}
	if h1 == "" {
		t.Error("expected non-empty hash")
	}
}

func TestVerifyTokenValid(t *testing.T) {
	token, hash, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if !VerifyToken(token, hash) {
		t.Error("VerifyToken should return true for valid token")
	}
}

func TestVerifyTokenInvalid(t *testing.T) {
	_, hash, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if VerifyToken("wrong-token", hash) {
		t.Error("VerifyToken should return false for invalid token")
	}
}

func TestEnsureTokenGeneration(t *testing.T) {
	var hash string
	raw, err := EnsureToken(&hash)
	if err != nil {
		t.Fatalf("EnsureToken failed: %v", err)
	}
	if raw == "" {
		t.Fatal("expected raw token when hash is empty")
	}
	if hash == "" {
		t.Fatal("expected hash to be populated")
	}
	if !VerifyToken(raw, hash) {
		t.Fatal("generated token should verify against populated hash")
	}
}

func TestEnsureTokenNoRegeneration(t *testing.T) {
	// Pre-populate a hash
	existingHash := HashToken("my-token")
	hash := existingHash

	raw, err := EnsureToken(&hash)
	if err != nil {
		t.Fatalf("EnsureToken failed: %v", err)
	}
	if raw != "" {
		t.Error("expected empty raw token when hash already exists")
	}
	if hash != existingHash {
		t.Error("hash should not be changed when already set")
	}
}

func TestEnsureTokenMultipleCalls(t *testing.T) {
	var hash string

	raw1, err := EnsureToken(&hash)
	if err != nil {
		t.Fatalf("first EnsureToken failed: %v", err)
	}
	if raw1 == "" {
		t.Fatal("first call should generate token")
	}

	savedHash := hash

	raw2, err := EnsureToken(&hash)
	if err != nil {
		t.Fatalf("second EnsureToken failed: %v", err)
	}
	if raw2 != "" {
		t.Error("second call should not regenerate")
	}
	if hash != savedHash {
		t.Error("hash should remain the same")
	}
}
