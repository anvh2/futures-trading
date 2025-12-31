package testdata

import (
	"math/rand"
	"time"
)

// TestConfig holds configuration for testing
type TestConfig struct {
	// Default test symbols
	Symbols []string

	// Default test intervals
	Intervals []string

	// Default price ranges for different symbols
	PriceRanges map[string]PriceRange

	// Test settings
	DefaultSeed    int64
	DefaultVolume  float64
	DefaultCapital float64
	TestDuration   time.Duration
	MaxSignalAge   time.Duration

	// Risk thresholds for testing
	MaxRiskScore    float64
	MinConfidence   float64
	MaxPositionSize float64
}

type PriceRange struct {
	Min     float64
	Max     float64
	Default float64
}

// DefaultTestConfig returns a default test configuration
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		Symbols:   []string{"BTCUSDT", "ETHUSDT", "ADAUSDT", "BNBUSDT", "SOLUSDT"},
		Intervals: []string{"1m", "5m", "15m", "1h", "4h", "1d"},

		PriceRanges: map[string]PriceRange{
			"BTCUSDT": {Min: 20000, Max: 100000, Default: 50000},
			"ETHUSDT": {Min: 1000, Max: 5000, Default: 3000},
			"ADAUSDT": {Min: 0.2, Max: 2.0, Default: 0.5},
			"BNBUSDT": {Min: 200, Max: 800, Default: 400},
			"SOLUSDT": {Min: 50, Max: 300, Default: 150},
		},

		DefaultSeed:    42,
		DefaultVolume:  1000.0,
		DefaultCapital: 10000.0,
		TestDuration:   24 * time.Hour,
		MaxSignalAge:   1 * time.Hour,

		MaxRiskScore:    5.0,
		MinConfidence:   0.6,
		MaxPositionSize: 0.1, // 10% of capital
	}
}

// GetRandomSymbol returns a random symbol from the test config
func (tc *TestConfig) GetRandomSymbol() string {
	if len(tc.Symbols) == 0 {
		return "BTCUSDT"
	}
	return tc.Symbols[rand.Intn(len(tc.Symbols))]
}

// GetDefaultPrice returns the default price for a symbol
func (tc *TestConfig) GetDefaultPrice(symbol string) float64 {
	if priceRange, exists := tc.PriceRanges[symbol]; exists {
		return priceRange.Default
	}
	return 50000.0 // Default fallback
}

// GetPriceRange returns the price range for a symbol
func (tc *TestConfig) GetPriceRange(symbol string) PriceRange {
	if priceRange, exists := tc.PriceRanges[symbol]; exists {
		return priceRange
	}
	return PriceRange{Min: 1000, Max: 100000, Default: 50000}
}
