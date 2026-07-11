package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var writeMu sync.Mutex

func atomicWriteFile(targetPath string, data []byte, perm os.FileMode) error {
	writeMu.Lock()
	defer writeMu.Unlock()

	dir := filepath.Dir(targetPath)

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	tmpFile, err := os.CreateTemp(dir, ".tmp-config-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	success := false
	defer func() {
		if !success {
			tmpFile.Close()
			os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("sync temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Chmod(tmpPath, perm); err != nil {
		return fmt.Errorf("chmod temp file: %w", err)
	}

	if err := os.Rename(tmpPath, targetPath); err != nil {
		return fmt.Errorf("atomic rename: %w", err)
	}

	success = true
	return nil
}

func backupFilePath(targetPath string) string {
	return targetPath + ".bak"
}

func saveWithBackup(targetPath string, data []byte, perm os.FileMode) error {
	if _, err := os.Stat(targetPath); err == nil {
		backupPath := backupFilePath(targetPath)
		_ = copyFile(targetPath, backupPath)
	}

	return atomicWriteFile(targetPath, data, perm)
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o600)
}
