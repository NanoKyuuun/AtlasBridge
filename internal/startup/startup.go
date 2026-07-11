package startup

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/atlasbridge/atlasbridge/internal/config"
	"github.com/gofrs/flock"
)

var (
	fileLock    *flock.Flock
	lockPath    string
	pidPath     string
	PIDMetadata *PIDInfo
)

type PIDInfo struct {
	PID       int       `json:"pid"`
	StartedAt time.Time `json:"started_at"`
}

func Init() error {
	lockPath = filepath.Join(config.ConfigDir(), ".instance.lock")
	pidPath = filepath.Join(config.ConfigDir(), ".instance.pid")

	if err := os.MkdirAll(config.ConfigDir(), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	fileLock = flock.New(lockPath)

	locked, err := fileLock.TryLock()
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	if !locked {
		if stalePID, readErr := readPIDFile(); readErr == nil && !isProcessAlive(stalePID) {
			os.Remove(lockPath)
			os.Remove(pidPath)
			locked, err = fileLock.TryLock()
			if err != nil || !locked {
				return fmt.Errorf("failed to recover stale lock: %w", err)
			}
		} else {
			return fmt.Errorf("another instance is already running (lock held by process %d)", stalePID)
		}
	}

	PIDMetadata = &PIDInfo{
		PID:       os.Getpid(),
		StartedAt: time.Now(),
	}

	writePIDFile(os.Getpid())

	return nil
}

func ReleaseLock() {
	if fileLock != nil {
		fileLock.Unlock()
		fileLock = nil
	}
	PIDMetadata = nil
	os.Remove(pidPath)
}

func writePIDFile(pid int) {
	os.WriteFile(pidPath, []byte(strconv.Itoa(pid)), 0o600)
}

func readPIDFile() (int, error) {
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, err
	}
	return pid, nil
}

func SetRunAtLogin(enable bool) error {
	return setRunAtLoginPlatform(enable)
}

func IsRegistered() bool {
	return isRegisteredPlatform()
}
