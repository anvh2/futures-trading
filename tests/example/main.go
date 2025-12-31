package main

import (
	"fmt"
	"log"

	"github.com/anvh2/futures-trading/tests/testdata"
)

func main() {
	fmt.Println("🚀 Futures Trading E2E Test Data Generator")
	fmt.Println("==========================================")

	// Initialize generator
	generator := testdata.NewMarketDataGenerator(42) // Fixed seed for reproducible results

	// Test 1: Generate basic market data
	fmt.Println("\n📊 Test 1: Generating basic market data...")
	testBasicMarketData(generator)

	// Test 2: Test market scenarios
	fmt.Println("\n🎭 Test 2: Testing market scenarios...")
	testMarketScenarios(generator)

	// Test 3: Generate trading signals and decisions
	fmt.Println("\n📈 Test 3: Generating trading signals and decisions...")
	testTradingFlow(generator)

	// Test 4: Test multi-symbol data
	fmt.Println("\n🔄 Test 4: Testing multi-symbol data generation...")
	testMultiSymbolData(generator)

	fmt.Println("\n✅ All tests completed successfully!")
	fmt.Println("\nTo run the full E2E test suite:")
	fmt.Println("  go test -v ./tests/e2e/...")
	fmt.Println("\nTo run specific scenarios:")
	fmt.Println("  go test -v ./tests/e2e/ -run TestMarketScenarios")
}

func testBasicMarketData(generator *testdata.MarketDataGenerator) {
	symbols := []string{"BTCUSDT", "ETHUSDT", "ADAUSDT"}
	intervals := []string{"1m", "5m", "1h", "4h"}

	for _, symbol := range symbols {
		fmt.Printf("  Generating data for %s...\n", symbol)

		summary := generator.GenerateCandleSummary(symbol, intervals)
		if summary == nil {
			log.Fatalf("Failed to generate candle summary for %s", symbol)
		}

		fmt.Printf("    Symbol: %s\n", summary.Symbol)
		for interval, data := range summary.Candles {
			if len(data.Candles) > 0 {
				firstCandle := data.Candles[0]
				lastCandle := data.Candles[len(data.Candles)-1]

				fmt.Printf("    %s: %d candles (%.2f -> %.2f)\n",
					interval,
					len(data.Candles),
					parsePrice(firstCandle.Close),
					parsePrice(lastCandle.Close))
			}
		}
	}
}

func testMarketScenarios(generator *testdata.MarketDataGenerator) {
	scenarios := []string{"bull_run", "bear_market", "sideways_chop", "flash_crash"}

	for _, scenarioName := range scenarios {
		fmt.Printf("  Testing scenario: %s\n", scenarioName)

		summary, err := generator.GenerateScenarioData(scenarioName, "1h")
		if err != nil {
			log.Printf("    Error generating scenario %s: %v", scenarioName, err)
			continue
		}

		candlesData := summary.Candles["1h"]
		if len(candlesData.Candles) == 0 {
			log.Printf("    No candles generated for scenario %s", scenarioName)
			continue
		}

		firstPrice := parsePrice(candlesData.Candles[0].Close)
		lastPrice := parsePrice(candlesData.Candles[len(candlesData.Candles)-1].Close)
		priceChange := (lastPrice - firstPrice) / firstPrice * 100

		fmt.Printf("    Price change: %.2f%% (%.2f -> %.2f)\n",
			priceChange, firstPrice, lastPrice)

		// Generate realistic decision input for this scenario
		scenario := testdata.PredefinedScenarios[scenarioName]
		decisionInput := generator.GenerateRealisticDecisionInput(scenario, lastPrice, 1.0)

		fmt.Printf("    RSI: %.1f, Trend: %s, Strength: %.1f\n",
			decisionInput.RSI, decisionInput.Trend_1h, decisionInput.TrendStrength)
	}
}

func testTradingFlow(generator *testdata.MarketDataGenerator) {
	symbol := "BTCUSDT"
	price := 50000.0

	// Generate decision input
	fmt.Printf("  Generating decision input for %s at $%.2f...\n", symbol, price)
	decisionInput := generator.GenerateDecisionInput(symbol, price)

	fmt.Printf("    RSI: %.1f\n", decisionInput.RSI)
	fmt.Printf("    Trend (1h): %s\n", decisionInput.Trend_1h)
	fmt.Printf("    Support: $%.2f, Resistance: $%.2f\n",
		decisionInput.SupportLevel, decisionInput.ResistanceLevel)
	fmt.Printf("    Position: %s, Capital: $%.2f\n",
		decisionInput.CurrentPosition, decisionInput.Capital)

	// Generate signal
	fmt.Printf("\n  Generating trading signal...\n")
	signal := generator.GenerateSignal(symbol, price)

	fmt.Printf("    Action: %s\n", signal.Action)
	fmt.Printf("    Strength: %.2f, Confidence: %.2f\n",
		signal.Strength, signal.Confidence)
	fmt.Printf("    Price: $%.2f\n", signal.Price)

	if signal.StopLoss > 0 {
		fmt.Printf("    Stop Loss: $%.2f\n", signal.StopLoss)
	}
	if signal.TakeProfit > 0 {
		fmt.Printf("    Take Profit: $%.2f\n", signal.TakeProfit)
	}

	// Show indicators
	fmt.Printf("    Indicators:\n")
	for name, value := range signal.Indicators {
		fmt.Printf("      %s: %.2f\n", name, value)
	}
}

func testMultiSymbolData(generator *testdata.MarketDataGenerator) {
	config := testdata.DefaultTestConfig()

	fmt.Printf("  Testing %d symbols...\n", len(config.Symbols))

	for _, symbol := range config.Symbols {
		priceRange := config.GetPriceRange(symbol)
		currentPrice := priceRange.Default

		// Generate some basic data
		decisionInput := generator.GenerateDecisionInput(symbol, currentPrice)
		signal := generator.GenerateSignal(symbol, currentPrice)

		fmt.Printf("    %s: $%.2f, RSI=%.1f, Signal=%s (%.2f conf)\n",
			symbol, currentPrice, decisionInput.RSI,
			signal.Action, signal.Confidence)
	}
}

func parsePrice(priceStr string) float64 {
	// Simple price parsing - in real implementation, use proper parsing
	var price float64
	if _, err := fmt.Sscanf(priceStr, "%f", &price); err == nil {
		return price
	}
	return 0.0
}
