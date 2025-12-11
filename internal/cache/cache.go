package cache

import (
	"github.com/anvh2/futures-trading/internal/cache/exchange"
	"github.com/anvh2/futures-trading/internal/cache/market"
)

//go:generate moq -pkg cachemock -out ./mocks/market_mock.go . Market
type Market interface {
	CandleSummary(symbol string) (*market.CandleSummary, error)
	CreateSummary(symbol string) *market.CandleSummary
	UpdateSummary(symbol string) *market.CandleSummary
}

//go:generate moq -pkg cachemock -out ./mocks/exchange_mock.go . Exchange
type Exchange interface {
	Set(symbols []*exchange.Symbol)
	Get(symbol string) (*exchange.Symbol, error)
	Symbols() []string
}

//go:generate moq -pkg cachemock -out ./mocks/basic_mock.go . Basic
// type Basic interface {
// 	Set(key string, value interface{})
// 	Get(key string) interface{}
// 	Exs(key string) bool
// 	SetEX(key string, value interface{}) (interface{}, bool)
// }
