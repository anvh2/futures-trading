package decision

import (
	"math"
	"strings"

	"github.com/anvh2/futures-trading/internal/models"
	"github.com/anvh2/futures-trading/internal/services/risk"
)

// Weights per the specification
const (
	weightMarketStructure  = 0.25
	weightVolumeOrderFlow  = 0.20
	weightFundingLongShort = 0.15
	weightOnChain          = 0.10
	weightMacroSentiment   = 0.10
	weightQuantModels      = 0.15
	weightRiskManagement   = 0.05

	longThreshold  = 0.8
	shortThreshold = -0.8
)

// clamp limits v to [min, max]
func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// sign returns -1, 0, 1
func sign(v float64) float64 {
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	default:
		return 0
	}
}

// Score computation helpers return value in [-2, 2]

func scoreMarketStructure(in *models.DecisionInput) float64 {
	score := 0.0

	// 1. Breakout/Breakdown analysis with enhanced precision
	breakoutThreshold := 0.002 // 0.2% buffer
	if in.Price >= in.RecentHigh*(1+breakoutThreshold) {
		score += 2.0 // Strong bullish breakout

		// Amplify if swing high was broken
		if in.SwingHighBroken {
			score += 0.3
		}
	} else if in.Price <= in.RecentLow*(1-breakoutThreshold) {
		score -= 2.0 // Strong bearish breakdown

		// Amplify if swing low was broken
		if in.SwingLowBroken {
			score -= 0.3
		}
	}

	// 2. Support/Resistance proximity analysis
	if in.SupportLevel > 0 && in.ResistanceLevel > 0 {
		supportDist := math.Abs(in.Price-in.SupportLevel) / in.Price
		resistanceDist := math.Abs(in.Price-in.ResistanceLevel) / in.Price

		// Close to support = potential bounce (bullish)
		if supportDist < 0.01 { // Within 1%
			score += 0.8
		} else if supportDist < 0.02 { // Within 2%
			score += 0.4
		}

		// Close to resistance = potential rejection (bearish)
		if resistanceDist < 0.01 { // Within 1%
			score -= 0.8
		} else if resistanceDist < 0.02 { // Within 2%
			score -= 0.4
		}

		// Range positioning bias
		rangeSize := in.ResistanceLevel - in.SupportLevel
		if rangeSize > 0 {
			priceInRange := (in.Price - in.SupportLevel) / rangeSize
			if priceInRange < 0.3 {
				score += 0.3 // In lower third of range = bullish bias
			} else if priceInRange > 0.7 {
				score -= 0.3 // In upper third of range = bearish bias
			}
		}
	} else {
		// Fallback to simple high/low analysis
		distHigh := (in.RecentHigh - in.Price) / in.RecentHigh
		distLow := (in.Price - in.RecentLow) / in.RecentLow
		if distLow < distHigh {
			score += 0.5 // Closer to support
		} else {
			score -= 0.5 // Closer to resistance
		}
	}

	// 3. Trend strength consideration
	if in.TrendStrength > 0 {
		if in.TrendStrength > 50 { // Strong trend
			// In strong trends, favor continuation over reversal
			trendBias := sign(in.Price - (in.RecentHigh+in.RecentLow)/2)
			score += trendBias * 0.4 * (in.TrendStrength / 100.0)
		} else if in.TrendStrength < 25 { // Weak trend/ranging
			// In ranging markets, favor mean reversion
			rangeMid := (in.RecentHigh + in.RecentLow) / 2
			if in.Price > rangeMid {
				score -= 0.3 // Above mid = bearish bias in range
			} else {
				score += 0.3 // Below mid = bullish bias in range
			}
		}
	}

	return clamp(score, -2, 2)
}

