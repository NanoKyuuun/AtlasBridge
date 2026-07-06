package runtime

import (
	"testing"
	"time"
)

func TestNewState(t *testing.T) {
	s := NewState(ModeManual)
	if s.GetMode() != ModeManual {
		t.Errorf("expected mode manual, got %s", s.GetMode())
	}
	if s.GetStatus() != StatusStopped {
		t.Errorf("expected status stopped, got %s", s.GetStatus())
	}
}

func TestNewStateDisabled(t *testing.T) {
	s := NewState(ModeDisabled)
	if s.GetStatus() != StatusDisabled {
		t.Errorf("expected status disabled, got %s", s.GetStatus())
	}
}

func TestStartStop(t *testing.T) {
	s := NewState(ModeManual)

	if err := s.Start(); err != nil {
		t.Fatalf("unexpected error starting: %v", err)
	}
	if s.GetStatus() != StatusRunning {
		t.Errorf("expected status running, got %s", s.GetStatus())
	}
	if s.GetMode() != ModeAlwaysOn {
		t.Errorf("expected mode always_on, got %s", s.GetMode())
	}

	if err := s.Stop(); err != nil {
		t.Fatalf("unexpected error stopping: %v", err)
	}
	if s.GetStatus() != StatusStopped {
		t.Errorf("expected status stopped, got %s", s.GetStatus())
	}
}

func TestStartAlreadyRunning(t *testing.T) {
	s := NewState(ModeManual)
	s.Start()

	if err := s.Start(); err == nil {
		t.Error("expected error starting already running proxy")
	}
}

func TestStopNotRunning(t *testing.T) {
	s := NewState(ModeManual)

	if err := s.Stop(); err == nil {
		t.Error("expected error stopping non-running proxy")
	}
}

func TestRestart(t *testing.T) {
	s := NewState(ModeManual)
	s.Start()

	if err := s.Restart(); err != nil {
		t.Fatalf("unexpected error restarting: %v", err)
	}
	if s.GetStatus() != StatusRunning {
		t.Errorf("expected status running after restart, got %s", s.GetStatus())
	}
}

func TestSetMode(t *testing.T) {
	s := NewState(ModeManual)
	s.SetMode(ModeDisabled)
	if s.GetMode() != ModeDisabled {
		t.Errorf("expected mode disabled, got %s", s.GetMode())
	}
}

func TestSetError(t *testing.T) {
	s := NewState(ModeManual)
	s.Start()
	s.SetError("something broke")

	if s.GetStatus() != StatusError {
		t.Errorf("expected status error, got %s", s.GetStatus())
	}
	if s.GetErrorMsg() != "something broke" {
		t.Errorf("expected error message 'something broke', got %s", s.GetErrorMsg())
	}
}

func TestGetStateSnapshot(t *testing.T) {
	s := NewState(ModeManual)
	s.Start()

	snap := s.GetStateSnapshot()
	if snap["mode"] != "always_on" {
		t.Errorf("expected mode always_on, got %v", snap["mode"])
	}
	if snap["status"] != "running" {
		t.Errorf("expected status running, got %v", snap["status"])
	}
	if snap["uptime"] == nil {
		t.Error("expected uptime in snapshot")
	}
}

func TestGetStateSnapshotError(t *testing.T) {
	s := NewState(ModeManual)
	s.SetError("test error")

	snap := s.GetStateSnapshot()
	if snap["error"] != "test error" {
		t.Errorf("expected error in snapshot, got %v", snap["error"])
	}
}

func TestGetUptime(t *testing.T) {
	s := NewState(ModeManual)

	if s.GetUptime() != 0 {
		t.Errorf("expected 0 uptime when not running, got %v", s.GetUptime())
	}

	s.Start()
	// Small sleep to ensure measurable uptime
	time.Sleep(10 * time.Millisecond)
	uptime := s.GetUptime()
	if uptime <= 0 {
		t.Errorf("expected positive uptime, got %v", uptime)
	}
}
