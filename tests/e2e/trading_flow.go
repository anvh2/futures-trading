package e2e

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/anvh2/futures-trading/internal/config"
	"github.com/anvh2/futures-trading/internal/libs/logger"
	"github.com/anvh2/futures-trading/internal/models"
	"github.com/anvh2/futures-trading/tests/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// E2ETestSuite represents the end-to-end test suite
type E2ETestSuite struct {
	t         *testing.T
	logger    *logger.Logger
	generator *testdata.MarketDataGenerator
	config    config.Config
}

// NewE2ETestSuite creates a new E2E test suite
func NewE2ETestSuite(t *testing.T) *E2ETestSuite {
	logger := logger.NewDev()
	generator := testdata.NewMarketDataGenerator(0)

	return &E2ETestSuite{
		t:         t,
		logger:    logger,
		generator: generator,
		config:    config.Config{}, // Mock config
	}
}

// TestFullTradingFlow tests the complete trading flow from market data to order execution
func (suite *E2ETestSuite) TestFullTradingFlow() {
	symbol := "BTCUSDT"
	intervals := []string{"1m", "5m", "15m", "1h", "4h"}

	suite.t.Log("Starting E2E trading flow test")

	// Step 1: Generate market data (Market Service)
	suite.t.Log("Step 1: Generating market data")
	candleSummary := suite.generator.GenerateCandleSummary(symbol, intervals)
	require.NotNil(suite.t, candleSummary, "Candle summary should not be nil")
	assert.Equal(suite.t, symbol, candleSummary.Symbol)
	assert.Len(suite.t, candleSummary.Candles, len(intervals))

	// Validate candle data
	for interval, candlesData := range candleSummary.Candles {
		suite.t.Logf("Validating %s interval data", interval)
		assert.NotEmpty(suite.t, candlesData.Candles, "Candles should not be empty")
		assert.True(suite.t, len(candlesData.Candles) > 0, "Should have candle data")

		// Check first and last candle
		firstCandle := candlesData.Candles[0]
		lastCandle := candlesData.Candles[len(candlesData.Candles)-1]

		assert.True(suite.t, firstCandle.OpenTime < lastCandle.OpenTime, "Candles should be chronological")
		suite.validateCandlestick(firstCandle)
		suite.validateCandlestick(lastCandle)
	}

	// Extract current price from latest 1h candle
	currentPrice := suite.extractCurrentPrice(candleSummary)
	suite.t.Logf("Current price extracted: %f", currentPrice)

	// Step 2: Analyze market data (Analyze Service)
	suite.t.Log("Step 2: Analyzing market data")
	decisionInput := suite.generator.GenerateDecisionInput(symbol, currentPrice)
	require.NotNil(suite.t, decisionInput, "Decision input should not be nil")

	suite.validateDecisionInput(decisionInput)
	suite.t.Logf("Generated decision input: RSI=%.2f, Price=%.2f, Trend=%s",
		decisionInput.RSI, decisionInput.Price, decisionInput.Trend_1h)

	// Step 3: Generate trading signal (Signal Service)
	suite.t.Log("Step 3: Generating trading signal")
	signal := suite.generator.GenerateSignal(symbol, currentPrice)
	require.NotNil(suite.t, signal, "Signal should not be nil")

	suite.validateSignal(signal)
	suite.t.Logf("Generated signal: Action=%s, Strength=%.2f, Confidence=%.2f",
		signal.Action, signal.Strength, signal.Confidence)

	// Step 4: Make trading decision (Decision Engine)
	suite.t.Log("Step 4: Making trading decision")
	tradingDecision := suite.makeTradingDecision(decisionInput, signal)
	require.NotNil(suite.t, tradingDecision, "Trading decision should not be nil")

	suite.validateTradingDecision(tradingDecision)
	suite.t.Logf("Trading decision: Action=%s, Confidence=%.2f",
		tradingDecision.Action, tradingDecision.Confidence)

	// Step 5: Risk assessment (Risk Engine)
	suite.t.Log("Step 5: Performing risk assessment")
	riskAssessment := suite.assessRisk(tradingDecision, decisionInput)
	require.NotNil(suite.t, riskAssessment, "Risk assessment should not be nil")

	suite.validateRiskAssessment(riskAssessment)
	suite.t.Logf("Risk assessment: Approved=%t, RiskScore=%.2f",
		riskAssessment.Approved, riskAssessment.RiskScore)

	// Step 6: Generate order (Order Engine)
	if riskAssessment.Approved {
		suite.t.Log("Step 6: Generating order")
		order := suite.generateOrder(tradingDecision, currentPrice)
		require.NotNil(suite.t, order, "Order should not be nil")

		suite.validateOrder(order)
		suite.t.Logf("Generated order: Side=%s, Quantity=%s, Price=%s",
			order.Side, order.Quantity, order.Price)
	} else {
		suite.t.Log("Step 6: Order rejected by risk assessment")
	}

	suite.t.Log("E2E trading flow test completed successfully")
}

