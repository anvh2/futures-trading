package state

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/anvh2/futures-trading/internal/libs/logger"
	"github.com/anvh2/futures-trading/internal/libs/storage/simpledb"
	"github.com/anvh2/futures-trading/internal/models"
	"go.uber.org/zap"
)

// TradingState represents the current state of trading operations
type TradingState struct {
	Positions      map[string]*Position     `json:"positions"`
	PendingOrders  map[string]*PendingOrder `json:"pending_orders"`
	TradingHistory []*TradeRecord           `json:"trading_history"`
	RiskMetrics    *RiskMetrics             `json:"risk_metrics"`
	SystemStatus   SystemStatus             `json:"system_status"`
	LastUpdated    time.Time                `json:"last_updated"`
	TotalPnL       float64                  `json:"total_pnl"`
	DailyPnL       float64                  `json:"daily_pnl"`
	DrawDown       float64                  `json:"drawdown"`
	MaxDrawDown    float64                  `json:"max_drawdown"`
}

// Position represents an open trading position
type Position struct {
	Symbol        string    `json:"symbol"`
	Side          string    `json:"side"` // LONG or SHORT
	Size          float64   `json:"size"`
	EntryPrice    float64   `json:"entry_price"`
	CurrentPrice  float64   `json:"current_price"`
	UnrealizedPnL float64   `json:"unrealized_pnl"`
	Leverage      float64   `json:"leverage"`
	OpenTime      time.Time `json:"open_time"`
	StopLoss      float64   `json:"stop_loss,omitempty"`
	TakeProfit    float64   `json:"take_profit,omitempty"`
	IsActive      bool      `json:"is_active"`
}

