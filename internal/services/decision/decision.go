package decision

import (
	"time"

	"github.com/anvh2/futures-trading/internal/models"
	"github.com/anvh2/futures-trading/internal/services/state"
	"go.uber.org/zap"
)

// MakeDecision generates trading decisions based on current analysis using the comprehensive scoring engine
func (de *Maker) MakeDecision(signal *models.Signal) *models.TradingDecision {
	de.logger.Debug("Making trading decisions using comprehensive scoring engine...")

	// Check if system is in a state to make decisions
	systemStatus := de.state.GetSystemStatus()
	if systemStatus != state.SystemStatusActive {
		de.logger.Debug("System not active, skipping decision")
		return nil
	}

	// Convert signal to decision input for the scoring engine
	decisionInput := de.convertSignalToDecisionInput(signal)
	if decisionInput == nil {
		de.logger.Warn("Failed to convert signal to decision input")
		return nil
	}

	// Get current positions to avoid over-positioning
	positions := de.state.GetPositions()
	if _, exists := positions[signal.Symbol]; exists {
		de.logger.Debug("Position already exists for symbol", zap.String("symbol", signal.Symbol))
		decisionInput.CurrentPosition = "Long" // Assume existing position is long for now
	}

	// Use the comprehensive scoring engine to compute decision
	engineOutput := ComputeDecision(decisionInput)

	// Convert engine output to trading decision
	return de.convertEngineOutputToTradingDecision(signal, engineOutput)
}

// convertSignalToDecisionInput converts a trading signal to decision input for the scoring engine
func (de *Maker) convertSignalToDecisionInput(signal *models.Signal) *models.DecisionInput {
	if !signal.IsValid() {
		return nil
	}

	// Extract technical indicators from signal
	rsi, _ := signal.GetIndicatorValue("rsi")
	k, _ := signal.GetIndicatorValue("k")
	d, _ := signal.GetIndicatorValue("d")
	j, _ := signal.GetIndicatorValue("j")
	atr, _ := signal.GetIndicatorValue("atr_percent")
	vwap, _ := signal.GetIndicatorValue("vwap")
	recentHigh, _ := signal.GetIndicatorValue("recent_high")
	recentLow, _ := signal.GetIndicatorValue("recent_low")

	// Extract volume data
	relativeVolume, _ := signal.GetIndicatorValue("relative_volume")
	volumeRatio, _ := signal.GetIndicatorValue("volume_ratio")
	spotVolumeChange, _ := signal.GetIndicatorValue("spot_volume_change")
	futuresVolumeChange, _ := signal.GetIndicatorValue("futures_volume_change")

	// Extract funding and positioning data
	fundingRate, _ := signal.GetIndicatorValue("funding_rate")
	longShortRatio, _ := signal.GetIndicatorValue("long_short_ratio")
	oiChange, _ := signal.GetIndicatorValue("oi_change")

	// Extract market structure indicators
	trendStrength, _ := signal.GetIndicatorValue("trend_strength")
	supportLevel, _ := signal.GetIndicatorValue("support_level")
	resistanceLevel, _ := signal.GetIndicatorValue("resistance_level")

	// Extract multi-timeframe data
	rsi5m, _ := signal.GetIndicatorValue("rsi_5m")
	rsi15m, _ := signal.GetIndicatorValue("rsi_15m")
	rsi1h, _ := signal.GetIndicatorValue("rsi_1h")

	// Extract trend data from metadata
	trend5m, _ := signal.GetMetadata("trend_5m")
	trend15m, _ := signal.GetMetadata("trend_15m")
	trend1h, _ := signal.GetMetadata("trend_1h")

	// Extract swing break data
	swingHighBroken, _ := signal.GetMetadata("swing_high_broken")
	swingLowBroken, _ := signal.GetMetadata("swing_low_broken")

	// Extract on-chain and macro data
	exchangeInflows, _ := signal.GetIndicatorValue("exchange_inflows")
	macroSentiment, _ := signal.GetIndicatorValue("macro_sentiment")
	newsSentiment, _ := signal.GetIndicatorValue("news_sentiment")
	fearGreed, _ := signal.GetIndicatorValue("fear_greed_index")

	// Convert string to bool safely
	toSafeBool := func(v interface{}) bool {
		if b, ok := v.(bool); ok {
			return b
		}
		return false
	}

	// Convert string safely
	toSafeString := func(v interface{}) string {
		if s, ok := v.(string); ok {
			return s
		}
		return ""
	}

	return &models.DecisionInput{
		Symbol:              signal.Symbol,
		Timeframe:           signal.Interval,
		RSI:                 rsi,
		K:                   k,
		D:                   d,
		J:                   j,
		RSI_5m:              rsi5m,
		RSI_15m:             rsi15m,
		RSI_1h:              rsi1h,
		Trend_5m:            toSafeString(trend5m),
		Trend_15m:           toSafeString(trend15m),
		Trend_1h:            toSafeString(trend1h),
		Price:               signal.Price,
		RecentHigh:          recentHigh,
		RecentLow:           recentLow,
		TrendStrength:       trendStrength,
		SwingHighBroken:     toSafeBool(swingHighBroken),
		SwingLowBroken:      toSafeBool(swingLowBroken),
		SupportLevel:        supportLevel,
		ResistanceLevel:     resistanceLevel,
		OIChange:            oiChange,
		FundingRate:         fundingRate,
		LongShortRatio:      longShortRatio,
		SpotVolumeChange:    spotVolumeChange,
		FuturesVolumeChange: futuresVolumeChange,
		VWAP:                vwap,
		RelativeVolume:      relativeVolume,
		VolumeRatio:         volumeRatio,
		ExchangeInflows:     exchangeInflows,
		MacroSentimentScore: macroSentiment,
		NewsSentimentScore:  newsSentiment,
		FearGreedIndex:      int(fearGreed),
		ATRPercent:          atr,
		Capital:             100000, // Default capital, should be configurable
		CurrentPosition:     "",     // Will be set based on current positions
	}
}

