// Package model contains Gio-independent values shared by both windows.
package model

import "time"

type ConnectionState string

const (
	Loading      ConnectionState = "Loading"
	Online       ConnectionState = "Online"
	NoSignal     ConnectionState = "No Signal"
	APIError     ConnectionState = "API Error"
	Disconnected ConnectionState = "Disconnected"
)

// Value distinguishes unavailable measurements from valid zero values.
type Value[T any] struct {
	Value T
	Valid bool
}

func Some[T any](v T) Value[T] { return Value[T]{Value: v, Valid: true} }

type Reading struct {
	Network    Value[string]
	SignalBars Value[int]
	RSRP       Value[float64]
	RSRQ       Value[float64]
	SINR       Value[float64]
	RSSI       Value[float64]
	Band       Value[string]
	PCI        Value[int]
	Download   Value[float64]
	Upload     Value[float64]
	Ping       Value[float64]
}

type Sample struct {
	At               time.Time
	RSRP, RSRQ, SINR Value[float64]
}

// Update is one complete polling cycle. Only Online carries fresh measurements.
// Message contains a safe explanation, never an HTTP body or device identifier.
type Update struct {
	State      ConnectionState
	Reading    Reading
	ReceivedAt time.Time
	Message    string
}

// Snapshot is a detached view. Mutating its History cannot affect the store.
type Snapshot struct {
	State     ConnectionState
	Message   string
	Reading   Reading
	HasData   bool
	Stale     bool
	UpdatedAt time.Time
	History   []Sample
}
