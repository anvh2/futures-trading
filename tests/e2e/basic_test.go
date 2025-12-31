package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/anvh2/futures-trading/tests/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBasicMarketDataGeneration tests basic market data generation
func TestBasicMarketDataGeneration(t *testing.T) {
	t.Log("📊 Testing basic market data generation")

	generator := testdata.NewMarketDataGenerator(42)
	symbols := []string{"BTCUSDT", "ETHUSDT", "ADAUSDT"}
	intervals := []string{"1m", "5m", "15m"}

	for _, symbol := range symbols {
		t.Run(symbol, func(t *testing.T) {
			for _, interval := range intervals {
				config := testdata.CandleGeneratorConfig{
					Symbol:     symbol,
					Interval:   interval,
					StartPrice: 50000.0,
					Count:      100,
					Trend:      "up",
					Volatility: 0.02,
					VolumeBase: 100.0,
					TickerLen:  5 * time.Minute,
				}

				candles := generator.GenerateCandles(config)
				require.Len(t, candles, 100, "Should generate 100 candles")
				require.NotEmpty(t, candles[0].Open, "Open price should not be empty")
				require.NotEmpty(t, candles[0].Close, "Close price should not be empty")
			}
			t.Logf("✓ %s: Generated 3 intervals with complete data", symbol)
		})
	}
}

// TestMarketScenarios tests different market scenarios
func TestMarketScenarios(t *testing.T) {
	t.Log("🎭 Testing market scenarios")

	scenarios := []string{"bull_run", "bear_market", "sideways_chop", "flash_crash"}
	generator := testdata.NewMarketDataGenerator(123)

	for _, scenario := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			scenarioData, err := generator.GenerateScenarioData(scenario, "5m")
			require.NoError(t, err, "Should generate scenario data")
			require.NotEmpty(t, scenarioData.Candles, "Should have candles")

			// Get candles for first interval
			firstInterval := ""
			for interval := range scenarioData.Candles {
				firstInterval = interval
				break
			}

			candles := scenarioData.Candles[firstInterval].Candles
			require.NotEmpty(t, candles, "Should have candles")

			// Calculate price change
			firstPrice := parseFloat(candles[0].Open)
			lastPrice := parseFloat(candles[len(candles)-1].Close)
			priceChange := ((lastPrice - firstPrice) / firstPrice) * 100

			t.Logf("✓ %s: %.1f%% price change (%.2f -> %.2f)",
				scenario, priceChange, firstPrice, lastPrice)
		})
	}
}

// TestTradingSignalGeneration tests trading signal generation
func TestTradingSignalGeneration(t *testing.T) {
	t.Log("📈 Testing trading signal generation")

	symbols := []string{"BTC", "ETH", "ADA"}
	generator := testdata.NewMarketDataGenerator(456)

	for _, symbol := range symbols {
		t.Run(symbol, func(t *testing.T) {
			config := testdata.CandleGeneratorConfig{
				Symbol:     symbol + "USDT",
				Interval:   "5m",
				StartPrice: 45000.0,
				Count:      50,
				Trend:      "up",
				Volatility: 0.03,
				VolumeBase: 200.0,
				TickerLen:  5 * time.Minute,
			}

			candles := generator.GenerateCandles(config)
			require.NotEmpty(t, candles, "Should generate candles")

			// Generate decision input from market data
			decisionInput := generator.GenerateDecisionInput(symbol+"USDT", 45000.0)
			require.NotNil(t, decisionInput, "Should generate decision input")

			// Generate signal from decision input
			signal := generator.GenerateSignal(symbol+"USDT", 45000.0)
			require.NotNil(t, signal, "Should generate signal")
			require.True(t, signal.Confidence >= 0 && signal.Confidence <= 1, "Confidence should be between 0 and 1")

			t.Logf("✓ %s: RSI=%.1f, Signal=%s (%.1f%% confidence)",
				symbol, decisionInput.RSI, signal.Action, signal.Confidence*100)
		})
	}
}

