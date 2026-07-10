package startup

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/atlasbridge/atlasbridge/internal/config"
)

var (
	instanceLock *sync.Mutex
	once         sync.Once
	lockFile     string
)

func Init() error {
	once.Do(func() {
		lockFile = filepath.Join(config.ConfigDir(), ".instance.lock")
	})
	return acquireLock()
}

func acquireLock() error {
	instanceLock = &sync.Mutex{}
	instanceLock.Lock()

	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		f, err := os.Create(lockFile)
		if err != nil {
			return fmt.Errorf("failed to create lock file: %w", err)
		}
		f.Close()
		return nil
	}

	instanceLock.Unlock()
	return fmt.Errorf("another instance is already running")
}

func ReleaseLock() {
	if instanceLock != nil {
		os.Remove(lockFile)
		instanceLock.Unlock()
	}
}

func SetRunAtLogin(enable bool) error {
	return setRunAtLoginPlatform(enable)
}

func IsRegistered() bool {
	return isRegisteredPlatform()
}