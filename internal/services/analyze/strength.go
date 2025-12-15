package analyzer

import (
	"math"

	"github.com/anvh2/futures-trading/internal/models"
)

// calculateConfidence calculates signal confidence based on multiple indicators with realistic scaling
func (s *Analyzer) calculateConfidence(oscillator *models.Oscillator) float64 {
	tradingStoch := oscillator.Stoch[s.settings.TradingInterval]
	if tradingStoch == nil {
		return 0.0
	}

	confidence := 0.0

	// RSI confidence (base signal strength)
	rsi := tradingStoch.RSI
	if rsi >= 80 || rsi <= 20 {
		confidence += 0.5 // Strong extreme levels
	} else if rsi >= 70 || rsi <= 30 {
		confidence += 0.3 // Moderate extreme levels
	}

	// Stochastic %K and %D alignment (max 0.2)
	k := tradingStoch.K
	d := tradingStoch.D
	if (k >= 80 && d >= 80) || (k <= 20 && d <= 20) {
		confidence += 0.2 // Both in same extreme zone
	}

	// K and D crossover signals (max 0.15)
	if k > d && rsi <= 30 {
		confidence += 0.15 // Bullish crossover in oversold
	} else if k < d && rsi >= 70 {
		confidence += 0.15 // Bearish crossover in overbought
	}

	// Multi-timeframe confluence (max 0.15)
	confluenceBonus := s.calculateTimeframeConfluence(oscillator, rsi >= 70)
	confidence += confluenceBonus * 0.15

	// Cap confidence at realistic levels
	return math.Min(confidence, 0.85) // Max 85% confidence
}

// calculateStrength calculates signal strength based on RSI extremes and distance from thresholds
func (s *Analyzer) calculateStrength(oscillator *models.Oscillator) float64 {
	tradingStoch := oscillator.Stoch[s.settings.TradingInterval]
	if tradingStoch == nil {
		return 0.0
	}

	rsi := tradingStoch.RSI
	if rsi >= 70 {
		// Stronger signal the more overbought (70-100 range)
		return math.Min((rsi-70)/20, 1.0) // Scale 70-90+ to 0-1
	} else if rsi <= 30 {
		// Stronger signal the more oversold (0-30 range)
		return math.Min((30-rsi)/20, 1.0) // Scale 10-30 to 0-1
	}
	return 0.0 // No signal in neutral zone
}

// calculateTimeframeConfluence checks if other timeframes support the signal
func (s *Analyzer) calculateTimeframeConfluence(oscillator *models.Oscillator, isBearish bool) float64 {
	supportingTimeframes := 0
	totalTimeframes := 0

	for interval, stoch := range oscillator.Stoch {
		if interval == s.settings.TradingInterval {
			continue
		}

		totalTimeframes++

		if isBearish && stoch.RSI >= 60 { // Bearish confluence
			supportingTimeframes++
		} else if !isBearish && stoch.RSI <= 40 { // Bullish confluence
			supportingTimeframes++
		}
	}

	if totalTimeframes == 0 {
		return 0.0
	}

	return float64(supportingTimeframes) / float64(totalTimeframes)
}
