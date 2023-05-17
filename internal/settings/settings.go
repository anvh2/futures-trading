package settings

import "github.com/anvh2/futures-trading/internal/services/binance"

var (
	DefaultSettings = NewDefaultSettings()
)

type TradingStrategy byte

const (
	TradingStrategyInvalid = iota
	TradingStrategyInstantNoodles
	TradingStrategyDollarCostAveraging // recommended within 1h interval
)

type PNL struct {
	GainPricePercent float64 `json:"gain_price,omitempty"`
	LossPricePercent float64 `json:"loss_price,omitempty"`
	DesiredProfit    float64 `json:"disired_profit,omitempty"`
	DesiredLoss      float64 `json:"disired_loss,omitempty"`
	// ProfitROE        float64 `json:"profit_roe,omitempty"`
	// LossROE          float64 `json:"loss_roe,omitempty"`
}

type Settings struct {
	SignalDisabled         bool            `json:"signal_disabled,omitempty"`
	TradingEnabled         bool            `json:"trading_enabled,omitempty"`
	TradingCost            float64         `json:"trading_cost,omitempty"`
	TradingInterval        string          `json:"trading_interval,omitempty"`
	TradingStrategy        TradingStrategy `json:"trading_strategy,omitempty"`
	MaxPositionsDaily      int32           `json:"max_positions_daily,omitempty"`
	MaxPositionsPerTime    int32           `json:"max_positions_per_time,omitempty"`
	PreferLeverageBrackets []int           `json:"prefer_leverage_brackets,omitempty"`
	LongPNL                *PNL            `json:"long_pnl,omitempty"`
	ShortPNL               *PNL            `json:"short_pnl,omitempty"`
}

func NewDefaultSettings() *Settings {
	return &Settings{
		SignalDisabled:         false,
		TradingEnabled:         true,
		TradingCost:            10, // USD
		TradingInterval:        "15m",
		TradingStrategy:        TradingStrategyInstantNoodles,
		MaxPositionsDaily:      300,
		MaxPositionsPerTime:    3,
		PreferLeverageBrackets: []int{20, 10},
		LongPNL: &PNL{
			GainPricePercent: 1.2,
			LossPricePercent: 0.8,
			DesiredProfit:    1.2,
			DesiredLoss:      -10, // TODO
		},
		ShortPNL: &PNL{
			GainPricePercent: 0.8,
			LossPricePercent: 1.2,
			DesiredProfit:    1.2,
			DesiredLoss:      -10, // TODO
		},
	}
}

func (s *Settings) GetPreferLeverage(leverageBrackets []*binance.LeverageBracket) int {
	for _, lv := range s.PreferLeverageBrackets {
		for _, lb := range leverageBrackets {
			for _, b := range lb.Brackets {
				if b.InitialLeverage == lv {
					return lv
				}
			}
		}
	}
	return 5
}
