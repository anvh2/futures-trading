package e2e

// import (
// 	"fmt"
// 	"strconv"
// 	"testing"
// 	"time"

// 	"github.com/anvh2/futures-trading/internal/models"
// 	"github.com/anvh2/futures-trading/tests/testdata"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// // TestBasicMarketDataGeneration tests basic market data generation
// func TestBasicMarketDataGeneration(t *testing.T) {
// 	t.Log("📊 Testing basic market data generation")

// 	generator := testdata.NewMarketDataGenerator(42)

// 	symbols := []string{"BTCUSDT", "ETHUSDT", "ADAUSDT"}
// 	intervals := []string{"1m", "5m", "1h"}

// 	for _, symbol := range symbols {
// 		t.Run(symbol, func(t *testing.T) {
// 			summary := generator.GenerateCandleSummary(symbol, intervals)
// 			require.NotNil(t, summary)
// 			assert.Equal(t, symbol, summary.Symbol)

// 			for _, interval := range intervals {
// 				assert.Contains(t, summary.Candles, interval)
// 				candlesData := summary.Candles[interval]
// 				assert.True(t, len(candlesData.Candles) > 0)

// 				// Validate first and last candle
// 				firstCandle := candlesData.Candles[0]
// 				lastCandle := candlesData.Candles[len(candlesData.Candles)-1]

// 				validateCandleBasic(t, firstCandle, fmt.Sprintf("%s-%s[0]", symbol, interval))
// 				validateCandleBasic(t, lastCandle, fmt.Sprintf("%s-%s[last]", symbol, interval))
// 			}

// 			t.Logf("✓ %s: Generated %d intervals with complete data", symbol, len(intervals))
// 		})
// 	}
// }

// // TestMarketScenarios tests market scenarios
// func TestMarketScenarios(t *testing.T) {
// 	t.Log("🎭 Testing market scenarios")

// 	generator := testdata.NewMarketDataGenerator(42)

// 	scenarios := []string{"bull_run", "bear_market", "sideways_chop", "flash_crash"}

// 	for _, scenarioName := range scenarios {
// 		t.Run(scenarioName, func(t *testing.T) {
// 			summary, err := generator.GenerateScenarioData(scenarioName, "1h")
// 			require.NoError(t, err)
// 			require.NotNil(t, summary)

// 			candlesData := summary.Candles["1h"]
// 			assert.True(t, len(candlesData.Candles) > 0)

// 			firstPrice := parseFloatSafe(candlesData.Candles[0].Close)
// 			lastPrice := parseFloatSafe(candlesData.Candles[len(candlesData.Candles)-1].Close)
// 			priceChange := (lastPrice - firstPrice) / firstPrice * 100

// 			t.Logf("✓ %s: %.1f%% price change (%.2f -> %.2f)",
// 				scenarioName, priceChange, firstPrice, lastPrice)
// 		})
// 	}
// }

// // TestTradingSignalGeneration tests trading signal generation
// func TestTradingSignalGeneration(t *testing.T) {
// 	t.Log("📈 Testing trading signal generation")

// 	generator := testdata.NewMarketDataGenerator(42)

// 	testCases := []struct {
// 		name   string
// 		symbol string
// 		price  float64
// 	}{
// 		{"BTC", "BTCUSDT", 50000.0},
// 		{"ETH", "ETHUSDT", 3000.0},
// 		{"ADA", "ADAUSDT", 0.5},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			// Generate decision input
// 			decisionInput := generator.GenerateDecisionInput(tc.symbol, tc.price)
// 			require.NotNil(t, decisionInput)

// 			assert.Equal(t, tc.symbol, decisionInput.Symbol)
// 			assert.Equal(t, tc.price, decisionInput.Price)
// 			assert.True(t, decisionInput.RSI >= 0 && decisionInput.RSI <= 100)
// 			assert.True(t, decisionInput.Capital > 0)

// 			// Generate signal
// 			signal := generator.GenerateSignal(tc.symbol, tc.price)
// 			require.NotNil(t, signal)

// 			assert.Equal(t, tc.symbol, signal.Symbol)
// 			assert.Equal(t, tc.price, signal.Price)
// 			assert.True(t, signal.Strength >= 0 && signal.Strength <= 1)
// 			assert.True(t, signal.Confidence >= 0 && signal.Confidence <= 1)
// 			assert.True(t, signal.IsActive)

// 			t.Logf("✓ %s: RSI=%.1f, Signal=%s (%.1f%% confidence)",
// 				tc.name, decisionInput.RSI, signal.Action, signal.Confidence*100)
// 		})
// 	}
// }

