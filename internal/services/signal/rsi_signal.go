package signal

import (
	"fmt"
	"time"

	"github.com/anvh2/futures-trading/internal/models"
)

// RSISignal is a concrete implementation of the Signal interface
// that uses RSI (Relative Strength Index) for signal scoring and prioritization.
type RSISignal struct {
	oscillator *models.Oscillator
	interval   string
	score      float64
	timestamp  time.Time
}

// NewRSISignal creates a new RSI-based signal from an oscillator.
// The RSI value for the specified interval is used as the priority score.
func NewRSISignal(o *models.Oscillator, interval string) *RSISignal {
	return &RSISignal{
		oscillator: o,
		interval:   interval,
		score:      o.GetRSI(interval),
		timestamp:  time.Now(),
	}
}

// ID returns a unique identifier for the signal.
func (s *RSISignal) ID() string {
	return fmt.Sprintf("%s-%s", s.oscillator.Symbol, s.interval)
}

// Symbol returns the trading symbol (e.g., "BTCUSDT")
func (s *RSISignal) Symbol() string {
	return s.oscillator.Symbol
}

// Interval returns the timeframe interval (e.g., "1m", "5m", "1h")
func (s *RSISignal) Interval() string {
	return s.interval
}

// Score returns the RSI value as the priority score
// Higher RSI values indicate stronger overbought conditions and higher priority
func (s *RSISignal) Score() float64 {
	return s.score
}

// Timestamp returns when the signal was generated
func (s *RSISignal) Timestamp() time.Time {
	return s.timestamp
}

// Payload returns the full oscillator data for downstream analysis
// Decision and risk engines can access the complete oscillator context
func (s *RSISignal) Payload() any {
	return s.oscillator
}