// TestMultiSymbolFlow tests the flow with multiple symbols
func (suite *E2ETestSuite) TestMultiSymbolFlow() {
	symbols := []string{"BTCUSDT", "ETHUSDT", "ADAUSDT"}
	intervals := []string{"1h", "4h"}

	suite.t.Log("Starting multi-symbol flow test")

	for _, symbol := range symbols {
		suite.t.Logf("Testing flow for symbol: %s", symbol)

		// Generate market data
		candleSummary := suite.generator.GenerateCandleSummary(symbol, intervals)
		assert.Equal(suite.t, symbol, candleSummary.Symbol)

		currentPrice := suite.extractCurrentPrice(candleSummary)

		// Generate signal and decision input
		signal := suite.generator.GenerateSignal(symbol, currentPrice)
		decisionInput := suite.generator.GenerateDecisionInput(symbol, currentPrice)

		// Validate they're consistent
		assert.Equal(suite.t, symbol, signal.Symbol)
		assert.Equal(suite.t, symbol, decisionInput.Symbol)

		suite.t.Logf("Symbol %s: Price=%.2f, Signal=%s", symbol, currentPrice, signal.Action)
	}

	suite.t.Log("Multi-symbol flow test completed")
}

// TestMarketConditions tests different market conditions
func (suite *E2ETestSuite) TestMarketConditions() {
	symbol := "BTCUSDT"
	conditions := []struct {
		name       string
		trend      string
		volatility float64
	}{
		{"Bull Market", "up", 0.02},
		{"Bear Market", "down", 0.03},
		{"Sideways Market", "sideways", 0.015},
		{"Volatile Market", "volatile", 0.05},
	}

	suite.t.Log("Starting market conditions test")

	for _, condition := range conditions {
		suite.t.Logf("Testing %s condition", condition.name)

		// Generate specific market condition
		config := testdata.CandleGeneratorConfig{
			Symbol:     symbol,
			Interval:   "1h",
			StartPrice: 50000.0,
			Count:      50,
			Trend:      condition.trend,
			Volatility: condition.volatility,
			VolumeBase: 1000.0,
			TickerLen:  time.Hour,
		}

		candles := suite.generator.GenerateCandles(config)
		assert.Len(suite.t, candles, 50)

		// Analyze trend
		firstPrice := suite.parsePrice(candles[0].Close)
		lastPrice := suite.parsePrice(candles[len(candles)-1].Close)
		priceChange := (lastPrice - firstPrice) / firstPrice

		suite.t.Logf("%s: Price change = %.2f%%", condition.name, priceChange*100)

		// Validate trend matches expectation
		if condition.trend == "up" {
			assert.True(suite.t, priceChange > -0.05, "Uptrend should have positive or small negative change")
		} else if condition.trend == "down" {
			assert.True(suite.t, priceChange < 0.05, "Downtrend should have negative or small positive change")
		}
	}

	suite.t.Log("Market conditions test completed")
}

// Helper methods for validation

