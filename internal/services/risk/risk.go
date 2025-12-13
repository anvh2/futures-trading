package risk

import (
	"fmt"
	"math"
)

// Config defines tunable risk parameters. Defaults align to TRADE.md.
type Config struct {
	// Max percent of account equity allocated per trade (as margin).
	MaxPositionPercent float64
	// Stop-loss/Take-profit ATR multipliers
	StopATRMultiplier   float64
	TargetATRMultiplier float64
	// Minimum absolute buffers as fraction of price
	MinSLBuffer float64
	MinTPBuffer float64
}

func DefaultConfig() Config {
	return Config{
		MaxPositionPercent:  5.0,
		StopATRMultiplier:   1.5,
		TargetATRMultiplier: 2.5,
		MinSLBuffer:         0.002, // 0.2%
		MinTPBuffer:         0.004, // 0.4%
	}
}

// RecommendLeverage adjusts leverage based on ATR% and signal confidence.
func RecommendLeverage(atrPercent float64, confidence int) int {
	switch {
	case confidence >= 90 && atrPercent <= 1.2:
		return 10
	case confidence >= 85 && atrPercent <= 1.5:
		return 8
	case atrPercent >= 5.0:
		return 2
	case atrPercent >= 3.0:
		return 3
	default:
		return 4
	}
}

// PositionSizePercent scales within [0, MaxPositionPercent] based on confidence and ATR volatility.
// 60% -> ~2%, 100% -> MaxPositionPercent, adjusted down for high volatility.
func PositionSizePercent(confidence int, atrPercent float64, cfg Config) float64 {
	if confidence < 60 {
		return 0
	}

	// Base position size from confidence
	base := 2.0
	span := cfg.MaxPositionPercent - base
	pct := base + (float64(confidence-60)/40.0)*span

	// ATR volatility adjustment - reduce position size for high volatility
	var volatilityAdjustment float64
	switch {
	case atrPercent > 5.0: // Very high volatility
		volatilityAdjustment = 0.5
	case atrPercent > 3.0: // High volatility
		volatilityAdjustment = 0.7
	case atrPercent > 2.0: // Medium volatility
		volatilityAdjustment = 0.85
	case atrPercent > 1.0: // Normal volatility
		volatilityAdjustment = 1.0
	default: // Low volatility
		volatilityAdjustment = 1.1
	}

	pct *= volatilityAdjustment

	// Ensure bounds
	if pct < 0 {
		pct = 0
	}
	if pct > cfg.MaxPositionPercent {
		pct = cfg.MaxPositionPercent
	}
	return pct
}

// QuantityFromMargin computes contract quantity from allocated margin percent and leverage.
// qty = (capital * percent * leverage) / entryPrice
func QuantityFromMargin(capitalUSD, percent, entryPrice float64, leverage int) float64 {
	if entryPrice <= 0 || leverage <= 0 || capitalUSD <= 0 || percent <= 0 {
		return 0
	}
	margin := capitalUSD * (percent / 100.0)
	notional := margin * float64(leverage)
	return notional / entryPrice
}

// QuantityFromRisk computes quantity from risk-per-trade percent and stop distance.
// riskUSD = capital * riskPercent; qty = riskUSD / |entry - stop|
func QuantityFromRisk(capitalUSD, riskPercent, entryPrice, stopPrice float64) float64 {
	if entryPrice <= 0 || stopPrice <= 0 || capitalUSD <= 0 || riskPercent <= 0 {
		return 0
	}
	riskUSD := capitalUSD * (riskPercent / 100.0)
	dist := math.Abs(entryPrice - stopPrice)
	if dist <= 0 {
		return 0
	}
	return riskUSD / dist
}

// StopsTargets computes entry (pass-through), SL and TP using ATR% and bias.
// bias: "Bullish" or "Bearish". ATRPercent is in percent of price.
func StopsTargets(entryPrice, atrPercent float64, bias string, cfg Config) (entry, sl, tp float64) {
	entry = entryPrice
	atr := atrPercent / 100.0
	slDist := cfg.StopATRMultiplier * atr * entryPrice
	tpDist := cfg.TargetATRMultiplier * atr * entryPrice
	// enforce minimum buffers
	if slDist < cfg.MinSLBuffer*entryPrice {
		slDist = cfg.MinSLBuffer * entryPrice
	}
	if tpDist < cfg.MinTPBuffer*entryPrice {
		tpDist = cfg.MinTPBuffer * entryPrice
	}
	switch bias {
	case "Bullish":
		sl = entryPrice - slDist
		tp = entryPrice + tpDist
	case "Bearish":
		sl = entryPrice + slDist
		tp = entryPrice - tpDist
	default:
		sl = entryPrice
		tp = entryPrice
	}
	return
}

// ScaleInPlan returns a concise plan string given ATR context.
func ScaleInPlan(atrPercent float64) string {
	if atrPercent >= 4.0 {
		return "Enter 33%, add 33% at 0.5x ATR retrace, final 34% at 1.0x ATR; cancel on SL."
	}
	return "Enter 50% at signal, add 50% on 0.5x ATR pullback; cancel on SL."
}

// ScaleOutPlan returns a concise plan string given ATR context.
func ScaleOutPlan(atrPercent float64) string {
	if atrPercent >= 4.0 {
		return "Take 40% at 1.2R, 30% at 2.0R, trail stop to BE, exit rest at 3.0R or reversal."
	}
	return "Take 50% at 1.5R, trail to BE, exit rest at 2.5R or reversal."
}

// StringPlan returns a human-readable summary of the computed risk.
func StringPlan(symbol string, qty float64, entry, sl, tp float64, lev int) string {
	return fmt.Sprintf("%s qty=%.6f lev=%dx entry=%.6f SL=%.6f TP=%.6f", symbol, qty, lev, entry, sl, tp)
}
