package testdata

import (
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/anvh2/futures-trading/internal/models"
)

// MarketDataGenerator generates realistic market data for testing
type MarketDataGenerator struct {
	seed     int64
	rand     *rand.Rand
	baseTime time.Time
}

// NewMarketDataGenerator creates a new market data generator
func NewMarketDataGenerator(seed int64) *MarketDataGenerator {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	return &MarketDataGenerator{
		seed:     seed,
		rand:     rand.New(rand.NewSource(seed)),
		baseTime: time.Now().Add(-24 * time.Hour), // Start 24 hours ago
	}
}

// CandleGeneratorConfig holds configuration for candle generation
type CandleGeneratorConfig struct {
	Symbol     string
	Interval   string
	StartPrice float64
	Count      int
	Trend      string  // "up", "down", "sideways", "volatile"
	Volatility float64 // 0.01 = 1% volatility
	VolumeBase float64
	TickerLen  time.Duration
}

// GenerateCandles creates realistic candlestick data
func (g *MarketDataGenerator) GenerateCandles(config CandleGeneratorConfig) []*models.Candlestick {
	candles := make([]*models.Candlestick, config.Count)

	currentPrice := config.StartPrice
	currentTime := g.baseTime

	// Parse interval to duration
	intervalDuration := g.parseInterval(config.Interval)

	for i := 0; i < config.Count; i++ {
		// Generate OHLC for this candle
		open := currentPrice

		// Add trend bias
		trendBias := g.getTrendBias(config.Trend, config.Volatility)

		// Generate random price movements
		priceChange := g.rand.NormFloat64() * config.Volatility * currentPrice
		priceChange += trendBias * currentPrice

		close := open + priceChange

		// Generate high and low within reasonable bounds
		volatilityRange := config.Volatility * currentPrice * 0.5
		high := math.Max(open, close) + math.Abs(g.rand.NormFloat64())*volatilityRange
		low := math.Min(open, close) - math.Abs(g.rand.NormFloat64())*volatilityRange

		// Ensure high >= max(open, close) and low <= min(open, close)
		high = math.Max(high, math.Max(open, close))
		low = math.Min(low, math.Min(open, close))

		// Generate volume
		baseVolume := config.VolumeBase
		volumeVariation := g.rand.Float64()*0.5 + 0.75 // 75% to 125% of base
		volume := baseVolume * volumeVariation

		// Create candlestick
		candle := &models.Candlestick{
			OpenTime:  currentTime.UnixMilli(),
			CloseTime: currentTime.Add(intervalDuration).UnixMilli(),
			Open:      g.formatPrice(open),
			High:      g.formatPrice(high),
			Low:       g.formatPrice(low),
			Close:     g.formatPrice(close),
			Volume:    g.formatVolume(volume),
		}

		candles[i] = candle

		// Update for next iteration
		currentPrice = close
		currentTime = currentTime.Add(intervalDuration)
	}

	return candles
}

// GenerateCandleSummary creates a complete candle summary with multiple intervals
func (g *MarketDataGenerator) GenerateCandleSummary(symbol string, intervals []string) *models.CandleSummary {
	summary := &models.CandleSummary{
		Symbol:  symbol,
		Candles: make(map[string]*models.CandlesData),
	}

	basePrice := 50000.0 // Starting price (e.g., BTC)
	if symbol == "ETHUSDT" {
		basePrice = 3000.0
	} else if symbol == "ADAUSDT" {
		basePrice = 0.5
	}

	now := time.Now()

	for _, interval := range intervals {
		config := CandleGeneratorConfig{
			Symbol:     symbol,
			Interval:   interval,
			StartPrice: basePrice + g.rand.Float64()*basePrice*0.02 - basePrice*0.01, // ±1% variation
			Count:      100,                                                          // Last 100 candles
			Trend:      g.getRandomTrend(),
			Volatility: g.getVolatilityForInterval(interval),
			VolumeBase: g.getVolumeForSymbol(symbol),
			TickerLen:  g.parseInterval(interval),
		}

		candles := g.GenerateCandles(config)

		summary.Candles[interval] = &models.CandlesData{
			Candles:    candles,
			CreateTime: now.UnixMilli(),
			UpdateTime: now.UnixMilli(),
		}
	}

	return summary
}

