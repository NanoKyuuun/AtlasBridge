package runtime

import (
	"fmt"
	"sync"
	"time"
)

type Mode string

const (
	ModeAlwaysOn Mode = "always_on"
	ModeManual   Mode = "manual"
	ModeDisabled Mode = "disabled"
)

type Status string

const (
	StatusStarting      Status = "starting"
	StatusRunning       Status = "running"
	StatusStopped       Status = "stopped"
	StatusDisabled      Status = "disabled"
	StatusError         Status = "error"
	StatusPortConflict  Status = "port_conflict"
	StatusDownstreamOff Status = "downstream_offline"
)

type State struct {
	mu         sync.RWMutex
	mode       Mode
	status     Status
	errorMsg   string
	startedAt  time.Time
	lastAction string
	lastError  error
}

func NewState(initialMode Mode) *State {
	status := StatusStopped
	if initialMode == ModeDisabled {
		status = StatusDisabled
	}
	return &State{
		mode:   initialMode,
		status: status,
	}
}

func (s *State) GetMode() Mode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mode
}

func (s *State) GetStatus() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

func (s *State) GetErrorMsg() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.errorMsg
}

func (s *State) GetUptime() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.startedAt.IsZero() || s.status != StatusRunning {
		return 0
	}
	return time.Since(s.startedAt)
}

func (s *State) GetStateSnapshot() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snap := map[string]interface{}{
		"mode":   string(s.mode),
		"status": string(s.status),
	}
	if s.errorMsg != "" {
		snap["error"] = s.errorMsg
	}
	if s.status == StatusRunning && !s.startedAt.IsZero() {
		snap["uptime"] = time.Since(s.startedAt).String()
	}
	if !s.startedAt.IsZero() {
		snap["started_at"] = s.startedAt.Format(time.RFC3339)
	}
	return snap
}

func (s *State) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.status == StatusRunning {
		return fmt.Errorf("proxy already running")
	}

	s.status = StatusRunning
	s.mode = ModeAlwaysOn
	s.errorMsg = ""
	s.startedAt = time.Now()
	s.lastAction = "start"
	return nil
}

func (s *State) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.status == StatusStopped || s.status == StatusDisabled {
		return fmt.Errorf("proxy not running")
	}

	s.status = StatusStopped
	s.mode = ModeManual
	s.errorMsg = ""
	s.lastAction = "stop"
	return nil
}

func (s *State) Restart() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	wasRunning := s.status == StatusRunning

	s.status = StatusStopped
	s.errorMsg = ""
	s.lastAction = "restart"

	if wasRunning {
		s.status = StatusRunning
		s.startedAt = time.Now()
	}

	return nil
}

func (s *State) SetMode(mode Mode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mode = mode
}

func (s *State) SetError(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errorMsg = msg
	s.status = StatusError
	s.lastAction = "error"
}

func (s *State) SetStatus(status Status) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = status
}
