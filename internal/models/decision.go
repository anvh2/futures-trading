package models

// DecisionInput carries all metrics required by the framework
type DecisionInput struct {
	Symbol    string  `json:"symbol"`
	Timeframe string  `json:"timeframe"`
	RSI       float64 `json:"rsi"`
	K         float64 `json:"k"`
	D         float64 `json:"d"`
	J         float64 `json:"j"`
	// Multi-timeframe confluence
	RSI_5m     float64 `json:"rsi_5m"`   // 5-minute RSI
	RSI_15m    float64 `json:"rsi_15m"`  // 15-minute RSI
	RSI_1h     float64 `json:"rsi_1h"`   // 1-hour RSI
	Trend_5m   string  `json:"trend_5m"` // "UP", "DOWN", "SIDEWAYS"
	Trend_15m  string  `json:"trend_15m"`
	Trend_1h   string  `json:"trend_1h"`
	Price      float64 `json:"price"`
	RecentHigh float64 `json:"recent_high"`
	RecentLow  float64 `json:"recent_low"`
	// Enhanced market structure
	TrendStrength       float64 `json:"trend_strength"`    // ADX or similar trend strength 0-100
	SwingHighBroken     bool    `json:"swing_high_broken"` // Recent swing high broken
	SwingLowBroken      bool    `json:"swing_low_broken"`  // Recent swing low broken
	SupportLevel        float64 `json:"support_level"`     // Nearest support level
	ResistanceLevel     float64 `json:"resistance_level"`  // Nearest resistance level
	OIChange            float64 `json:"oi_change"`
	FundingRate         float64 `json:"funding_rate"`
	LongShortRatio      float64 `json:"long_short_ratio"`
	SpotVolumeChange    float64 `json:"spot_vol_change"`
	FuturesVolumeChange float64 `json:"futures_vol_change"`
	OrderBookImbalance  float64 `json:"order_book_imbalance"`
	// Enhanced volume analysis
	VWAP                float64 `json:"vwap"`                 // Volume Weighted Average Price
	RelativeVolume      float64 `json:"relative_volume"`      // Current vs 20-period average
	VolumeRatio         float64 `json:"volume_ratio"`         // Buy vs sell volume estimation
	VolumeAtPriceHigh   bool    `json:"volume_at_price_high"` // High volume at current price level
	ExchangeInflows     float64 `json:"exchange_inflows"`
	WhaleTxCount        int     `json:"whale_tx_count"`
	SpotFuturesPremium  float64 `json:"spot_futures_premium"`
	MacroSentimentScore float64 `json:"macro_sentiment_score"` // expected in [-1,1]
	NewsSentimentScore  float64 `json:"news_sentiment_score"`  // expected in [-1,1]
	FearGreedIndex      int     `json:"fear_greed_index"`      // 0..100
	ATRPercent          float64 `json:"atr_percent"`
	Capital             float64 `json:"capital"`
	CurrentPosition     string  `json:"current_position"`
}

type CategoryScores struct {
	MarketStructure  float64 `json:"market_structure"`
	VolumeOrderFlow  float64 `json:"volume_order_flow"`
	FundingLongShort float64 `json:"funding_long_short"`
	OnChain          float64 `json:"on_chain"`
	MacroSentiment   float64 `json:"macro_sentiment"`
	QuantModels      float64 `json:"quant_models"`
	RiskManagement   float64 `json:"risk_management"`
}

type DecisionOutput struct {
	Prediction          string         `json:"prediction"`
	Confidence          int            `json:"confidence"`
	Bias                string         `json:"bias"`
	TotalScore          float64        `json:"total_score"`
	CategoryScores      CategoryScores `json:"category_scores"`
	Action              string         `json:"action"`
	PositionSizePercent float64        `json:"position_size_percent"`
	Leverage            int            `json:"leverage"`
	EntryPrice          float64        `json:"entry_price"`
	StopLoss            float64        `json:"stop_loss"`
	TakeProfit          float64        `json:"take_profit"`
	ScaleInPlan         string         `json:"scale_in_plan"`
	ScaleOutPlan        string         `json:"scale_out_plan"`
	Reasoning           string         `json:"reasoning"`
}
