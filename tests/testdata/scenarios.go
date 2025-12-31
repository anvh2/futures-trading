package testdata

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/anvh2/futures-trading/internal/models"
)

// MarketScenario represents a predefined market scenario for testing
type MarketScenario struct {
	Name        string
	Description string
	Symbol      string
	Duration    time.Duration
	Trend       string
	Volatility  float64
	StartPrice  float64
	Events      []MarketEvent
}

// MarketEvent represents a significant market event during the scenario
type MarketEvent struct {
	Time        time.Duration // Relative to scenario start
	Type        string        // "pump", "dump", "consolidation", "breakout"
	Magnitude   float64       // Percentage change
	Duration    time.Duration // How long the event lasts
	Description string
}

// PredefinedScenarios contains common market scenarios for testing
var PredefinedScenarios = map[string]*MarketScenario{
	"bull_run": {
		Name:        "Bull Run",
		Description: "Strong upward trend with occasional corrections",
		Symbol:      "BTCUSDT",
		Duration:    24 * time.Hour,
		Trend:       "up",
		Volatility:  0.03,
		StartPrice:  45000.0,
		Events: []MarketEvent{
			{Time: 4 * time.Hour, Type: "pump", Magnitude: 0.08, Duration: 2 * time.Hour, Description: "Morning pump"},
			{Time: 8 * time.Hour, Type: "correction", Magnitude: -0.03, Duration: 1 * time.Hour, Description: "Small correction"},
			{Time: 16 * time.Hour, Type: "breakout", Magnitude: 0.12, Duration: 3 * time.Hour, Description: "Evening breakout"},
		},
	},
	"bear_market": {
		Name:        "Bear Market",
		Description: "Sustained downward pressure with dead cat bounces",
		Symbol:      "BTCUSDT",
		Duration:    24 * time.Hour,
		Trend:       "down",
		Volatility:  0.04,
		StartPrice:  55000.0,
		Events: []MarketEvent{
			{Time: 2 * time.Hour, Type: "dump", Magnitude: -0.06, Duration: 1 * time.Hour, Description: "Early dump"},
			{Time: 8 * time.Hour, Type: "bounce", Magnitude: 0.03, Duration: 30 * time.Minute, Description: "Dead cat bounce"},
			{Time: 18 * time.Hour, Type: "dump", Magnitude: -0.10, Duration: 2 * time.Hour, Description: "Capitulation"},
		},
	},
	"sideways_chop": {
		Name:        "Sideways Chop",
		Description: "Range-bound market with false breakouts",
		Symbol:      "ETHUSDT",
		Duration:    12 * time.Hour,
		Trend:       "sideways",
		Volatility:  0.025,
		StartPrice:  3000.0,
		Events: []MarketEvent{
			{Time: 2 * time.Hour, Type: "false_breakout", Magnitude: 0.04, Duration: 30 * time.Minute, Description: "False upside break"},
			{Time: 3 * time.Hour, Type: "rejection", Magnitude: -0.05, Duration: 45 * time.Minute, Description: "Rejection back to range"},
			{Time: 8 * time.Hour, Type: "false_breakdown", Magnitude: -0.04, Duration: 30 * time.Minute, Description: "False downside break"},
			{Time: 9 * time.Hour, Type: "bounce", Magnitude: 0.04, Duration: 30 * time.Minute, Description: "Bounce back to range"},
		},
	},
	"flash_crash": {
		Name:        "Flash Crash",
		Description: "Sudden sharp decline followed by recovery",
		Symbol:      "BTCUSDT",
		Duration:    6 * time.Hour,
		Trend:       "volatile",
		Volatility:  0.08,
		StartPrice:  50000.0,
		Events: []MarketEvent{
			{Time: 30 * time.Minute, Type: "normal", Magnitude: 0.01, Duration: 30 * time.Minute, Description: "Normal trading"},
			{Time: 1 * time.Hour, Type: "flash_crash", Magnitude: -0.15, Duration: 15 * time.Minute, Description: "Flash crash"},
			{Time: 80 * time.Minute, Type: "recovery", Magnitude: 0.12, Duration: 2 * time.Hour, Description: "V-shaped recovery"},
		},
	},
	"squeeze_breakout": {
		Name:        "Squeeze Breakout",
		Description: "Low volatility followed by explosive move",
		Symbol:      "ADAUSDT",
		Duration:    8 * time.Hour,
		Trend:       "up",
		Volatility:  0.015,
		StartPrice:  0.50,
		Events: []MarketEvent{
			{Time: 0, Type: "consolidation", Magnitude: 0.005, Duration: 4 * time.Hour, Description: "Tight consolidation"},
			{Time: 4 * time.Hour, Type: "squeeze_break", Magnitude: 0.20, Duration: 2 * time.Hour, Description: "Explosive breakout"},
			{Time: 6 * time.Hour, Type: "continuation", Magnitude: 0.08, Duration: 2 * time.Hour, Description: "Follow-through"},
		},
	},
}

