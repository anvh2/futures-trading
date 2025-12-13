package models

import "time"

// TradingDecision represents a trading decision from the decision engine
type TradingDecision struct {
	Symbol     string                 `json:"symbol"`
	Action     string                 `json:"action"` // BUY, SELL, CLOSE
	Size       float64                `json:"size"`
	Price      float64                `json:"price"`
	Confidence float64                `json:"confidence"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}
