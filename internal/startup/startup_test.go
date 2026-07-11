package startup

import (
	"testing"
)

func TestInitAndRelease(t *testing.T) {
	if err := Init(); err != nil {
		t.Fatalf("expected Init to succeed, got: %v", err)
	}
	if fileLock == nil {
		t.Fatal("expected fileLock to be set")
	}
	if PIDMetadata == nil {
		t.Fatal("expected PIDMetadata to be set")
	}
	if PIDMetadata.PID <= 0 {
		t.Errorf("expected positive PID, got %d", PIDMetadata.PID)
	}
	ReleaseLock()
	if fileLock != nil {
		t.Error("expected fileLock to be nil after release")
	}
}

func TestSecondInstanceFails(t *testing.T) {
	if err := Init(); err != nil {
		t.Fatalf("first instance should succeed, got: %v", err)
	}
	defer ReleaseLock()

	err := Init()
	if err == nil {
		ReleaseLock()
		t.Fatal("expected second Init to fail, but it succeeded")
	}
	t.Logf("second instance correctly rejected: %v", err)
}
