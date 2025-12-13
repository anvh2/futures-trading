package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anvh2/futures-trading/internal/cache"
	"github.com/anvh2/futures-trading/internal/config"
	"github.com/anvh2/futures-trading/internal/externals/binance"
	"github.com/anvh2/futures-trading/internal/externals/telegram"
	"github.com/anvh2/futures-trading/internal/libs/channel"
	"github.com/anvh2/futures-trading/internal/libs/logger"
	"github.com/anvh2/futures-trading/internal/libs/queue"
	"github.com/anvh2/futures-trading/internal/libs/storage/simpledb"
	"github.com/anvh2/futures-trading/internal/models"
	analyzer "github.com/anvh2/futures-trading/internal/services/analyze"
	"github.com/anvh2/futures-trading/internal/services/decision"
	"github.com/anvh2/futures-trading/internal/services/guard"
	"github.com/anvh2/futures-trading/internal/services/market"
	notify "github.com/anvh2/futures-trading/internal/services/notify"
	orderer "github.com/anvh2/futures-trading/internal/services/order"
	risk "github.com/anvh2/futures-trading/internal/services/risk"
	"github.com/anvh2/futures-trading/internal/services/settings"
	"github.com/anvh2/futures-trading/internal/services/state"
	"go.uber.org/zap"
)

// ServiceOrchestrator coordinates all trading services according to the architecture
type ServiceOrchestrator struct {
	config config.Config
	logger *logger.Logger

	// Core services
	stateManager   *state.StateManager
	safetyGuard    *guard.SafetyGuard
	marketService  *market.Market
	analyzer       *analyzer.Analyzer
	decisionEngine decision.Maker
	riskEngine     risk.Checker
	orderExecutor  *orderer.Orderer
	notifier       *notify.Notifier

	// External dependencies
	binance       *binance.Binance
	telegram      *telegram.TelegramBot
	marketCache   cache.Market
	exchangeCache cache.Exchange
	queue         *queue.Queue
	channel       *channel.Channel
	settings      *settings.Settings

	// Control
	isRunning bool
	stopChan  chan struct{}
	wg        *sync.WaitGroup
	mutex     *sync.RWMutex
}

// NewServiceOrchestrator creates a new service orchestrator
func NewServiceOrchestrator(
	config config.Config,
	logger *logger.Logger,
	binance *binance.Binance,
	telegram *telegram.TelegramBot,
	marketCache cache.Market,
	exchangeCache cache.Exchange,
	queue *queue.Queue,
	channel *channel.Channel,
	settings *settings.Settings,
) (*ServiceOrchestrator, error) {
	// Initialize state management
	db, err := simpledb.NewStorage(logger, "./data/state.json", "./data/backups")
	if err != nil {
		return nil, fmt.Errorf("failed to create state db: %w", err)
	}

	stateManager := state.New(logger, db)

	// Initialize safety guard
	safetyGuard := guard.New(logger, stateManager)

	// Initialize market service
	marketService := market.New(logger, binance, telegram, marketCache, exchangeCache, channel)

	// Initialize analyzer
	analyzer := analyzer.New(config, logger, telegram, marketCache, exchangeCache, queue, channel, settings)

	// Initialize decision engine
	decisionEngine := decision.NewMaker(logger, stateManager, safetyGuard)

	// Initialize risk engine
	defaultConfig := risk.DefaultConfig()
	riskEngine := risk.NewChecker(logger, &defaultConfig, stateManager, safetyGuard)

	// Initialize order executor
	orderExecutor := orderer.New(config, logger, telegram, marketCache, exchangeCache, queue, settings)

	// Initialize notifier
	notifier := notify.New(config, logger, binance, telegram, queue)

	so := &ServiceOrchestrator{
		config:         config,
		logger:         logger,
		stateManager:   stateManager,
		safetyGuard:    safetyGuard,
		marketService:  marketService,
		analyzer:       analyzer,
		decisionEngine: decisionEngine,
		riskEngine:     riskEngine,
		orderExecutor:  orderExecutor,
		notifier:       notifier,
		binance:        binance,
		telegram:       telegram,
		marketCache:    marketCache,
		exchangeCache:  exchangeCache,
		queue:          queue,
		channel:        channel,
		settings:       settings,
		stopChan:       make(chan struct{}),
		wg:             &sync.WaitGroup{},
		mutex:          &sync.RWMutex{},
	}

	// Set up cross-service communication
	so.setupCommunication()

	return so, nil
}

