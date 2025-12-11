package decision

import (
	"math"
	"strings"

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

func scoreMarketStructure(in *DecisionInput) float64 {
	// Breakout / breakdown
	if in.Price >= in.RecentHigh*1.002 { // small buffer for breakout
		return 2
	}
	if in.Price <= in.RecentLow*0.998 { // small buffer for breakdown
		return -2
	}

	// Range; mild bias by proximity
	distHigh := (in.RecentHigh - in.Price) / in.RecentHigh
	distLow := (in.Price - in.RecentLow) / in.RecentLow
	if distLow < distHigh {
		return 0.5 // closer to support
	}
	return -0.5 // closer to resistance
}

func scoreVolumeOrderFlow(in *DecisionInput) float64 {
	// Futures vs Spot confirmation
	spotUp := in.SpotVolumeChange > 0
	futUp := in.FuturesVolumeChange > 0

	score := 0.0
	switch {
	case spotUp && futUp:
		score += 1.0
	case !spotUp && futUp:
		score -= 0.5 // potential fakeout
	case spotUp && !futUp:
		score += 0.3 // accumulation
	}

	// Open Interest
	if in.OIChange > 0 && in.PriceChangeSign() > 0 {
		score += 0.7
	} else if in.OIChange > 0 && in.PriceChangeSign() < 0 {
		score -= 0.7
	} else if in.OIChange < 0 {
		score -= 0.3
	}

	// Order book imbalance
	if in.OrderBookImbalance > 0.05 {
		score += 0.5
	} else if in.OrderBookImbalance < -0.05 {
		score -= 0.5
	}

	return clamp(score, -2, 2)
}

func scoreFundingLongShort(in *DecisionInput) float64 {
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

func scoreOnChain(in *DecisionInput) float64 {
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

func scoreMacroSentiment(in *DecisionInput) float64 {
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

func scoreQuant(in *DecisionInput) float64 {
	s := 0.0
	// RSI
	if in.RSI >= 70 {
		s -= 0.6
	} else if in.RSI <= 30 {
		s += 0.6
	} else if in.RSI >= 55 {
		s -= 0.2
	} else if in.RSI <= 45 {
		s += 0.2
	}

	// KDJ: K over D bullish, J extreme indicates exhaustion risk
	kd := in.K - in.D
	if kd > 5 {
		s += 0.4
	} else if kd < -5 {
		s -= 0.4
	}
	if in.J >= 100 {
		s -= 0.3
	} else if in.J <= 0 {
		s += 0.3
	}

	// Volatility clustering bias: use ATR% change if provided via recentHigh/Low proximity is already considered; we avoid double counting
	return clamp(s, -2, 2)
}

func scoreRisk(in *DecisionInput, bias string) float64 {
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

func (in *DecisionInput) PriceChangeSign() float64 {
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

func ComputeDecision(in *DecisionInput) *DecisionOutput {
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

	// Position sizing: max 5% of account
	positionSizePercent := 0.0
	if action != "Hold" {
		// scale with confidence between 60-100 -> 2% to 5%
		scaled := 2.0 + (float64(conf-60)/40.0)*3.0
		if scaled < 0 {
			scaled = 0
		}
		if scaled > 5 {
			scaled = 5
		}
		positionSizePercent = scaled
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

	out := &DecisionOutput{
		Prediction: prediction,
		Confidence: conf,
		Bias:       bias,
		TotalScore: weighted,
		CategoryScores: CategoryScores{
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
