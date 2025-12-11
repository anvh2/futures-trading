package decision

// DecisionInput carries all metrics required by the framework
type DecisionInput struct {
	Symbol              string  `json:"symbol"`
	Timeframe           string  `json:"timeframe"`
	RSI                 float64 `json:"rsi"`
	K                   float64 `json:"k"`
	D                   float64 `json:"d"`
	J                   float64 `json:"j"`
	Price               float64 `json:"price"`
	RecentHigh          float64 `json:"recent_high"`
	RecentLow           float64 `json:"recent_low"`
	OIChange            float64 `json:"oi_change"`
	FundingRate         float64 `json:"funding_rate"`
	LongShortRatio      float64 `json:"long_short_ratio"`
	SpotVolumeChange    float64 `json:"spot_vol_change"`
	FuturesVolumeChange float64 `json:"futures_vol_change"`
	OrderBookImbalance  float64 `json:"order_book_imbalance"`
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