// convertEngineOutputToTradingDecision converts engine output to a trading decision
func (de *Maker) convertEngineOutputToTradingDecision(signal *models.Signal, engineOutput *models.DecisionOutput) *models.TradingDecision {
	if engineOutput == nil {
		return nil
	}

	// Map engine action to trading action
	var action string
	switch engineOutput.Action {
	case "Long":
		action = "BUY"
	case "Short":
		action = "SELL"
	case "Hold":
		action = "HOLD"
	default:
		action = "HOLD"
	}

	// Calculate actual position size from percentage and capital
	capital := 100000.0 // Should be configurable or retrieved from account
	positionSize := (engineOutput.PositionSizePercent / 100.0) * capital

	decision := &models.TradingDecision{
		Symbol:     signal.Symbol,
		Action:     action,
		Size:       positionSize,
		Price:      engineOutput.EntryPrice,
		Confidence: float64(engineOutput.Confidence) / 100.0, // Convert to 0-1 range
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"strategy":          signal.Strategy,
			"interval":          signal.Interval,
			"prediction":        engineOutput.Prediction,
			"bias":              engineOutput.Bias,
			"total_score":       engineOutput.TotalScore,
			"category_scores":   engineOutput.CategoryScores,
			"position_size_pct": engineOutput.PositionSizePercent,
			"leverage":          engineOutput.Leverage,
			"stop_loss":         engineOutput.StopLoss,
			"take_profit":       engineOutput.TakeProfit,
			"scale_in_plan":     engineOutput.ScaleInPlan,
			"scale_out_plan":    engineOutput.ScaleOutPlan,
			"reasoning":         engineOutput.Reasoning,
			"engine_confidence": engineOutput.Confidence,
		},
	}

	return decision
}

// GetEngineDecision provides direct access to the engine decision output for advanced use cases
func (de *Maker) GetEngineDecision(signal *models.Signal) *models.DecisionOutput {
	decisionInput := de.convertSignalToDecisionInput(signal)
	if decisionInput == nil {
		return nil
	}

	// Get current positions
	positions := de.state.GetPositions()
	if _, exists := positions[signal.Symbol]; exists {
		decisionInput.CurrentPosition = "Long" // Assume existing position is long
	}

	return ComputeDecision(decisionInput)
}
