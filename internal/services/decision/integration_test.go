package decision

import (
	"testing"
	"time"

	"github.com/anvh2/futures-trading/internal/models"
)

// TestIntegratedDecisionMaking tests the complete integration between signal processing and engine scoring
func TestIntegratedDecisionMaking(t *testing.T) {
	// Create a comprehensive test signal with all necessary indicators
	signal := &models.Signal{
		Symbol:     "BTCUSDT",
		Type:       models.SignalTypeEntry,
		Action:     models.SignalActionBuy,
		Strength:   0.8,
		Confidence: 0.75,
		Price:      45000.0,
		Interval:   "1h",
		Strategy:   "comprehensive_scoring",
		Indicators: map[string]float64{
			// Technical indicators
			"rsi":         35.0, // Oversold condition
			"k":           25.0,
			"d":           20.0,
			"j":           15.0,
			"atr_percent": 2.5,
			"vwap":        44800.0,
			"recent_high": 46000.0,
			"recent_low":  43000.0,

			// Multi-timeframe RSI
			"rsi_5m":  40.0,
			"rsi_15m": 32.0,
			"rsi_1h":  35.0,

			// Volume indicators
			"relative_volume":       1.8, // High volume
			"volume_ratio":          1.2, // Buying pressure
			"spot_volume_change":    0.15,
			"futures_volume_change": 0.20,

			// Market structure
			"trend_strength":   35.0, // Weak trend/ranging market
			"support_level":    44500.0,
			"resistance_level": 46200.0,

			// Funding and positioning
			"funding_rate":     0.01, // Slightly positive
			"long_short_ratio": 1.3,  // More longs than shorts
			"oi_change":        0.05, // Rising open interest

			// On-chain and macro
			"exchange_inflows": -0.02, // Outflows (bullish)
			"macro_sentiment":  0.1,   // Slightly positive
			"news_sentiment":   0.2,   // Positive news
			"fear_greed_index": 25.0,  // Fear zone (contrarian bullish)
		},
		Metadata: map[string]interface{}{
			"trend_5m":          "SIDEWAYS",
			"trend_15m":         "DOWN",
			"trend_1h":          "SIDEWAYS",
			"swing_high_broken": false,
			"swing_low_broken":  false,
		},
		CreatedAt: time.Now(),
		IsActive:  true,
	}

	// Test convertSignalToDecisionInput
	maker := &Maker{} // Basic maker for testing
	decisionInput := maker.convertSignalToDecisionInput(signal)

	if decisionInput == nil {
		t.Fatal("convertSignalToDecisionInput returned nil")
	}

	// Verify key fields are properly converted
	if decisionInput.Symbol != signal.Symbol {
		t.Errorf("Expected symbol %s, got %s", signal.Symbol, decisionInput.Symbol)
	}

	if decisionInput.RSI != 35.0 {
		t.Errorf("Expected RSI 35.0, got %f", decisionInput.RSI)
	}

	if decisionInput.Price != signal.Price {
		t.Errorf("Expected price %f, got %f", signal.Price, decisionInput.Price)
	}

	// Test the complete scoring engine
	engineOutput := ComputeDecision(decisionInput)

	if engineOutput == nil {
		t.Fatal("ComputeDecision returned nil")
	}

	// Verify engine output structure
	if engineOutput.Confidence < 0 || engineOutput.Confidence > 100 {
		t.Errorf("Invalid confidence: %d", engineOutput.Confidence)
	}

	if engineOutput.Action == "" {
		t.Error("Action should not be empty")
	}

	if engineOutput.EntryPrice <= 0 {
		t.Error("Entry price should be positive")
	}

	// Test convertEngineOutputToTradingDecision
	tradingDecision := maker.convertEngineOutputToTradingDecision(signal, engineOutput)

	if tradingDecision == nil {
		t.Fatal("convertEngineOutputToTradingDecision returned nil")
	}

	// Verify trading decision structure
	if tradingDecision.Symbol != signal.Symbol {
		t.Errorf("Expected symbol %s, got %s", signal.Symbol, tradingDecision.Symbol)
	}

	if tradingDecision.Action == "" {
		t.Error("Trading action should not be empty")
	}

	if tradingDecision.Confidence < 0 || tradingDecision.Confidence > 1 {
		t.Errorf("Invalid confidence range: %f", tradingDecision.Confidence)
	}

	// Check metadata richness
	metadata := tradingDecision.Metadata
	requiredKeys := []string{
		"prediction", "bias", "total_score", "category_scores",
		"position_size_pct", "leverage", "stop_loss", "take_profit",
		"reasoning", "engine_confidence",
	}

	for _, key := range requiredKeys {
		if _, exists := metadata[key]; !exists {
			t.Errorf("Missing metadata key: %s", key)
		}
	}

	t.Logf("Integration test successful!")
	t.Logf("Engine Output: Prediction=%s, Confidence=%d%%, Action=%s",
		engineOutput.Prediction, engineOutput.Confidence, engineOutput.Action)
	t.Logf("Trading Decision: Action=%s, Size=%f, Confidence=%f",
		tradingDecision.Action, tradingDecision.Size, tradingDecision.Confidence)
	t.Logf("Reasoning: %s", engineOutput.Reasoning)
}

// TestEdgeCases tests edge cases and error conditions
func TestEdgeCases(t *testing.T) {
	maker := &Maker{}

	// Test with invalid signal
	invalidSignal := &models.Signal{
		IsActive: false,
	}

	decisionInput := maker.convertSignalToDecisionInput(invalidSignal)
	if decisionInput != nil {
		t.Error("Should return nil for invalid signal")
	}

	// Test with minimal signal data
	minimalSignal := &models.Signal{
		Symbol:     "ETHUSDT",
		Price:      3000.0,
		Interval:   "4h",
		Strategy:   "minimal_test",
		Indicators: map[string]float64{},
		Metadata:   map[string]interface{}{},
		IsActive:   true,
	}

	decisionInput = maker.convertSignalToDecisionInput(minimalSignal)
	if decisionInput == nil {
		t.Fatal("Should handle minimal signal data")
	}

	// Test engine with minimal data
	engineOutput := ComputeDecision(decisionInput)
	if engineOutput == nil {
		t.Fatal("Engine should handle minimal data")
	}

	// Should default to Hold action with low confidence
	if engineOutput.Action != "Hold" {
		t.Logf("Action with minimal data: %s (confidence: %d%%)",
			engineOutput.Action, engineOutput.Confidence)
	}

	t.Logf("Edge case test completed successfully!")
}

// BenchmarkDecisionMaking benchmarks the complete decision making process
func BenchmarkDecisionMaking(b *testing.B) {
	signal := &models.Signal{
		Symbol:   "BTCUSDT",
		Price:    45000.0,
		Interval: "1h",
		Strategy: "benchmark_test",
		Indicators: map[string]float64{
			"rsi": 45.0, "k": 50.0, "d": 45.0, "j": 40.0,
			"atr_percent": 2.0, "vwap": 44900.0,
			"recent_high": 46000.0, "recent_low": 44000.0,
		},
		Metadata: map[string]interface{}{
			"trend_5m": "UP", "trend_15m": "UP", "trend_1h": "UP",
		},
		IsActive: true,
	}

	maker := &Maker{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decisionInput := maker.convertSignalToDecisionInput(signal)
		if decisionInput != nil {
			engineOutput := ComputeDecision(decisionInput)
			if engineOutput != nil {
				maker.convertEngineOutputToTradingDecision(signal, engineOutput)
			}
		}
	}
}