// GenerateScenarioData generates market data based on a predefined scenario
func (g *MarketDataGenerator) GenerateScenarioData(scenarioName string, interval string) (*models.CandleSummary, error) {
	scenario, exists := PredefinedScenarios[scenarioName]
	if !exists {
		return nil, fmt.Errorf("scenario %s not found", scenarioName)
	}

	// Calculate number of candles based on scenario duration and interval
	intervalDuration := g.parseInterval(interval)
	numCandles := int(scenario.Duration / intervalDuration)

	// Generate base candles
	config := CandleGeneratorConfig{
		Symbol:     scenario.Symbol,
		Interval:   interval,
		StartPrice: scenario.StartPrice,
		Count:      numCandles,
		Trend:      scenario.Trend,
		Volatility: scenario.Volatility,
		VolumeBase: g.getVolumeForSymbol(scenario.Symbol),
		TickerLen:  intervalDuration,
	}

	candles := g.GenerateCandles(config)

	// Apply market events
	for _, event := range scenario.Events {
		g.applyMarketEvent(candles, event, intervalDuration, scenario.StartPrice)
	}

	// Create candle summary
	summary := &models.CandleSummary{
		Symbol: scenario.Symbol,
		Candles: map[string]*models.CandlesData{
			interval: {
				Candles:    candles,
				CreateTime: time.Now().UnixMilli(),
				UpdateTime: time.Now().UnixMilli(),
			},
		},
	}

	return summary, nil
}

// applyMarketEvent modifies candles to reflect a market event
func (g *MarketDataGenerator) applyMarketEvent(candles []*models.Candlestick, event MarketEvent, intervalDuration time.Duration, basePrice float64) {
	// Calculate which candles are affected by the event
	startCandleIndex := int(event.Time / intervalDuration)
	eventDurationCandles := int(event.Duration / intervalDuration)
	if eventDurationCandles == 0 {
		eventDurationCandles = 1
	}

	if startCandleIndex >= len(candles) {
		return
	}

	endCandleIndex := startCandleIndex + eventDurationCandles
	if endCandleIndex > len(candles) {
		endCandleIndex = len(candles)
	}

	// Apply the event effect
	for i := startCandleIndex; i < endCandleIndex; i++ {
		candle := candles[i]

		// Parse current prices
		open := g.parsePrice(candle.Open)
		close := g.parsePrice(candle.Close)
		high := g.parsePrice(candle.High)
		low := g.parsePrice(candle.Low)

		// Calculate event impact (stronger at the beginning, weaker at the end)
		progress := float64(i-startCandleIndex) / float64(eventDurationCandles)
		impact := g.calculateEventImpact(event, progress)

		// Apply impact based on event type
		switch event.Type {
		case "pump", "breakout", "squeeze_break":
			newClose := close * (1 + impact)
			newHigh := math.Max(high, newClose*1.01)
			candle.Close = g.formatPrice(newClose)
			candle.High = g.formatPrice(newHigh)

		case "dump", "flash_crash":
			newClose := close * (1 + impact) // impact is negative
			newLow := math.Min(low, newClose*0.99)
			candle.Close = g.formatPrice(newClose)
			candle.Low = g.formatPrice(newLow)

		case "bounce", "recovery":
			if impact > 0 {
				newClose := close * (1 + impact)
				newHigh := math.Max(high, newClose*1.005)
				candle.Close = g.formatPrice(newClose)
				candle.High = g.formatPrice(newHigh)
			}

		case "consolidation":
			// Reduce volatility
			range_ := high - low
			newRange := range_ * 0.3 // Tight range
			mid := (high + low) / 2
			candle.High = g.formatPrice(mid + newRange/2)
			candle.Low = g.formatPrice(mid - newRange/2)

		case "false_breakout", "false_breakdown":
			if i == startCandleIndex {
				// Initial false move
				if event.Type == "false_breakout" {
					newHigh := high * (1 + math.Abs(impact))
					candle.High = g.formatPrice(newHigh)
				} else {
					newLow := low * (1 - math.Abs(impact))
					candle.Low = g.formatPrice(newLow)
				}
			} else {
				// Rejection back
				rejectPrice := open * (1 - impact)
				candle.Close = g.formatPrice(rejectPrice)
			}
		}

		// Increase volume for significant events
		if math.Abs(impact) > 0.02 {
			currentVolume := g.parsePrice(candle.Volume)
			newVolume := currentVolume * (1 + math.Abs(impact)*5) // 5x volume multiplier
			candle.Volume = g.formatPrice(newVolume)
		}

		// Update next candle's open if this isn't the last candle
		if i+1 < len(candles) {
			candles[i+1].Open = candle.Close
		}
	}
}