// setupCommunication sets up communication channels between services
func (so *ServiceOrchestrator) setupCommunication() {
	// Add state manager as safety guard listener
	so.safetyGuard.AddListener(&SafetyEventHandler{
		logger:       so.logger,
		stateManager: so.stateManager,
		notifier:     so.notifier,
	})

	// Add state listener for state changes
	so.stateManager.AddListener(&StateEventHandler{
		logger:   so.logger,
		notifier: so.notifier,
		guard:    so.safetyGuard,
	})
}

// Start starts all services in the correct order
func (so *ServiceOrchestrator) Start(ctx context.Context) error {
	so.mutex.Lock()
	defer so.mutex.Unlock()

	if so.isRunning {
		return fmt.Errorf("service orchestrator is already running")
	}

	so.logger.Info("Starting futures trading system...")

	// 1. Start state management
	so.logger.Info("Starting state management...")

	// 2. Start safety guard
	so.logger.Info("Starting safety guard...")
	if err := so.safetyGuard.Start(); err != nil {
		return fmt.Errorf("failed to start safety guard: %w", err)
	}

	// 3. Start market service (data collection)
	so.logger.Info("Starting market service...")
	if err := so.marketService.Start(); err != nil {
		return fmt.Errorf("failed to start market service: %w", err)
	}

	// 4. Start analyzer
	so.logger.Info("Starting analyzer...")
	if err := so.analyzer.Start(); err != nil {
		return fmt.Errorf("failed to start analyzer: %w", err)
	}

	// 5. Start main trading loop
	so.logger.Info("Starting main trading loop...")
	so.wg.Add(1)
	go so.tradingLoop(ctx)

	so.isRunning = true
	so.logger.Info("Futures trading system started successfully")

	return nil
}

// Stop gracefully stops all services
func (so *ServiceOrchestrator) Stop() error {
	so.mutex.Lock()
	defer so.mutex.Unlock()

	if !so.isRunning {
		return fmt.Errorf("service orchestrator is not running")
	}

	so.logger.Info("Stopping futures trading system...")

	// Signal stop to all services
	close(so.stopChan)

	// Wait for trading loop to finish
	so.wg.Wait()

	// Stop services in reverse order
	so.logger.Info("Stopping analyzer...")
	so.analyzer.Stop()

	so.logger.Info("Stopping market service...")
	so.marketService.Stop()

	so.logger.Info("Stopping safety guard...")
	if err := so.safetyGuard.Stop(); err != nil {
		so.logger.Error("Error stopping safety guard", zap.Error(err))
	}

	so.logger.Info("Stopping order executor...")
	so.orderExecutor.Stop()

	so.logger.Info("Stopping notifier...")
	so.notifier.Stop()

	so.logger.Info("Shutting down state manager...")
	if err := so.stateManager.Shutdown(); err != nil {
		so.logger.Error("Error shutting down state manager", zap.Error(err))
	}

	so.isRunning = false
	so.logger.Info("Futures trading system stopped")

	return nil
}

// tradingLoop implements the main trading flow according to the architecture
func (so *ServiceOrchestrator) tradingLoop(ctx context.Context) {
	defer so.wg.Done()

	so.logger.Info("Trading loop started")

	for {
		select {
		case <-ctx.Done():
			so.logger.Info("Trading loop stopped by context")
			return
		case <-so.stopChan:
			so.logger.Info("Trading loop stopped by stop signal")
			return
		default:
			// Check system status
			systemStatus := so.stateManager.GetSystemStatus()
			if systemStatus != state.SystemStatusActive {
				so.logger.Debug("Trading loop paused", zap.String("status", string(systemStatus)))
				continue
			}

			// Execute trading cycle
			so.executeTradingCycle()
		}
	}
}