// PendingOrder represents orders that are waiting to be executed
type PendingOrder struct {
	OrderID       string         `json:"order_id"`
	Symbol        string         `json:"symbol"`
	Side          string         `json:"side"`
	Type          string         `json:"type"`
	Size          float64        `json:"size"`
	Price         float64        `json:"price"`
	StopPrice     float64        `json:"stop_price,omitempty"`
	Status        OrderStatus    `json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	TriggerSignal *models.Signal `json:"trigger_signal,omitempty"`
}

// TradeRecord represents a completed trade
type TradeRecord struct {
	ID          string    `json:"id"`
	Symbol      string    `json:"symbol"`
	Side        string    `json:"side"`
	EntryPrice  float64   `json:"entry_price"`
	ExitPrice   float64   `json:"exit_price"`
	Size        float64   `json:"size"`
	RealizedPnL float64   `json:"realized_pnl"`
	Commission  float64   `json:"commission"`
	OpenTime    time.Time `json:"open_time"`
	CloseTime   time.Time `json:"close_time"`
	Duration    int64     `json:"duration_seconds"`
	Strategy    string    `json:"strategy"`
	WinLoss     string    `json:"win_loss"` // WIN, LOSS, or BREAKEVEN
}

// RiskMetrics tracks risk-related metrics
type RiskMetrics struct {
	ExposureRatio    float64   `json:"exposure_ratio"`     // Total exposure / Account balance
	PositionCount    int       `json:"position_count"`     // Number of open positions
	MaxPositions     int       `json:"max_positions"`      // Maximum allowed positions
	DailyLossLimit   float64   `json:"daily_loss_limit"`   // Maximum daily loss allowed
	CurrentDailyLoss float64   `json:"current_daily_loss"` // Current daily loss
	WinRate          float64   `json:"win_rate"`           // Percentage of winning trades
	ProfitFactor     float64   `json:"profit_factor"`      // Gross profit / Gross loss
	SharpeRatio      float64   `json:"sharpe_ratio"`       // Risk-adjusted return
	LastRiskUpdate   time.Time `json:"last_risk_update"`
}

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusSubmitted OrderStatus = "SUBMITTED"
	OrderStatusFilled    OrderStatus = "FILLED"
	OrderStatusCanceled  OrderStatus = "CANCELED"
	OrderStatusRejected  OrderStatus = "REJECTED"
)

// SystemStatus represents the overall system status
type SystemStatus string

const (
	SystemStatusActive      SystemStatus = "ACTIVE"
	SystemStatusPaused      SystemStatus = "PAUSED"
	SystemStatusEmergency   SystemStatus = "EMERGENCY"
	SystemStatusMaintenance SystemStatus = "MAINTENANCE"
)

// StateManager manages the trading state
type StateManager struct {
	db        simpledb.DB
	logger    *logger.Logger
	state     *TradingState
	mutex     *sync.RWMutex
	listeners []StateListener
}

// StateListener interface for state change notifications
type StateListener interface {
	OnStateChanged(state *TradingState)
	OnPositionUpdated(position *Position)
	OnOrderUpdated(order *PendingOrder)
	OnTradeCompleted(trade *TradeRecord)
}

// New creates a new StateManager instance
func New(logger *logger.Logger, db simpledb.DB) *StateManager {
	sm := &StateManager{
		logger:    logger,
		state:     &TradingState{},
		mutex:     &sync.RWMutex{},
		listeners: make([]StateListener, 0),
		db:        db,
	}

	// Try to load existing state
	if db != nil {
		if err := db.Load(sm.state); err == nil {
			logger.Info("Loaded existing state")
		} else {
			logger.Warn("Failed to load existing state, creating new", zap.Error(err))
		}
	}

	// Initialize with empty state if no state loaded
	if sm.state == nil {
		sm.state = &TradingState{
			Positions:      make(map[string]*Position),
			PendingOrders:  make(map[string]*PendingOrder),
			TradingHistory: make([]*TradeRecord, 0),
			RiskMetrics: &RiskMetrics{
				MaxPositions:   10,     // Default max positions
				DailyLossLimit: 1000.0, // Default daily loss limit
			},
			SystemStatus: SystemStatusActive,
			LastUpdated:  time.Now(),
		}
	}

	return sm
}

// AddListener adds a state listener
func (sm *StateManager) AddListener(listener StateListener) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.listeners = append(sm.listeners, listener)
}

// GetState returns the current state (read-only copy)
func (sm *StateManager) GetState() *TradingState {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// Return a deep copy to prevent external modifications
	stateCopy := *sm.state

	// Deep copy maps and slices
	stateCopy.Positions = make(map[string]*Position)
	for k, v := range sm.state.Positions {
		positionCopy := *v
		stateCopy.Positions[k] = &positionCopy
	}

	stateCopy.PendingOrders = make(map[string]*PendingOrder)
	for k, v := range sm.state.PendingOrders {
		orderCopy := *v
		stateCopy.PendingOrders[k] = &orderCopy
	}

	stateCopy.TradingHistory = make([]*TradeRecord, len(sm.state.TradingHistory))
	for i, v := range sm.state.TradingHistory {
		tradeCopy := *v
		stateCopy.TradingHistory[i] = &tradeCopy
	}

	if sm.state.RiskMetrics != nil {
		riskCopy := *sm.state.RiskMetrics
		stateCopy.RiskMetrics = &riskCopy
	}

	return &stateCopy
}

// UpdatePosition updates or creates a position
func (sm *StateManager) UpdatePosition(position *Position) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.state.Positions[position.Symbol] = position
	sm.state.LastUpdated = time.Now()

	// Update risk metrics
	sm.updateRiskMetrics()

	// Save state
	sm.saveState()

	// Notify listeners
	for _, listener := range sm.listeners {
		listener.OnPositionUpdated(position)
		listener.OnStateChanged(sm.state)
	}
}

// ClosePosition closes a position and records the trade
func (sm *StateManager) ClosePosition(symbol string, exitPrice float64, commission float64) *TradeRecord {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	position, exists := sm.state.Positions[symbol]
	if !exists || !position.IsActive {
		sm.logger.Warn("Attempted to close non-existent or inactive position", zap.String("symbol", symbol))
		return nil
	}

	// Create trade record
	trade := &TradeRecord{
		ID:         sm.generateTradeID(),
		Symbol:     position.Symbol,
		Side:       position.Side,
		EntryPrice: position.EntryPrice,
		ExitPrice:  exitPrice,
		Size:       position.Size,
		Commission: commission,
		OpenTime:   position.OpenTime,
		CloseTime:  time.Now(),
		Duration:   int64(time.Since(position.OpenTime).Seconds()),
	}

	// Calculate PnL
	if position.Side == "LONG" {
		trade.RealizedPnL = (exitPrice-position.EntryPrice)*position.Size - commission
	} else {
		trade.RealizedPnL = (position.EntryPrice-exitPrice)*position.Size - commission
	}

	// Determine win/loss
	if trade.RealizedPnL > 0 {
		trade.WinLoss = "WIN"
	} else if trade.RealizedPnL < 0 {
		trade.WinLoss = "LOSS"
	} else {
		trade.WinLoss = "BREAKEVEN"
	}

	// Update totals
	sm.state.TotalPnL += trade.RealizedPnL
	sm.state.DailyPnL += trade.RealizedPnL

	// Mark position as inactive
	position.IsActive = false

	// Add to trading history
	sm.state.TradingHistory = append(sm.state.TradingHistory, trade)

	// Update timestamps
	sm.state.LastUpdated = time.Now()

	// Update risk metrics
	sm.updateRiskMetrics()

	// Save state
	sm.saveState()

	// Notify listeners
	for _, listener := range sm.listeners {
		listener.OnTradeCompleted(trade)
		listener.OnStateChanged(sm.state)
	}

	return trade
}

// UpdateOrder updates a pending order
func (sm *StateManager) UpdateOrder(order *PendingOrder) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	order.UpdatedAt = time.Now()
	sm.state.PendingOrders[order.OrderID] = order
	sm.state.LastUpdated = time.Now()

	// Save state
	sm.saveState()

	// Notify listeners
	for _, listener := range sm.listeners {
		listener.OnOrderUpdated(order)
		listener.OnStateChanged(sm.state)
	}
}

// RemoveOrder removes a pending order
func (sm *StateManager) RemoveOrder(orderID string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	delete(sm.state.PendingOrders, orderID)
	sm.state.LastUpdated = time.Now()

	// Save state
	sm.saveState()

	// Notify listeners
	for _, listener := range sm.listeners {
		listener.OnStateChanged(sm.state)
	}
}

// SetSystemStatus sets the system status
func (sm *StateManager) SetSystemStatus(status SystemStatus) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.state.SystemStatus = status
	sm.state.LastUpdated = time.Now()

	// Save state
	sm.saveState()

	sm.logger.Info("System status changed", zap.String("status", string(status)))

	// Notify listeners
	for _, listener := range sm.listeners {
		listener.OnStateChanged(sm.state)
	}
}

// GetSystemStatus returns the current system status
func (sm *StateManager) GetSystemStatus() SystemStatus {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.state.SystemStatus
}

// GetPositions returns current positions
func (sm *StateManager) GetPositions() map[string]*Position {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	positions := make(map[string]*Position)
	for k, v := range sm.state.Positions {
		if v.IsActive {
			positionCopy := *v
			positions[k] = &positionCopy
		}
	}
	return positions
}

// GetPendingOrders returns current pending orders
func (sm *StateManager) GetPendingOrders() map[string]*PendingOrder {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	orders := make(map[string]*PendingOrder)
	for k, v := range sm.state.PendingOrders {
		orderCopy := *v
		orders[k] = &orderCopy
	}
	return orders
}

// GetRiskMetrics returns current risk metrics
func (sm *StateManager) GetRiskMetrics() *RiskMetrics {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if sm.state.RiskMetrics == nil {
		return nil
	}

	riskCopy := *sm.state.RiskMetrics
	return &riskCopy
}

// updateRiskMetrics updates risk metrics (internal method, should be called with lock held)
func (sm *StateManager) updateRiskMetrics() {
	if sm.state.RiskMetrics == nil {
		return
	}

	// Update position count
	activePositions := 0
	for _, position := range sm.state.Positions {
		if position.IsActive {
			activePositions++
		}
	}
	sm.state.RiskMetrics.PositionCount = activePositions

	// Calculate win rate from recent trades
	if len(sm.state.TradingHistory) > 0 {
		wins := 0
		for _, trade := range sm.state.TradingHistory {
			if trade.WinLoss == "WIN" {
				wins++
			}
		}
		sm.state.RiskMetrics.WinRate = float64(wins) / float64(len(sm.state.TradingHistory)) * 100
	}

	sm.state.RiskMetrics.LastRiskUpdate = time.Now()
}

// generateTradeID generates a unique trade ID
func (sm *StateManager) generateTradeID() string {
	return "trade_" + time.Now().Format("20060102_150405") + "_" + time.Now().Format("000")
}

// saveState persists the current state
func (sm *StateManager) saveState() {
	if sm.db != nil {
		if err := sm.db.Save(sm.state); err != nil {
			sm.logger.Error("Failed to save state", zap.Error(err))
		}
	}
}

// String returns string representation of state
func (sm *StateManager) String() string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	data, err := json.MarshalIndent(sm.state, "", "  ")
	if err != nil {
		return "Error marshaling state: " + err.Error()
	}
	return string(data)
}

// Shutdown gracefully shuts down the state manager
func (sm *StateManager) Shutdown() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Final state save
	if sm.db != nil {
		if err := sm.db.Save(sm.state); err != nil {
			return err
		}
		// Create backup
		if err := sm.db.Backup(); err != nil {
			sm.logger.Warn("Failed to create backup", zap.Error(err))
		}
	}

	sm.logger.Info("State manager shutdown complete")
	return nil
}