// calculateEventImpact calculates the impact of an event based on its progress
func (g *MarketDataGenerator) calculateEventImpact(event MarketEvent, progress float64) float64 {
	switch event.Type {
	case "pump", "dump":
		// Gradual impact that builds up
		return event.Magnitude * progress

	case "flash_crash", "squeeze_break":
		// Sharp initial impact, then stabilizes
		if progress < 0.3 {
			return event.Magnitude * (progress / 0.3)
		}
		return event.Magnitude

	case "bounce", "recovery":
		// Quick initial recovery, then slows down
		return event.Magnitude * (1 - progress*0.5)

	case "consolidation":
		// Consistent low impact
		return event.Magnitude * 0.1

	case "false_breakout", "false_breakdown":
		// Sharp reversal pattern
		if progress < 0.5 {
			return event.Magnitude
		}
		return -event.Magnitude * 0.8

	default:
		return event.Magnitude * progress
	}
}

// GenerateRealisticDecisionInput creates decision input that matches the market scenario
func (g *MarketDataGenerator) GenerateRealisticDecisionInput(scenario *MarketScenario, currentPrice float64, timeProgress float64) *models.DecisionInput {
	input := g.GenerateDecisionInput(scenario.Symbol, currentPrice)

	// Adjust indicators based on scenario
	switch scenario.Trend {
	case "up":
		input.RSI = 55 + g.rand.Float64()*25 // 55-80 (bullish bias)
		input.Trend_1h = "UP"
		input.Trend_15m = "UP"
		input.TrendStrength = 60 + g.rand.Float64()*30 // 60-90

	case "down":
		input.RSI = 20 + g.rand.Float64()*25 // 20-45 (bearish bias)
		input.Trend_1h = "DOWN"
		input.Trend_15m = "DOWN"
		input.TrendStrength = 60 + g.rand.Float64()*30 // 60-90

	case "sideways":
		input.RSI = 40 + g.rand.Float64()*20 // 40-60 (neutral)
		input.Trend_1h = "SIDEWAYS"
		input.Trend_15m = "SIDEWAYS"
		input.TrendStrength = 20 + g.rand.Float64()*30 // 20-50 (weak trend)

	case "volatile":
		input.RSI = g.rand.Float64() * 100 // Any RSI
		trends := []string{"UP", "DOWN", "SIDEWAYS"}
		input.Trend_1h = trends[g.rand.Intn(len(trends))]
		input.Trend_15m = trends[g.rand.Intn(len(trends))]
		input.TrendStrength = 80 + g.rand.Float64()*20 // 80-100 (very strong but unclear direction)
	}

	// Adjust volatility-related metrics
	input.ATRPercent = scenario.Volatility

	// Set support/resistance based on recent price action
	priceChange := (currentPrice - scenario.StartPrice) / scenario.StartPrice
	if priceChange > 0.02 {
		input.ResistanceLevel = currentPrice * 1.02
		input.SupportLevel = scenario.StartPrice
		input.SwingHighBroken = true
	} else if priceChange < -0.02 {
		input.ResistanceLevel = scenario.StartPrice
		input.SupportLevel = currentPrice * 0.98
		input.SwingLowBroken = true
	} else {
		input.ResistanceLevel = scenario.StartPrice * 1.02
		input.SupportLevel = scenario.StartPrice * 0.98
	}

	return input
}

// Helper method to parse price from string (with error handling)
func (g *MarketDataGenerator) parsePrice(priceStr string) float64 {
	if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
		return price
	}
	return 0.0
}