// TestCandleDataIntegrity tests candle data integrity
func TestCandleDataIntegrity(t *testing.T) {
	t.Log("🔍 Testing candle data integrity")

	generator := testdata.NewMarketDataGenerator(789)
	config := testdata.CandleGeneratorConfig{
		Symbol:     "BTCUSDT",
		Interval:   "1h",
		StartPrice: 50000.0,
		Count:      100,
		Trend:      "sideways",
		Volatility: 0.02,
		VolumeBase: 150.0,
		TickerLen:  time.Hour,
	}

	candles := generator.GenerateCandles(config)
	require.Len(t, candles, 100, "Should generate exactly 100 candles")

	for i, candle := range candles {
		// Basic field validation
		require.NotEmpty(t, candle.Open, "Open should not be empty")
		require.NotEmpty(t, candle.High, "High should not be empty")
		require.NotEmpty(t, candle.Low, "Low should not be empty")
		require.NotEmpty(t, candle.Close, "Close should not be empty")
		require.NotEmpty(t, candle.Volume, "Volume should not be empty")
		require.True(t, candle.OpenTime > 0, "OpenTime should be set")
		require.True(t, candle.CloseTime > candle.OpenTime, "CloseTime should be after OpenTime")

		// Price relationship validation
		high := parseFloat(candle.High)
		low := parseFloat(candle.Low)
		open := parseFloat(candle.Open)
		close := parseFloat(candle.Close)

		assert.True(t, high >= open, "Candle %d: High should be >= Open", i)
		assert.True(t, high >= close, "Candle %d: High should be >= Close", i)
		assert.True(t, low <= open, "Candle %d: Low should be <= Open", i)
		assert.True(t, low <= close, "Candle %d: Low should be <= Close", i)
		assert.True(t, high >= low, "Candle %d: High should be >= Low", i)
	}

	t.Logf("✓ Validated %d candles for data integrity", len(candles))
}

// TestHighVolatilityScenario tests high volatility scenarios
func TestHighVolatilityScenario(t *testing.T) {
	t.Log("💥 Testing high volatility scenario")

	generator := testdata.NewMarketDataGenerator(999)
	config := testdata.CandleGeneratorConfig{
		Symbol:     "BTCUSDT",
		Interval:   "1m",
		StartPrice: 45000.0,
		Count:      200,
		Trend:      "volatile",
		Volatility: 0.05, // 5% volatility (very high)
		VolumeBase: 500.0,
		TickerLen:  time.Minute,
	}

	candles := generator.GenerateCandles(config)
	require.Len(t, candles, 200, "Should generate 200 candles")

	// Calculate price statistics
	var minPrice, maxPrice float64
	prices := make([]float64, len(candles))

	for i, candle := range candles {
		price := parseFloat(candle.Close)
		prices[i] = price

		if i == 0 {
			minPrice = price
			maxPrice = price
		} else {
			if price < minPrice {
				minPrice = price
			}
			if price > maxPrice {
				maxPrice = price
			}
		}
	}

	priceRange := ((maxPrice - minPrice) / minPrice) * 100
	require.True(t, priceRange > 10, "High volatility should show significant price range")

	t.Logf("✓ High volatility test: %.1f%% price range ($%.2f - $%.2f)",
		priceRange, minPrice, maxPrice)

	// Test system response to volatility
	decisionInput := generator.GenerateDecisionInput("BTCUSDT", maxPrice)
	signal := generator.GenerateSignal("BTCUSDT", maxPrice)

	require.NotNil(t, signal, "System should handle high volatility")
	require.True(t, signal.Confidence >= 0, "Confidence should be valid even in high volatility")

	t.Logf("✓ System handles high volatility: RSI=%.1f, Signal=%s",
		decisionInput.RSI, signal.Action)
}

// Helper function to parse float from string
func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	// Simple float parsing for test purposes
	var f float64
	if _, err := fmt.Sscanf(s, "%f", &f); err != nil {
		return 0
	}
	return f
}
