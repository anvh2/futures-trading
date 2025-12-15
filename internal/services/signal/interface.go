package signal

import (
	"github.com/anvh2/futures-trading/internal/models"
)

type Signal models.Signal

func New(signal *models.Signal) *Signal {
	sig := Signal(*signal)
	return &sig
}

func Extract(signal *Signal) *models.Signal {
	return (*models.Signal)(signal)
}

func (s Signal) ID() string {
	return s.Symbol
}

func (s Signal) Score() float64 {
	return s.Indicators["rsi"]
}

// Service defines the interface for the signal service that manages
// multiple interval-isolated priority queues of signals.
type Service interface {
	// Ingest adds a signal to the appropriate interval queue
	Ingest(signal *models.Signal)

	// Peek returns the highest priority signal for an interval without removing it
	Peek(interval string) *models.Signal

	// Pop removes and returns the highest priority signal for an interval
	Pop(interval string) *models.Signal

	// Intervals returns all available intervals that have signals
	Intervals() []string

	// Stats returns statistics about the signal service including the number
	// of signals per interval. This is useful for monitoring and debugging.
	Stats(intervals ...string) map[string]int
}
