package startup

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/atlasbridge/atlasbridge/internal/config"
	"github.com/gofrs/flock"
)

var (
	fileLock    *flock.Flock
	lockPath    string
	PIDMetadata *PIDInfo
)

type PIDInfo struct {
	PID       int       `json:"pid"`
	StartedAt time.Time `json:"started_at"`
}

func Init() error {
	lockPath = filepath.Join(config.ConfigDir(), ".instance.lock")

	if err := os.MkdirAll(config.ConfigDir(), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	fileLock = flock.New(lockPath)

	locked, err := fileLock.TryLock()
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	if !locked {
		return fmt.Errorf("another instance is already running (lock held by another process)")
	}

	PIDMetadata = &PIDInfo{
		PID:       os.Getpid(),
		StartedAt: time.Now(),
	}

	return nil
}

func ReleaseLock() {
	if fileLock != nil {
		fileLock.Unlock()
		fileLock = nil
	}
	PIDMetadata = nil
}

func SetRunAtLogin(enable bool) error {
	return setRunAtLoginPlatform(enable)
}

func IsRegistered() bool {
	return isRegisteredPlatform()
}
