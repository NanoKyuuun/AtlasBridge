package startup

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/atlasbridge/atlasbridge/internal/config"
	"golang.org/x/sys/windows/registry"
)

var (
	instanceLock *sync.Mutex
	once         sync.Once
	lockFile     string
)

const appName = "AtlasBridge"

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
	if runtime.GOOS != "windows" {
		return nil
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	return setWindowsRunKey(enable, exePath)
}

func IsRegistered() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	return checkWindowsRunKey()
}

func setWindowsRunKey(enable bool, exePath string) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	if enable {
		return key.SetStringValue(appName, exePath)
	}

	err = key.DeleteValue(appName)
	if err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("failed to delete registry value: %w", err)
	}
	return nil
}

func checkWindowsRunKey() bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.READ)
	if err != nil {
		return false
	}
	defer key.Close()

	_, _, err = key.GetStringValue(appName)
	return err == nil
}