// Package mock supplies a deterministic, in-memory demo; it never accesses a device.
package mock

import (
	"sync"
	"time"

	"example.com/4g-monitor/internal/model"
)

const HistoryLimit = 30

type Store struct {
	mu      sync.RWMutex
	state   model.Scenario
	startup bool
	next    int
	last    model.Reading
	hasData bool
	updated time.Time
	history []model.Sample
	changed chan struct{}
}

func NewStore() *Store {
	return &Store{state: model.Loading, startup: true, changed: make(chan struct{}, 1)}
}

// Changed coalesces repaint notifications without blocking the UI or the store.
func (s *Store) Changed() <-chan struct{} { return s.changed }

func (s *Store) notify() {
	select {
	case s.changed <- struct{}{}:
	default:
	}
}

// Tick is called by the single application ticker, including while stale so
// the UI can show increasing data age without changing the sample timestamp.
func (s *Store) Tick(now time.Time) {
	s.mu.Lock()
	if s.startup {
		s.startup = false
		s.state = model.Online
	}
	if s.state == model.Online {
		s.advance(now)
	}
	s.mu.Unlock()
}

// Select cancels startup auto-transition even when Loading is already selected.
func (s *Store) Select(state model.Scenario) {
	if !state.Valid() {
		return
	}
	s.mu.Lock()
	s.startup = false
	if state != s.state {
		s.state = state
		if state == model.Online {
			s.history = nil
			s.advance(time.Now())
		}
	}
	s.mu.Unlock()
	s.notify()
}

func (s *Store) advance(now time.Time) {
	s.last = fixtures[s.next]
	s.next = (s.next + 1) % len(fixtures)
	s.hasData = true
	s.updated = now
	s.history = append(s.history, model.Sample{
		At: now, RSRP: s.last.RSRP, RSRQ: s.last.RSRQ, SINR: s.last.SINR,
	})
	if len(s.history) > HistoryLimit {
		copy(s.history, s.history[len(s.history)-HistoryLimit:])
		s.history = s.history[:HistoryLimit]
	}
}

func (s *Store) Snapshot() model.Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	view := model.Snapshot{State: s.state}
	if s.state == model.Loading || s.state == model.NoSignal || !s.hasData {
		return view
	}
	view.Reading = s.last
	view.HasData = true
	view.Stale = s.state == model.APIError || s.state == model.Disconnected
	view.UpdatedAt = s.updated
	view.History = append([]model.Sample(nil), s.history...)
	return view
}
