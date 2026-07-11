package server

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/atlasbridge/atlasbridge/internal/config"
	"github.com/atlasbridge/atlasbridge/internal/forwarder"
)

// Snapshot is an immutable snapshot of all runtime state.
// Once created, it must never be modified. All updates create a new Snapshot
// and atomically swap it in via StateStore.
type Snapshot struct {
	Config    config.Config
	Routes    config.RoutesConfig
	Profiles  config.ProfilesConfig
	Forwarder *forwarder.Forwarder
	Version   uint64
	CreatedAt time.Time
}

// Clone returns a deep copy of the snapshot. Maps in Routes and Profiles are
// copied so the new snapshot is fully independent.
func (s *Snapshot) Clone() *Snapshot {
	cp := *s
	cp.Routes = copyRoutes(s.Routes)
	cp.Profiles = copyProfiles(s.Profiles)
	return &cp
}

// StateStore provides lock-free read access to the current runtime snapshot
// via atomic.Pointer, and serializes disk-persisting mutations via persistMu.
type StateStore struct {
	current   atomic.Pointer[Snapshot]
	persistMu sync.Mutex
}

// NewStateStore creates a StateStore with the given initial snapshot.
func NewStateStore(snap *Snapshot) *StateStore {
	s := &StateStore{}
	s.current.Store(snap)
	return s
}

// Load returns the current snapshot. This is safe for concurrent use and
// never blocks. Every request should call Load() once at the start and
// use that single snapshot for the entire request lifecycle.
func (s *StateStore) Load() *Snapshot {
	return s.current.Load()
}

// Swap atomically replaces the current snapshot. Callers must have built a
// completely new Snapshot (typically via Clone + modifications).
func (s *StateStore) Swap(next *Snapshot) {
	s.current.Store(next)
}

func copyRoutes(r config.RoutesConfig) config.RoutesConfig {
	cp := r
	cp.TaskRoutes = make(map[string]string, len(r.TaskRoutes))
	for k, v := range r.TaskRoutes {
		cp.TaskRoutes[k] = v
	}
	return cp
}

func copyProfiles(p config.ProfilesConfig) config.ProfilesConfig {
	cp := p
	cp.RouteProfiles = make(map[string]config.RouteProfile, len(p.RouteProfiles))
	for k, v := range p.RouteProfiles {
		cp.RouteProfiles[k] = v
	}
	return cp
}
