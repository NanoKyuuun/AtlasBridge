//go:build !windows

package startup

func setRunAtLoginPlatform(enable bool) error {
	return nil
}

func isRegisteredPlatform() bool {
	return false
}