// executeTradingCycle implements one cycle of the trading flow
func (so *ServiceOrchestrator) executeTradingCycle() {
	// 1. Market data is continuously collected by market service

	// 2. Analyze & generate signals (done by analyzer service)

	// 3. Decision making (Maker)
	decisions := so.decisionEngine.MakeDecisions()
	if len(decisions) == 0 {
		return // No decisions to process
	}

	// 4. Risk checking (Checker)
	for _, decision := range decisions {
		approved := so.riskEngine.CheckDecision(decision)
		if !approved {
			so.logger.Info("Decision rejected by risk engine",
				zap.String("symbol", decision.Symbol),
				zap.String("action", decision.Action))
			continue
		}

		// 5. Order execution (Executor)
		so.executeDecision(decision)
	}
}

// executeDecision executes an approved trading decision
func (so *ServiceOrchestrator) executeDecision(decision *models.TradingDecision) {
	so.logger.Info("Executing trading decision",
		zap.String("symbol", decision.Symbol),
		zap.String("action", decision.Action),
		zap.Float64("confidence", decision.Confidence))

	// Update state with pending order
	pendingOrder := &state.PendingOrder{
		OrderID:   so.generateOrderID(),
		Symbol:    decision.Symbol,
		Side:      decision.Action,
		Type:      "MARKET", // Default to market orders
		Size:      decision.Size,
		Price:     decision.Price,
		Status:    state.OrderStatusPending,
		CreatedAt: decision.Timestamp,
		UpdatedAt: decision.Timestamp,
	}

	so.stateManager.UpdateOrder(pendingOrder)

	// TODO: Implement actual order execution through order service
	// For now, we'll simulate successful execution
	so.simulateOrderExecution(pendingOrder)
}

// simulateOrderExecution simulates order execution (placeholder)
func (so *ServiceOrchestrator) simulateOrderExecution(order *state.PendingOrder) {
	// In a real implementation, this would go through the order service
	// and interact with Binance API

	so.logger.Info("Simulating order execution",
		zap.String("order_id", order.OrderID),
		zap.String("symbol", order.Symbol),
		zap.String("side", order.Side))

	// Update order status to filled
	order.Status = state.OrderStatusFilled
	so.stateManager.UpdateOrder(order)

	// Create position if it's a new position
	if order.Type != "CLOSE" {
		position := &state.Position{
			Symbol:        order.Symbol,
			Side:          order.Side,
			Size:          order.Size,
			EntryPrice:    order.Price,
			CurrentPrice:  order.Price,
			UnrealizedPnL: 0,
			Leverage:      1, // Default leverage
			OpenTime:      order.CreatedAt,
			IsActive:      true,
		}
		so.stateManager.UpdatePosition(position)
	}

	// Remove from pending orders
	so.stateManager.RemoveOrder(order.OrderID)
}

// generateOrderID generates a unique order ID
func (so *ServiceOrchestrator) generateOrderID() string {
	return fmt.Sprintf("ord_%d", time.Now().UnixNano())
}

// GetStatus returns the current status of all services
func (so *ServiceOrchestrator) GetStatus() map[string]interface{} {
	so.mutex.RLock()
	defer so.mutex.RUnlock()

	return map[string]interface{}{
		"is_running":        so.isRunning,
		"system_status":     string(so.stateManager.GetSystemStatus()),
		"active_positions":  len(so.stateManager.GetPositions()),
		"pending_orders":    len(so.stateManager.GetPendingOrders()),
		"circuit_breakers":  so.safetyGuard.GetStatus(),
		"violation_history": len(so.safetyGuard.GetViolationHistory()),
	}
}
