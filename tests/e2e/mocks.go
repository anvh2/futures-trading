package e2e

import (
	"sync"
	"time"

	"github.com/anvh2/futures-trading/internal/cache/exchange"
	"github.com/anvh2/futures-trading/internal/models"
)

// MockBinance implements a simple mock for testing
type MockBinance struct {
	mu             sync.RWMutex
	marketData     map[string][]*models.Candlestick
	isConnected    bool
	shouldFail     bool
	networkLatency time.Duration
}

func NewMockBinance() *MockBinance {
	return &MockBinance{
		marketData:     make(map[string][]*models.Candlestick),
		isConnected:    true,
		networkLatency: 50 * time.Millisecond,
	}
}

func (m *MockBinance) GetKlineData(symbol, interval string, limit int) ([]*models.Candlestick, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.shouldFail {
		return nil, &MockError{"mock binance error"}
	}

	// Simulate network latency
	time.Sleep(m.networkLatency)

	if data, exists := m.marketData[symbol]; exists {
		return data, nil
	}

	return []*models.Candlestick{}, nil
}

// Test control methods
func (m *MockBinance) SetMarketData(symbol string, candles []*models.Candlestick) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.marketData[symbol] = candles
}

func (m *MockBinance) SetShouldFail(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = fail
}

// MockTelegram implements the telegram interface for testing
type MockTelegram struct {
	mu       sync.RWMutex
	messages []string
	enabled  bool
}

func NewMockTelegram() *MockTelegram {
	return &MockTelegram{
		messages: make([]string, 0),
		enabled:  true,
	}
}

func (m *MockTelegram) SendMessage(message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.enabled {
		return &MockError{"telegram disabled"}
	}

	m.messages = append(m.messages, message)
	return nil
}

func (m *MockTelegram) GetMessages() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]string, len(m.messages))
	copy(result, m.messages)
	return result
}

func (m *MockTelegram) ClearMessages() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = make([]string, 0)
}

func (m *MockTelegram) SetEnabled(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = enabled
}

// MockMarketCache implements the market cache interface for testing
type MockMarketCache struct {
	mu      sync.RWMutex
	candles map[string][]*models.Candlestick
}

func NewMockMarketCache() *MockMarketCache {
	return &MockMarketCache{
		candles: make(map[string][]*models.Candlestick),
	}
}

func (m *MockMarketCache) Set(key string, candles []*models.Candlestick) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.candles[key] = candles
	return nil
}

func (m *MockMarketCache) Get(key string) ([]*models.Candlestick, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if data, exists := m.candles[key]; exists {
		return data, nil
	}
	return nil, &MockError{"not found"}
}

func (m *MockMarketCache) GetLatest(symbol, interval string, limit int) ([]*models.Candlestick, error) {
	key := symbol + ":" + interval
	candles, err := m.Get(key)
	if err != nil {
		return nil, err
	}

	if len(candles) <= limit {
		return candles, nil
	}

	return candles[len(candles)-limit:], nil
}

func (m *MockMarketCache) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.candles, key)
	return nil
}

// MockExchangeCache implements the exchange cache interface for testing
type MockExchangeCache struct {
	mu       sync.RWMutex
	symbols  map[string]*exchange.Symbol
	isLoaded bool
}

func NewMockExchangeCache() *MockExchangeCache {
	return &MockExchangeCache{
		symbols:  make(map[string]*exchange.Symbol),
		isLoaded: true,
	}
}

func (m *MockExchangeCache) GetSymbol(symbol string) (*exchange.Symbol, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if sym, exists := m.symbols[symbol]; exists {
		return sym, nil
	}

	return nil, &MockError{"symbol not found"}
}

func (m *MockExchangeCache) GetAllSymbols() (map[string]*exchange.Symbol, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]*exchange.Symbol)
	for k, v := range m.symbols {
		result[k] = v
	}
	return result, nil
}

func (m *MockExchangeCache) IsLoaded() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isLoaded
}

// Helper methods for testing
func (m *MockExchangeCache) SetExchangeInfo(symbols []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, symbolName := range symbols {
		m.symbols[symbolName] = &exchange.Symbol{
			Symbol:    symbolName,
			BaseAsset: symbolName[:3], // Simplified
		}
	}
}

// MockError represents a mock error for testing
type MockError struct {
	Message string
}

func (e *MockError) Error() string {
	return e.Message
}
