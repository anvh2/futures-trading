package exchange

import (
	"sync"

	"github.com/anvh2/futures-trading/internal/cache/errors"
	"github.com/anvh2/futures-trading/internal/logger"
)

type Exchange struct {
	logger   *logger.Logger
	mux      *sync.RWMutex
	symbols  []string
	internal map[string]*Symbol
}

func New(logger *logger.Logger) *Exchange {
	return &Exchange{
		logger:   logger,
		mux:      &sync.RWMutex{},
		internal: make(map[string]*Symbol),
	}
}

func (c *Exchange) Set(symbols []*Symbol) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.symbols = make([]string, len(symbols))

	for idx, symbol := range symbols {
		c.symbols[idx] = symbol.Symbol
		c.internal[symbol.Symbol] = symbol
	}
}

func (c *Exchange) Get(symbol string) (*Symbol, error) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	data, ok := c.internal[symbol]
	if !ok {
		return nil, errors.ErrorSymbolNotFound
	}

	return data, nil
}

func (c *Exchange) Symbols() []string {
	return c.symbols
}