func (suite *E2ETestSuite) validateCandlestick(candle *models.Candlestick) {
	assert.True(suite.t, candle.OpenTime > 0, "OpenTime should be positive")
	assert.True(suite.t, candle.CloseTime > candle.OpenTime, "CloseTime should be after OpenTime")

	high := suite.parsePrice(candle.High)
	low := suite.parsePrice(candle.Low)
	open := suite.parsePrice(candle.Open)
	close := suite.parsePrice(candle.Close)

	assert.True(suite.t, high >= low, "High should be >= Low")
	assert.True(suite.t, high >= open, "High should be >= Open")
	assert.True(suite.t, high >= close, "High should be >= Close")
	assert.True(suite.t, low <= open, "Low should be <= Open")
	assert.True(suite.t, low <= close, "Low should be <= Close")
}

func (suite *E2ETestSuite) validateDecisionInput(input *models.DecisionInput) {
	assert.NotEmpty(suite.t, input.Symbol, "Symbol should not be empty")
	assert.True(suite.t, input.Price > 0, "Price should be positive")
	assert.True(suite.t, input.RSI >= 0 && input.RSI <= 100, "RSI should be between 0-100")
	assert.True(suite.t, input.Capital > 0, "Capital should be positive")
}

func (suite *E2ETestSuite) validateSignal(signal *models.Signal) {
	assert.NotEmpty(suite.t, signal.Symbol, "Signal symbol should not be empty")
	assert.True(suite.t, signal.Price > 0, "Signal price should be positive")
	assert.True(suite.t, signal.Strength >= 0 && signal.Strength <= 1, "Strength should be 0-1")
	assert.True(suite.t, signal.Confidence >= 0 && signal.Confidence <= 1, "Confidence should be 0-1")
	assert.True(suite.t, signal.IsActive, "Signal should be active")
}

func (suite *E2ETestSuite) validateTradingDecision(decision *TradingDecision) {
	assert.NotEmpty(suite.t, decision.Symbol, "Decision symbol should not be empty")
	assert.True(suite.t, decision.Confidence >= 0 && decision.Confidence <= 1, "Confidence should be 0-1")
}

func (suite *E2ETestSuite) validateRiskAssessment(assessment *RiskAssessment) {
	assert.True(suite.t, assessment.RiskScore >= 0, "Risk score should be non-negative")
	assert.NotEmpty(suite.t, assessment.Factors, "Risk factors should not be empty")
}

func (suite *E2ETestSuite) validateOrder(order *models.Order) {
	assert.NotEmpty(suite.t, order.Symbol, "Order symbol should not be empty")
	assert.NotEmpty(suite.t, string(order.Side), "Order side should not be empty")
	assert.NotEmpty(suite.t, order.Quantity, "Order quantity should not be empty")
}

func (suite *E2ETestSuite) extractCurrentPrice(summary *models.CandleSummary) float64 {
	// Get latest 1h candle
	candlesData, ok := summary.Candles["1h"]
	if !ok || len(candlesData.Candles) == 0 {
		// Fallback to any available interval
		for _, data := range summary.Candles {
			if len(data.Candles) > 0 {
				candlesData = data
				break
			}
		}
	}

	if candlesData == nil || len(candlesData.Candles) == 0 {
		return 50000.0 // Default price
	}

	lastCandle := candlesData.Candles[len(candlesData.Candles)-1]
	return suite.parsePrice(lastCandle.Close)
}

func (suite *E2ETestSuite) parsePrice(priceStr string) float64 {
	var price float64
	if err := json.Unmarshal([]byte(`"`+priceStr+`"`), &price); err != nil {
		// Fallback parsing
		if priceStr != "" {
			if p, err := time.Parse("2006-01-02", priceStr); err == nil {
				return float64(p.Unix())
			}
		}
		return 0
	}
	return price
}