// GenerateDecisionInput creates realistic decision input data
func (g *MarketDataGenerator) GenerateDecisionInput(symbol string, price float64) *models.DecisionInput {
	return &models.DecisionInput{
		Symbol:    symbol,
		Timeframe: "1h",
		Price:     price,

		// RSI values
		RSI:     g.rand.Float64() * 100,
		RSI_5m:  g.rand.Float64() * 100,
		RSI_15m: g.rand.Float64() * 100,
		RSI_1h:  g.rand.Float64() * 100,

		// Stochastic values
		K: g.rand.Float64() * 100,
		D: g.rand.Float64() * 100,
		J: g.rand.Float64() * 200,

		// Trend analysis
		Trend_5m:  g.getRandomTrend(),
		Trend_15m: g.getRandomTrend(),
		Trend_1h:  g.getRandomTrend(),

		// Price levels
		RecentHigh:      price * (1 + g.rand.Float64()*0.05),
		RecentLow:       price * (1 - g.rand.Float64()*0.05),
		SupportLevel:    price * (1 - g.rand.Float64()*0.02),
		ResistanceLevel: price * (1 + g.rand.Float64()*0.02),

		// Market structure
		TrendStrength:   g.rand.Float64() * 100,
		SwingHighBroken: g.rand.Float64() > 0.7,
		SwingLowBroken:  g.rand.Float64() > 0.7,

		// Market data
		OIChange:            (g.rand.Float64() - 0.5) * 0.2,   // ±10%
		FundingRate:         (g.rand.Float64() - 0.5) * 0.001, // ±0.05%
		LongShortRatio:      0.8 + g.rand.Float64()*0.4,       // 0.8-1.2
		SpotVolumeChange:    (g.rand.Float64() - 0.5) * 0.4,   // ±20%
		FuturesVolumeChange: (g.rand.Float64() - 0.5) * 0.4,
		OrderBookImbalance:  (g.rand.Float64() - 0.5) * 0.2,

		// Volume analysis
		VWAP:               price * (1 + (g.rand.Float64()-0.5)*0.01),
		RelativeVolume:     0.5 + g.rand.Float64()*1.5, // 0.5-2.0
		VolumeRatio:        0.8 + g.rand.Float64()*0.4, // 0.8-1.2
		VolumeAtPriceHigh:  g.rand.Float64() > 0.6,
		ExchangeInflows:    (g.rand.Float64() - 0.5) * 1000000, // ±500k
		WhaleTxCount:       int(g.rand.Float64() * 20),
		SpotFuturesPremium: (g.rand.Float64() - 0.5) * 0.01,

		// Sentiment
		MacroSentimentScore: (g.rand.Float64() - 0.5) * 2, // -1 to 1
		NewsSentimentScore:  (g.rand.Float64() - 0.5) * 2, // -1 to 1
		FearGreedIndex:      int(g.rand.Float64() * 100),  // 0-100

		// Risk metrics
		ATRPercent:      g.rand.Float64() * 0.05, // 0-5%
		Capital:         10000.0,                 // $10k
		CurrentPosition: g.getRandomPosition(),
	}
}

// GenerateSignal creates a realistic trading signal
func (g *MarketDataGenerator) GenerateSignal(symbol string, price float64) *models.Signal {
	action := g.getRandomSignalAction()
	signalType := models.SignalTypeEntry
	if action == models.SignalActionClose {
		signalType = models.SignalTypeExit
	}

	strength := g.rand.Float64()
	confidence := 0.5 + g.rand.Float64()*0.5 // 50-100% confidence

	signal := &models.Signal{
		Symbol:     symbol,
		Type:       signalType,
		Action:     action,
		Strength:   strength,
		Confidence: confidence,
		Price:      price,
		Interval:   "1h",
		Strategy:   "test_strategy",
		Indicators: make(map[string]float64),
		Metadata:   make(map[string]interface{}),
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(1 * time.Hour),
		IsActive:   true,
	}

	// Add some indicator values
	signal.Indicators["rsi"] = g.rand.Float64() * 100
	signal.Indicators["macd"] = (g.rand.Float64() - 0.5) * 10
	signal.Indicators["ema"] = price * (1 + (g.rand.Float64()-0.5)*0.01)

	// Set stop loss and take profit for entry signals
	if signalType == models.SignalTypeEntry {
		if action == models.SignalActionBuy {
			signal.StopLoss = price * 0.98   // 2% stop loss
			signal.TakeProfit = price * 1.04 // 4% take profit
		} else if action == models.SignalActionSell {
			signal.StopLoss = price * 1.02   // 2% stop loss
			signal.TakeProfit = price * 0.96 // 4% take profit
		}
	}

	return signal
}

// Helper methods

func (g *MarketDataGenerator) parseInterval(interval string) time.Duration {
	switch interval {
	case "1m":
		return time.Minute
	case "5m":
		return 5 * time.Minute
	case "15m":
		return 15 * time.Minute
	case "1h":
		return time.Hour
	case "4h":
		return 4 * time.Hour
	case "1d":
		return 24 * time.Hour
	default:
		return time.Hour
	}
}

func (g *MarketDataGenerator) getTrendBias(trend string, volatility float64) float64 {
	switch trend {
	case "up":
		return volatility * 0.3
	case "down":
		return -volatility * 0.3
	case "volatile":
		if g.rand.Float64() > 0.5 {
			return volatility * 0.5
		}
		return -volatility * 0.5
	case "sideways":
		fallthrough
	default:
		return (g.rand.Float64() - 0.5) * volatility * 0.1
	}
}

func (g *MarketDataGenerator) getRandomTrend() string {
	trends := []string{"UP", "DOWN", "SIDEWAYS"}
	return trends[g.rand.Intn(len(trends))]
}

func (g *MarketDataGenerator) getVolatilityForInterval(interval string) float64 {
	switch interval {
	case "1m":
		return 0.005 // 0.5%
	case "5m":
		return 0.01 // 1%
	case "15m":
		return 0.015 // 1.5%
	case "1h":
		return 0.02 // 2%
	case "4h":
		return 0.03 // 3%
	case "1d":
		return 0.05 // 5%
	default:
		return 0.02
	}
}

func (g *MarketDataGenerator) getVolumeForSymbol(symbol string) float64 {
	switch symbol {
	case "BTCUSDT":
		return 1000.0
	case "ETHUSDT":
		return 5000.0
	case "ADAUSDT":
		return 100000.0
	default:
		return 10000.0
	}
}

func (g *MarketDataGenerator) getRandomPosition() string {
	positions := []string{"NONE", "LONG", "SHORT"}
	return positions[g.rand.Intn(len(positions))]
}

func (g *MarketDataGenerator) getRandomSignalAction() models.SignalAction {
	actions := []models.SignalAction{
		models.SignalActionBuy,
		models.SignalActionSell,
		models.SignalActionHold,
		models.SignalActionClose,
	}
	return actions[g.rand.Intn(len(actions))]
}

func (g *MarketDataGenerator) formatPrice(price float64) string {
	return strconv.FormatFloat(price, 'f', 2, 64)
}

func (g *MarketDataGenerator) formatVolume(volume float64) string {
	return strconv.FormatFloat(volume, 'f', 4, 64)
}