func scoreVolumeOrderFlow(in *models.DecisionInput) float64 {
	score := 0.0

	// 1. Futures vs Spot confirmation
	spotUp := in.SpotVolumeChange > 0
	futUp := in.FuturesVolumeChange > 0

	switch {
	case spotUp && futUp:
		score += 1.0 // Strong confirmation
	case !spotUp && futUp:
		score -= 0.5 // Potential fakeout
	case spotUp && !futUp:
		score += 0.3 // Accumulation pattern
	}

	// 2. Open Interest confirmation
	if in.OIChange > 0 && PriceChangeSign(in) > 0 {
		score += 0.7 // Rising price + rising OI = bullish
	} else if in.OIChange > 0 && PriceChangeSign(in) < 0 {
		score -= 0.7 // Falling price + rising OI = bearish
	} else if in.OIChange < 0 {
		score -= 0.3 // Trend exhaustion/profit taking
	}

	// 3. Order book imbalance
	if in.OrderBookImbalance > 0.05 {
		score += 0.5 // Buy pressure
	} else if in.OrderBookImbalance < -0.05 {
		score -= 0.5 // Sell pressure
	}

	// 4. Enhanced volume analysis
	if in.RelativeVolume > 0 {
		// High relative volume strengthens signals
		if in.RelativeVolume > 2.0 {
			if score > 0 {
				score += 0.4 // Amplify bullish signals with high volume
			} else {
				score -= 0.4 // Amplify bearish signals with high volume
			}
		} else if in.RelativeVolume > 1.5 {
			if score > 0 {
				score += 0.2
			} else {
				score -= 0.2
			}
		}

		// Low volume weakens conviction
		if in.RelativeVolume < 0.5 {
			score *= 0.6 // Reduce signal strength on low volume
		}
	}

	// 5. VWAP analysis
	if in.VWAP > 0 && in.Price > 0 {
		vwapDeviation := (in.Price - in.VWAP) / in.VWAP
		if math.Abs(vwapDeviation) > 0.02 { // 2% deviation threshold
			if vwapDeviation > 0 {
				score += 0.3 // Price well above VWAP = bullish
			} else {
				score -= 0.3 // Price well below VWAP = bearish
			}
		}
	}

	// 6. Volume ratio (buy vs sell estimation)
	if in.VolumeRatio != 0 {
		if in.VolumeRatio > 1.2 {
			score += 0.4 // Buying pressure
		} else if in.VolumeRatio < 0.8 {
			score -= 0.4 // Selling pressure
		}
	}

	// 7. Volume at price level analysis
	if in.VolumeAtPriceHigh {
		// High volume at current price suggests strong support/resistance
		score += 0.2 * sign(score) // Amplify existing bias
	}

	return clamp(score, -2, 2)
}

func scoreFundingLongShort(in *models.DecisionInput) float64 {
	// Funding rate extremes
	s := 0.0
	if in.FundingRate > 0.02 { // highly positive
		s -= 0.8
	} else if in.FundingRate < -0.02 {
		s += 0.8
	} else {
		s += 0.2 * sign(in.FundingRate)
	}

	// Long/Short ratio
	if in.LongShortRatio > 1.6 {
		s -= 0.7
	} else if in.LongShortRatio < 0.6 {
		s += 0.7
	} else if in.LongShortRatio > 1.1 {
		s -= 0.2
	} else if in.LongShortRatio < 0.9 {
		s += 0.2
	}

	return clamp(s, -2, 2)
}

func scoreOnChain(in *models.DecisionInput) float64 {
	s := 0.0
	// Exchange inflows/outflows (positive -> inflow -> sell pressure)
	if in.ExchangeInflows > 0 {
		s -= 0.5
	} else if in.ExchangeInflows < 0 {
		s += 0.5
	}

	// Whale activity
	if in.WhaleTxCount >= 10 {
		s += 0.3
	}

	// Spot-futures premium
	if in.SpotFuturesPremium > 1.0 { // overheated
		s -= 0.5
	} else if in.SpotFuturesPremium < -1.0 { // backwardation
		s += 0.5
	}
	return clamp(s, -2, 2)
}

func scoreMacroSentiment(in *models.DecisionInput) float64 {
	// Normalize macro + news + fear/greed into [-2,2]
	s := 0.0
	s += 1.0 * in.MacroSentimentScore // assume already in [-1,1]
	s += 0.7 * in.NewsSentimentScore  // assume already in [-1,1]

	// Fear & Greed Index: extreme greedy (>85) bearish; extreme fear (<15) bullish
	if in.FearGreedIndex >= 85 {
		s -= 0.7
	} else if in.FearGreedIndex <= 15 {
		s += 0.7
	}
	return clamp(s, -2, 2)
}