// Mock trading decision and risk assessment structures
type TradingDecision struct {
	Symbol     string  `json:"symbol"`
	Action     string  `json:"action"`
	Confidence float64 `json:"confidence"`
	Quantity   float64 `json:"quantity"`
	Price      float64 `json:"price"`
	Reasoning  string  `json:"reasoning"`
}

type RiskAssessment struct {
	Approved  bool               `json:"approved"`
	RiskScore float64            `json:"risk_score"`
	Factors   map[string]float64 `json:"factors"`
	Warnings  []string           `json:"warnings"`
}

// Mock implementations of decision and risk engines
func (suite *E2ETestSuite) makeTradingDecision(input *models.DecisionInput, signal *models.Signal) *TradingDecision {
	// Simple decision logic based on signal and input
	confidence := (signal.Confidence + signal.Strength) / 2

	// Reduce confidence if RSI is extreme
	if input.RSI > 80 || input.RSI < 20 {
		confidence *= 0.8
	}

	// Increase confidence for trend alignment
	if (signal.Action == models.SignalActionBuy && input.Trend_1h == "UP") ||
		(signal.Action == models.SignalActionSell && input.Trend_1h == "DOWN") {
		confidence *= 1.1
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return &TradingDecision{
		Symbol:     signal.Symbol,
		Action:     string(signal.Action),
		Confidence: confidence,
		Quantity:   0.01, // Mock quantity
		Price:      signal.Price,
		Reasoning:  "Generated by E2E test",
	}
}

func (suite *E2ETestSuite) assessRisk(decision *TradingDecision, input *models.DecisionInput) *RiskAssessment {
	riskScore := 0.0
	factors := make(map[string]float64)
	warnings := []string{}

	// Calculate risk factors

	// Volatility risk
	volatilityRisk := input.ATRPercent * 10 // Scale ATR to risk score
	factors["volatility"] = volatilityRisk
	riskScore += volatilityRisk

	// Position size risk
	positionValue := decision.Quantity * decision.Price
	positionRisk := (positionValue / input.Capital) * 5 // 5x multiplier
	factors["position_size"] = positionRisk
	riskScore += positionRisk

	// Market conditions risk
	marketRisk := 0.0
	if input.RSI > 85 || input.RSI < 15 {
		marketRisk += 2.0
		warnings = append(warnings, "Extreme RSI levels detected")
	}
	factors["market_conditions"] = marketRisk
	riskScore += marketRisk

	// Funding rate risk for futures
	if input.FundingRate > 0.01 || input.FundingRate < -0.01 {
		fundingRisk := math.Abs(input.FundingRate) * 100
		factors["funding_rate"] = fundingRisk
		riskScore += fundingRisk
		warnings = append(warnings, "High funding rate detected")
	}

	// Decision: approve if risk score is reasonable and confidence is high
	approved := riskScore < 5.0 && decision.Confidence > 0.6

	return &RiskAssessment{
		Approved:  approved,
		RiskScore: riskScore,
		Factors:   factors,
		Warnings:  warnings,
	}
}

func (suite *E2ETestSuite) generateOrder(decision *TradingDecision, currentPrice float64) *models.Order {
	// Convert decision to order
	order := &models.Order{
		Symbol:   decision.Symbol,
		Quantity: "0.01",                     // Mock quantity as string
		Price:    fmt.Sprint(decision.Price), // Use decision price
	}

	// Set order side based on action
	if decision.Action == "BUY" || decision.Action == string(models.SignalActionBuy) {
		order.Side = "BUY"
	} else if decision.Action == "SELL" || decision.Action == string(models.SignalActionSell) {
		order.Side = "SELL"
	}

	return order
}

// Test runner
func TestE2ETradingFlow(t *testing.T) {
	suite := NewE2ETestSuite(t)
	suite.TestFullTradingFlow()
}

func TestE2EMultiSymbol(t *testing.T) {
	suite := NewE2ETestSuite(t)
	suite.TestMultiSymbolFlow()
}

func TestE2EMarketConditions(t *testing.T) {
	suite := NewE2ETestSuite(t)
	suite.TestMarketConditions()
}
