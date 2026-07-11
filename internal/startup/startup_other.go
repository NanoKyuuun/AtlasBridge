//go:build !windows

package startup

import "syscall"

func setRunAtLoginPlatform(enable bool) error {
	return nil
}

func isRegisteredPlatform() bool {
	return false
}

func isProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil
}