// // TestCandleDataIntegrity tests candle data integrity
// func TestCandleDataIntegrity(t *testing.T) {
// 	t.Log("🔍 Testing candle data integrity")

// 	generator := testdata.NewMarketDataGenerator(time.Now().UnixNano())

// 	config := testdata.CandleGeneratorConfig{
// 		Symbol:     "BTCUSDT",
// 		Interval:   "1h",
// 		StartPrice: 50000.0,
// 		Count:      100,
// 		Trend:      "up",
// 		Volatility: 0.02,
// 		VolumeBase: 1000.0,
// 		TickerLen:  time.Hour,
// 	}

// 	candles := generator.GenerateCandles(config)
// 	require.Len(t, candles, 100)

// 	for i, candle := range candles {
// 		validateCandleBasic(t, candle, fmt.Sprintf("candle[%d]", i))

// 		// Test chronological order
// 		if i > 0 {
// 			prevCandle := candles[i-1]
// 			assert.True(t, candle.OpenTime > prevCandle.OpenTime,
// 				"Candle %d should be after candle %d", i, i-1)
// 		}
// 	}

// 	t.Logf("✓ Validated %d candles for data integrity", len(candles))
// }

// // TestHighVolatilityScenario tests high volatility scenarios
// func TestHighVolatilityScenario(t *testing.T) {
// 	t.Log("💥 Testing high volatility scenario")

// 	generator := testdata.NewMarketDataGenerator(0)

// 	config := testdata.CandleGeneratorConfig{
// 		Symbol:     "BTCUSDT",
// 		Interval:   "5m",
// 		StartPrice: 50000.0,
// 		Count:      50,
// 		Trend:      "volatile",
// 		Volatility: 0.10, // 10% volatility
// 		VolumeBase: 1000.0,
// 		TickerLen:  5 * time.Minute,
// 	}

// 	candles := generator.GenerateCandles(config)
// 	require.Len(t, candles, 50)

// 	// Calculate price range
// 	var minPrice, maxPrice float64
// 	prices := make([]float64, len(candles))

// 	for i, candle := range candles {
// 		price := parseFloatSafe(candle.Close)
// 		prices[i] = price

// 		if i == 0 {
// 			minPrice = price
// 			maxPrice = price
// 		} else {
// 			if price < minPrice {
// 				minPrice = price
// 			}
// 			if price > maxPrice {
// 				maxPrice = price
// 			}
// 		}
// 	}

// 	priceRange := (maxPrice - minPrice) / minPrice * 100

// 	t.Logf("✓ High volatility test: %.1f%% price range ($%.2f - $%.2f)",
// 		priceRange, minPrice, maxPrice)

// 	// Test decision making under high volatility
// 	lastPrice := prices[len(prices)-1]
// 	decisionInput := generator.GenerateDecisionInput("BTCUSDT", lastPrice)
// 	signal := generator.GenerateSignal("BTCUSDT", lastPrice)

// 	assert.NotNil(t, decisionInput)
// 	assert.NotNil(t, signal)

// 	t.Logf("✓ System handles high volatility: RSI=%.1f, Signal=%s",
// 		decisionInput.RSI, signal.Action)
// }

// // Helper functions

// func validateCandleBasic(t *testing.T, candle *models.Candlestick, context string) {
// 	assert.True(t, candle.OpenTime > 0, "%s: OpenTime should be positive", context)
// 	assert.True(t, candle.CloseTime > candle.OpenTime, "%s: CloseTime should be after OpenTime", context)

// 	high := parseFloatSafe(candle.High)
// 	low := parseFloatSafe(candle.Low)
// 	open := parseFloatSafe(candle.Open)
// 	close := parseFloatSafe(candle.Close)
// 	volume := parseFloatSafe(candle.Volume)

// 	assert.True(t, high >= low, "%s: High should be >= Low", context)
// 	assert.True(t, high >= open, "%s: High should be >= Open", context)
// 	assert.True(t, high >= close, "%s: High should be >= Close", context)
// 	assert.True(t, low <= open, "%s: Low should be <= Open", context)
// 	assert.True(t, low <= close, "%s: Low should be <= Close", context)
// 	assert.True(t, volume >= 0, "%s: Volume should be non-negative", context)
// }

// func parseFloatSafe(s string) float64 {
// 	if val, err := strconv.ParseFloat(s, 64); err == nil {
// 		return val
// 	}
// 	return 0.0
// }