func scoreQuant(in *models.DecisionInput) float64 {
	score := 0.0

	// 1. Primary timeframe RSI
	if in.RSI >= 70 {
		score -= 0.6 // Overbought
	} else if in.RSI <= 30 {
		score += 0.6 // Oversold
	} else if in.RSI >= 55 {
		score -= 0.2 // Mild overbought
	} else if in.RSI <= 45 {
		score += 0.2 // Mild oversold
	}

	// 2. Multi-timeframe RSI confluence
	confluenceScore := 0.0
	confluenceCount := 0

	// Analyze 5m, 15m, 1h RSI if available
	timeframes := []struct {
		rsi    float64
		weight float64
	}{
		{in.RSI_5m, 0.3},  // Lower timeframe, lower weight
		{in.RSI_15m, 0.5}, // Medium timeframe
		{in.RSI_1h, 0.7},  // Higher timeframe, higher weight
	}

	for _, tf := range timeframes {
		if tf.rsi > 0 { // Valid RSI value
			confluenceCount++
			if tf.rsi >= 70 {
				confluenceScore -= 0.5 * tf.weight
			} else if tf.rsi <= 30 {
				confluenceScore += 0.5 * tf.weight
			} else if tf.rsi >= 60 {
				confluenceScore -= 0.2 * tf.weight
			} else if tf.rsi <= 40 {
				confluenceScore += 0.2 * tf.weight
			}
		}
	}

	// Apply confluence if we have multiple timeframes
	if confluenceCount > 1 {
		score += confluenceScore / float64(confluenceCount)
	}

	// 3. Multi-timeframe trend alignment
	trendAlignment := 0.0
	trendCount := 0
	trends := []string{in.Trend_5m, in.Trend_15m, in.Trend_1h}
	weights := []float64{0.3, 0.5, 0.7}

	for i, trend := range trends {
		if trend != "" {
			trendCount++
			switch trend {
			case "UP":
				trendAlignment += weights[i]
			case "DOWN":
				trendAlignment -= weights[i]
				// SIDEWAYS contributes 0
			}
		}
	}

	if trendCount > 1 {
		trendAlignment = trendAlignment / float64(trendCount)
		score += trendAlignment * 0.4 // Trend alignment bonus
	}

	// 4. KDJ analysis (unchanged)
	kd := in.K - in.D
	if kd > 5 {
		score += 0.4 // K over D bullish
	} else if kd < -5 {
		score -= 0.4 // K under D bearish
	}

	// J extreme readings
	if in.J >= 100 {
		score -= 0.3 // Extreme overbought
	} else if in.J <= 0 {
		score += 0.3 // Extreme oversold
	}

	// 5. Divergence detection bonus (placeholder for future enhancement)
	// TODO: Implement price vs RSI divergence detection

	return clamp(score, -2, 2)
}

func scoreRisk(in *models.DecisionInput, bias string) float64 {
	s := 0.0
	// ATR-based tradability: very high ATR% reduces score
	if in.ATRPercent >= 6 {
		s -= 0.8
	} else if in.ATRPercent >= 3.5 {
		s -= 0.4
	} else if in.ATRPercent <= 1.0 {
		s += 0.2
	}

	// Current position conflict penalization
	if (bias == "Bullish" && strings.EqualFold(in.CurrentPosition, "Short")) ||
		(bias == "Bearish" && strings.EqualFold(in.CurrentPosition, "Long")) {
		s -= 0.5
	}

	return clamp(s, -2, 2)
}

func deriveBias(total float64) string {
	switch {
	case total >= longThreshold:
		return "Bullish"
	case total <= shortThreshold:
		return "Bearish"
	default:
		return "Neutral"
	}
}

func PriceChangeSign(in *models.DecisionInput) float64 {
	// Heuristic: price vs mid of recent range
	mid := (in.RecentHigh + in.RecentLow) / 2
	return sign(in.Price - mid)
}

// leverage and stops/targets are handled by risk module

func confidenceFromScores(weightedTotal float64, cats []float64) int {
	// Base on magnitude and agreement across categories
	mag := math.Abs(weightedTotal)
	base := 50.0 + mag*30.0 // up to ~50 + (2*30)=110 capped later

	agree := 0
	for _, c := range cats {
		if c > 0 {
			agree++
		}
	}
	agreementBoost := float64(agree) / float64(len(cats)) * 20.0 // up to +20
	conf := int(clamp(base+agreementBoost, 0, 100))
	if conf < 0 {
		conf = 0
	}
	if conf > 100 {
		conf = 100
	}
	return conf
}

