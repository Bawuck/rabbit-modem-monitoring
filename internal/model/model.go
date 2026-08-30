// Package model contains Gio-independent values shared by both windows.
package model

import "time"

type Scenario string

const (
	Loading      Scenario = "Loading"
	Online       Scenario = "Online"
	NoSignal     Scenario = "No Signal"
	APIError     Scenario = "API Error"
	Disconnected Scenario = "Disconnected"
)

func Scenarios() []Scenario {
	return []Scenario{Loading, Online, NoSignal, APIError, Disconnected}
}

func (s Scenario) Valid() bool {
	switch s {
	case Loading, Online, NoSignal, APIError, Disconnected:
		return true
	}
	return false
}

// Value distinguishes unavailable measurements from valid zero values.
type Value[T any] struct {
	Value T
	Valid bool
}

func Some[T any](v T) Value[T] { return Value[T]{Value: v, Valid: true} }

type Reading struct {
	Network  Value[string]
	Score    Value[int]
	Quality  string
	RSRP     Value[float64]
	RSRQ     Value[float64]
	SINR     Value[float64]
	RSSI     Value[float64]
	Band     Value[string]
	PCI      Value[int]
	Download Value[float64]
	Upload   Value[float64]
	Ping     Value[float64]
}

type Sample struct {
	At               time.Time
	RSRP, RSRQ, SINR Value[float64]
}

// Snapshot is a detached view. Mutating its History cannot affect the store.
type Snapshot struct {
	State     Scenario
	Reading   Reading
	HasData   bool
	Stale     bool
	UpdatedAt time.Time
	History   []Sample
}
