package market

import (
	"sync"

	"github.com/anvh2/futures-trading/internal/cache/circular"
	"github.com/anvh2/futures-trading/internal/cache/errors"
)

type Market struct {
	mutex *sync.Mutex
	cache map[string]*CandleSummary // map[symbol]summary
	limit int32
}

func NewMarket(limit int32) *Market {
	return &Market{
		mutex: &sync.Mutex{},
		cache: make(map[string]*CandleSummary),
		limit: limit,
	}
}

func (c *Market) CandleSummary(symbol string) (*CandleSummary, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.cache[symbol] == nil {
		return nil, errors.ErrorChartNotFound
	}

	return c.cache[symbol], nil
}

func (c *Market) CreateSummary(symbol string) *CandleSummary {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.cache[symbol] == nil {
		market := new(CandleSummary)
		c.cache[symbol] = market.Init(symbol, c.limit)
	}

	return c.cache[symbol]
}

func (c *Market) UpdateSummary(symbol string) *CandleSummary {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.cache[symbol] == nil {
		market := new(CandleSummary)
		c.cache[symbol] = market.Init(symbol, c.limit)
	}

	return c.cache[symbol]
}

func (c *Market) Candles(symbol, interval string) *circular.Cache {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.cache[symbol] == nil {
		market := new(CandleSummary)
		c.cache[symbol] = market.Init(symbol, c.limit)
	}

	summary := c.cache[symbol].cache[interval]
	if summary == nil {
		summary = &SummaryData{
			Candles: circular.New(c.limit),
		}
	}

	return summary.Candles
}
