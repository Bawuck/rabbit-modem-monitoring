// Package monitor owns the live snapshot and bounded, in-memory signal history.
package monitor

import (
	"sync"
	"time"

	"example.com/4g-monitor/internal/model"
)

const HistoryLimit = 30

type Store struct {
	mu      sync.RWMutex
	state   model.ConnectionState
	message string
	last    model.Reading
	hasData bool
	updated time.Time
	history []model.Sample
	changed chan struct{}
}

func NewStore() *Store {
	return &Store{state: model.Loading, changed: make(chan struct{}, 1)}
}

func (s *Store) Changed() <-chan struct{} { return s.changed }

// Apply publishes a whole cycle; failed cycles cannot replace any cached field.
func (s *Store) Apply(update model.Update) {
	s.mu.Lock()
	if update.State == model.Online {
		if s.state != model.Online {
			s.history = nil // Never connect chart lines across a monitoring outage.
		}
		s.last = update.Reading
		s.hasData = true
		s.updated = update.ReceivedAt
		s.history = append(s.history, model.Sample{
			At: update.ReceivedAt, RSRP: update.Reading.RSRP,
			RSRQ: update.Reading.RSRQ, SINR: update.Reading.SINR,
		})
		if len(s.history) > HistoryLimit {
			copy(s.history, s.history[len(s.history)-HistoryLimit:])
			s.history = s.history[:HistoryLimit]
		}
	}
	s.state = update.State
	s.message = update.Message
	s.mu.Unlock()
	select {
	case s.changed <- struct{}{}:
	default:
	}
}

func (s *Store) Snapshot() model.Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	view := model.Snapshot{State: s.state, Message: s.message}
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