func ComputeDecision(in *models.DecisionInput) *models.DecisionOutput {
	// Raw category scores in [-2,2]
	sMarket := scoreMarketStructure(in)
	sVol := scoreVolumeOrderFlow(in)
	sFunding := scoreFundingLongShort(in)
	sOnChain := scoreOnChain(in)
	sMacro := scoreMacroSentiment(in)
	sQuant := scoreQuant(in)
	// risk uses bias later; temporarily assume neutral
	tmpRisk := scoreRisk(in, "Neutral")

	weighted := sMarket*weightMarketStructure + sVol*weightVolumeOrderFlow + sFunding*weightFundingLongShort + sOnChain*weightOnChain + sMacro*weightMacroSentiment + sQuant*weightQuantModels + tmpRisk*weightRiskManagement

	bias := deriveBias(weighted)
	// Recompute risk with derived bias for better alignment
	sRisk := scoreRisk(in, bias)
	weighted = sMarket*weightMarketStructure + sVol*weightVolumeOrderFlow + sFunding*weightFundingLongShort + sOnChain*weightOnChain + sMacro*weightMacroSentiment + sQuant*weightQuantModels + sRisk*weightRiskManagement

	cats := []float64{sMarket * weightMarketStructure, sVol * weightVolumeOrderFlow, sFunding * weightFundingLongShort, sOnChain * weightOnChain, sMacro * weightMacroSentiment, sQuant * weightQuantModels, sRisk * weightRiskManagement}
	conf := confidenceFromScores(weighted, cats)

	prediction := "Sideways"
	action := "Hold"
	if weighted >= longThreshold {
		prediction = "Bump"
		action = "Long"
	} else if weighted <= shortThreshold {
		prediction = "Dump"
		action = "Short"
	}

	// Enforce rules: Never open trade with confidence < 60%
	if conf < 60 {
		action = "Hold"
	}

	cfg := risk.DefaultConfig()
	entry, sl, tp := risk.StopsTargets(in.Price, in.ATRPercent, bias, cfg)

	// Enhanced position sizing with ATR volatility adjustment
	positionSizePercent := 0.0
	if action != "Hold" {
		positionSizePercent = risk.PositionSizePercent(conf, in.ATRPercent, cfg)
	}

	leverage := 0
	if action != "Hold" {
		leverage = risk.RecommendLeverage(in.ATRPercent, conf)
	}

	// Reasoning: top 3 strongest signals
	reasons := []string{}
	if sMarket >= 1.5 || sMarket <= -1.5 {
		if sMarket > 0 {
			reasons = append(reasons, "Breakout/Support proximity bullish")
		} else {
			reasons = append(reasons, "Breakdown/Resistance proximity bearish")
		}
	}
	if math.Abs(sVol) >= 1.0 {
		if sVol > 0 {
			reasons = append(reasons, "Spot+Futures volume + OI confirm up")
		} else {
			reasons = append(reasons, "OI/Orderbook confirm down")
		}
	}
	if math.Abs(sQuant) >= 0.8 {
		if sQuant > 0 {
			reasons = append(reasons, "RSI/KDJ favor longs")
		} else {
			reasons = append(reasons, "RSI/KDJ favor shorts")
		}
	}
	if len(reasons) > 3 {
		reasons = reasons[:3]
	}

	out := &models.DecisionOutput{
		Prediction: prediction,
		Confidence: conf,
		Bias:       bias,
		TotalScore: weighted,
		CategoryScores: models.CategoryScores{
			MarketStructure:  sMarket * weightMarketStructure,
			VolumeOrderFlow:  sVol * weightVolumeOrderFlow,
			FundingLongShort: sFunding * weightFundingLongShort,
			OnChain:          sOnChain * weightOnChain,
			MacroSentiment:   sMacro * weightMacroSentiment,
			QuantModels:      sQuant * weightQuantModels,
			RiskManagement:   sRisk * weightRiskManagement,
		},
		Action:              action,
		PositionSizePercent: positionSizePercent,
		Leverage:            leverage,
		EntryPrice:          entry,
		StopLoss:            sl,
		TakeProfit:          tp,
		ScaleInPlan:         risk.ScaleInPlan(in.ATRPercent),
		ScaleOutPlan:        risk.ScaleOutPlan(in.ATRPercent),
		Reasoning:           strings.Join(reasons, "; "),
	}

	// Align action with bias majority if conflicting
	if out.Bias == "Neutral" {
		out.Action = "Hold"
	}
	return out
}
