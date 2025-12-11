package market

import (
	"sync"
	"time"

	"github.com/anvh2/futures-trading/internal/cache/errors"
	"github.com/anvh2/futures-trading/internal/libs/cache/circular"
	"github.com/anvh2/futures-trading/internal/models"
)

type SummaryData struct {
	Candles    *circular.Cache `json:"candles,omitempty"`
	CreateTime int64           `json:"create_time,omitempty"`
	UpdateTime int64           `json:"update_time,omitempty"`
}

type CandleSummary struct {
	mutex  *sync.RWMutex
	symbol string                  // key
	cache  map[string]*SummaryData // map[interval]candles
	limit  int32                   // limit of candles's length
}

func (m *CandleSummary) Init(symbol string, limit int32) *CandleSummary {
	return &CandleSummary{
		mutex:  &sync.RWMutex{},
		symbol: symbol,
		cache:  make(map[string]*SummaryData),
		limit:  limit,
	}
}

func (m *CandleSummary) Candles(interval string) (*circular.Cache, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.cache[interval] == nil {
		return nil, errors.ErrorCandlesNotFound
	}

	return m.cache[interval].Candles, nil
}

func (m *CandleSummary) CreateCandle(interval string, candle *models.Candlestick) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.cache[interval] == nil {
		m.cache[interval] = &SummaryData{
			Candles:    circular.New(m.limit),
			CreateTime: time.Now().UnixMilli(),
			UpdateTime: time.Now().UnixMilli(),
		}
	}

	m.cache[interval].Candles.Insert(candle)
	m.cache[interval].UpdateTime = time.Now().UnixMilli()
	return nil
}

func (m *CandleSummary) UpdateCandle(interval string, candleId int32, candle *models.Candlestick) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.cache[interval] == nil {
		m.cache[interval] = &SummaryData{
			Candles:    circular.New(m.limit),
			UpdateTime: time.Now().UnixMilli(),
		}
	}

	m.cache[interval].Candles.Update(candleId, candle)
	m.cache[interval].UpdateTime = time.Now().UnixMilli()
}

func (m *CandleSummary) SummaryData(interval string) *SummaryData {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.cache[interval] == nil {
		return &SummaryData{}
	}

	return m.cache[interval]
}
