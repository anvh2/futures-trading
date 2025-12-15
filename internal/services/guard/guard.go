package guard

import (
	"fmt"
	"sync"
	"time"

	"github.com/anvh2/futures-trading/internal/libs/logger"
	"github.com/anvh2/futures-trading/internal/libs/queue"
	"github.com/anvh2/futures-trading/internal/services/state"
	"go.uber.org/zap"
)

// ExecutorController interface for controlling executor operations
type ExecutorController interface {
	Terminate() error
	Pause() error
	Resume() error
	CloseAllPositions() error
}

// SafetyRule represents a safety rule that can trigger circuit breakers
type SafetyRule interface {
	Check(state *state.TradingState) *SafetyViolation
	Name() string
	Priority() int // Higher number = higher priority
}

// SafetyViolation represents a violation of a safety rule
type SafetyViolation struct {
	RuleName  string                 `json:"rule_name"`
	Severity  ViolationSeverity      `json:"severity"`
	Message   string                 `json:"message"`
	Action    RecommendedAction      `json:"action"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ViolationSeverity represents the severity of a safety violation
type ViolationSeverity string

const (
	SeverityLow      ViolationSeverity = "LOW"
	SeverityMedium   ViolationSeverity = "MEDIUM"
	SeverityHigh     ViolationSeverity = "HIGH"
	SeverityCritical ViolationSeverity = "CRITICAL"
)

// RecommendedAction represents the recommended action for a violation
type RecommendedAction string

const (
	ActionWarn           RecommendedAction = "WARN"
	ActionPauseTrading   RecommendedAction = "PAUSE_TRADING"
	ActionClosePositions RecommendedAction = "CLOSE_POSITIONS"
	ActionEmergencyStop  RecommendedAction = "EMERGENCY_STOP"
)

// CircuitBreakerStatus represents the status of the circuit breaker
type CircuitBreakerStatus string

const (
	StatusNormal    CircuitBreakerStatus = "NORMAL"
	StatusTriggered CircuitBreakerStatus = "TRIGGERED"
	StatusCooldown  CircuitBreakerStatus = "COOLDOWN"
	StatusDisabled  CircuitBreakerStatus = "DISABLED"
)

// CircuitBreaker manages trading safety and emergency stops
type CircuitBreaker struct {
	Name             string               `json:"name"`
	Status           CircuitBreakerStatus `json:"status"`
	TriggerCount     int                  `json:"trigger_count"`
	LastTriggered    time.Time            `json:"last_triggered"`
	CooldownUntil    time.Time            `json:"cooldown_until"`
	CooldownDuration time.Duration        `json:"cooldown_duration"`
	MaxTriggers      int                  `json:"max_triggers"`
	IsEnabled        bool                 `json:"is_enabled"`
}

// SafetyGuard manages safety rules and circuit breakers
type SafetyGuard struct {
	logger          *logger.Logger
	stateManager    *state.StateManager
	rules           []SafetyRule
	circuitBreakers map[string]*CircuitBreaker
	violations      []*SafetyViolation
	listeners       []SafetyListener
	mutex           *sync.RWMutex
	ticker          *time.Ticker
	quitChannel     chan struct{}
	isRunning       bool
	checkInterval   time.Duration
	executorCtrl    ExecutorController
	queue           *queue.Queue
}

// SafetyListener interface for safety event notifications
type SafetyListener interface {
	OnViolationDetected(violation *SafetyViolation)
	OnCircuitBreakerTriggered(breakerName string, violation *SafetyViolation)
	OnCircuitBreakerReset(breakerName string)
}

// New creates a new SafetyGuard instance
func New(logger *logger.Logger, stateManager *state.StateManager) *SafetyGuard {
	guard := &SafetyGuard{
		logger:          logger,
		stateManager:    stateManager,
		rules:           make([]SafetyRule, 0),
		circuitBreakers: make(map[string]*CircuitBreaker),
		violations:      make([]*SafetyViolation, 0),
		listeners:       make([]SafetyListener, 0),
		mutex:           &sync.RWMutex{},
		quitChannel:     make(chan struct{}),
		checkInterval:   time.Second * 30, // Check every 30 seconds
	}

	// Initialize default circuit breakers
	guard.initializeDefaultCircuitBreakers()

	// Initialize default safety rules
	guard.initializeDefaultRules()

	return guard
}

// SetExecutorController sets the executor controller for emergency operations
func (sg *SafetyGuard) SetExecutorController(ctrl ExecutorController) {
	sg.mutex.Lock()
	defer sg.mutex.Unlock()
	sg.executorCtrl = ctrl
}

// SetQueue sets the message queue for emergency notifications
func (sg *SafetyGuard) SetQueue(queue *queue.Queue) {
	sg.mutex.Lock()
	defer sg.mutex.Unlock()
	sg.queue = queue
}

// AddRule adds a safety rule to the guard
func (sg *SafetyGuard) AddRule(rule SafetyRule) {
	sg.mutex.Lock()
	defer sg.mutex.Unlock()

	sg.rules = append(sg.rules, rule)
	sg.logger.Info("Safety rule added", zap.String("rule", rule.Name()))
}

// AddListener adds a safety event listener
func (sg *SafetyGuard) AddListener(listener SafetyListener) {
	sg.mutex.Lock()
	defer sg.mutex.Unlock()

	sg.listeners = append(sg.listeners, listener)
}

// Start begins the safety monitoring process
func (sg *SafetyGuard) Start() error {
	sg.mutex.Lock()
	defer sg.mutex.Unlock()

	if sg.isRunning {
		return fmt.Errorf("safety guard is already running")
	}

	sg.ticker = time.NewTicker(sg.checkInterval)
	sg.isRunning = true

	go sg.monitoringLoop()

	sg.logger.Info("Safety guard started", zap.Duration("check_interval", sg.checkInterval))
	return nil
}

// Stop stops the safety monitoring process
func (sg *SafetyGuard) Stop() error {
	sg.mutex.Lock()
	defer sg.mutex.Unlock()

	if !sg.isRunning {
		return fmt.Errorf("safety guard is not running")
	}

	close(sg.quitChannel)
	if sg.ticker != nil {
		sg.ticker.Stop()
	}
	sg.isRunning = false

	sg.logger.Info("Safety guard stopped")
	return nil
}

// CheckSafety manually triggers a safety check
func (sg *SafetyGuard) CheckSafety() []*SafetyViolation {
	sg.mutex.Lock()
	defer sg.mutex.Unlock()

	return sg.performSafetyCheck()
}

// GetStatus returns the current status of all circuit breakers
func (sg *SafetyGuard) GetStatus() map[string]*CircuitBreaker {
	sg.mutex.RLock()
	defer sg.mutex.RUnlock()

	status := make(map[string]*CircuitBreaker)
	for name, breaker := range sg.circuitBreakers {
		breakerCopy := *breaker
		status[name] = &breakerCopy
	}
	return status
}

// EnableCircuitBreaker enables a circuit breaker
func (sg *SafetyGuard) EnableCircuitBreaker(name string) error {
	sg.mutex.Lock()
	defer sg.mutex.Unlock()

	breaker, exists := sg.circuitBreakers[name]
	if !exists {
		return fmt.Errorf("circuit breaker '%s' not found", name)
	}

	breaker.IsEnabled = true
	sg.logger.Info("Circuit breaker enabled", zap.String("breaker", name))
	return nil
}

// DisableCircuitBreaker disables a circuit breaker
func (sg *SafetyGuard) DisableCircuitBreaker(name string) error {
	sg.mutex.Lock()
	defer sg.mutex.Unlock()

	breaker, exists := sg.circuitBreakers[name]
	if !exists {
		return fmt.Errorf("circuit breaker '%s' not found", name)
	}

	breaker.IsEnabled = false
	breaker.Status = StatusDisabled
	sg.logger.Info("Circuit breaker disabled", zap.String("breaker", name))
	return nil
}

// ResetCircuitBreaker manually resets a circuit breaker
func (sg *SafetyGuard) ResetCircuitBreaker(name string) error {
	sg.mutex.Lock()
	defer sg.mutex.Unlock()

	breaker, exists := sg.circuitBreakers[name]
	if !exists {
		return fmt.Errorf("circuit breaker '%s' not found", name)
	}

	breaker.Status = StatusNormal
	breaker.TriggerCount = 0
	breaker.CooldownUntil = time.Time{}

	sg.logger.Info("Circuit breaker reset", zap.String("breaker", name))

	// Notify listeners
	for _, listener := range sg.listeners {
		listener.OnCircuitBreakerReset(name)
	}

	return nil
}

// GetViolationHistory returns recent safety violations
func (sg *SafetyGuard) GetViolationHistory() []*SafetyViolation {
	sg.mutex.RLock()
	defer sg.mutex.RUnlock()

	// Return copy of violations
	violations := make([]*SafetyViolation, len(sg.violations))
	for i, v := range sg.violations {
		violationCopy := *v
		violations[i] = &violationCopy
	}
	return violations
}

// monitoringLoop runs the continuous safety monitoring
func (sg *SafetyGuard) monitoringLoop() {
	for {
		select {
		case <-sg.ticker.C:
			sg.mutex.Lock()
			sg.performSafetyCheck()
			sg.updateCircuitBreakerStatus()
			sg.mutex.Unlock()
		case <-sg.quitChannel:
			return
		}
	}
}

// performSafetyCheck executes all safety rules and handles violations
func (sg *SafetyGuard) performSafetyCheck() []*SafetyViolation {
	state := sg.stateManager.GetState()
	violations := make([]*SafetyViolation, 0)

	// Check all rules
	for _, rule := range sg.rules {
		if violation := rule.Check(state); violation != nil {
			violations = append(violations, violation)
			sg.handleViolation(violation)
		}
	}

	return violations
}

// handleViolation processes a safety violation
func (sg *SafetyGuard) handleViolation(violation *SafetyViolation) {
	// Add to violation history
	sg.violations = append(sg.violations, violation)

	// Keep only last 1000 violations to prevent memory issues
	if len(sg.violations) > 1000 {
		sg.violations = sg.violations[len(sg.violations)-1000:]
	}

	sg.logger.Warn("Safety violation detected",
		zap.String("rule", violation.RuleName),
		zap.String("severity", string(violation.Severity)),
		zap.String("message", violation.Message),
		zap.String("action", string(violation.Action)))

	// Notify listeners
	for _, listener := range sg.listeners {
		listener.OnViolationDetected(violation)
	}

	// Handle based on severity and recommended action
	sg.executeViolationAction(violation)
}

// executeViolationAction executes the recommended action for a violation
func (sg *SafetyGuard) executeViolationAction(violation *SafetyViolation) {
	switch violation.Action {
	case ActionWarn:
		// Just log and notify (already done)
		return

	case ActionPauseTrading:
		sg.stateManager.SetSystemStatus(state.SystemStatusPaused)
		sg.triggerCircuitBreaker("pause_trading", violation)

		// Pause executor if controller is available
		if sg.executorCtrl != nil {
			if err := sg.executorCtrl.Pause(); err != nil {
				sg.logger.Error("Failed to pause executor", zap.Error(err))
			}
		}

	case ActionClosePositions:
		sg.stateManager.SetSystemStatus(state.SystemStatusEmergency)
		sg.triggerCircuitBreaker("emergency_close", violation)

		// Close all positions through executor
		if sg.executorCtrl != nil {
			if err := sg.executorCtrl.CloseAllPositions(); err != nil {
				sg.logger.Error("Failed to close all positions", zap.Error(err))
			}
		}

		// Send emergency notification through queue
		if sg.queue != nil {
			emergencyMsg := map[string]interface{}{
				"type":      "emergency_close",
				"violation": violation,
				"timestamp": time.Now(),
			}
			sg.queue.Push(nil, "emergency", emergencyMsg)
		}

	case ActionEmergencyStop:
		sg.stateManager.SetSystemStatus(state.SystemStatusEmergency)
		sg.triggerCircuitBreaker("emergency_stop", violation)

		// Terminate executor
		if sg.executorCtrl != nil {
			if err := sg.executorCtrl.Terminate(); err != nil {
				sg.logger.Error("Failed to terminate executor", zap.Error(err))
			}
		}

		// Send critical emergency notification
		if sg.queue != nil {
			emergencyMsg := map[string]interface{}{
				"type":      "emergency_stop",
				"violation": violation,
				"timestamp": time.Now(),
			}
			sg.queue.Push(nil, "critical-emergency", emergencyMsg)
		}
	}
}

// triggerCircuitBreaker triggers a circuit breaker
func (sg *SafetyGuard) triggerCircuitBreaker(name string, violation *SafetyViolation) {
	breaker, exists := sg.circuitBreakers[name]
	if !exists || !breaker.IsEnabled {
		return
	}

	breaker.Status = StatusTriggered
	breaker.TriggerCount++
	breaker.LastTriggered = time.Now()
	breaker.CooldownUntil = time.Now().Add(breaker.CooldownDuration)

	sg.logger.Error("Circuit breaker triggered",
		zap.String("breaker", name),
		zap.Int("trigger_count", breaker.TriggerCount),
		zap.String("violation", violation.RuleName))

	// Notify listeners
	for _, listener := range sg.listeners {
		listener.OnCircuitBreakerTriggered(name, violation)
	}

	// Check if max triggers reached
	if breaker.MaxTriggers > 0 && breaker.TriggerCount >= breaker.MaxTriggers {
		breaker.IsEnabled = false
		breaker.Status = StatusDisabled
		sg.logger.Error("Circuit breaker disabled due to max triggers",
			zap.String("breaker", name),
			zap.Int("max_triggers", breaker.MaxTriggers))
	}
}

// updateCircuitBreakerStatus updates the status of circuit breakers
func (sg *SafetyGuard) updateCircuitBreakerStatus() {
	now := time.Now()

	for _, breaker := range sg.circuitBreakers {
		if !breaker.IsEnabled {
			continue
		}

		switch breaker.Status {
		case StatusTriggered:
			if now.After(breaker.CooldownUntil) {
				breaker.Status = StatusCooldown
			}
		case StatusCooldown:
			// Additional cooldown logic can be added here
			// For now, reset after cooldown period
			if now.After(breaker.CooldownUntil.Add(time.Minute * 5)) {
				breaker.Status = StatusNormal
			}
		}
	}
}

// initializeDefaultCircuitBreakers sets up default circuit breakers
func (sg *SafetyGuard) initializeDefaultCircuitBreakers() {
	breakers := map[string]*CircuitBreaker{
		"daily_loss_limit": {
			Name:             "daily_loss_limit",
			Status:           StatusNormal,
			CooldownDuration: time.Hour * 1,
			MaxTriggers:      3,
			IsEnabled:        true,
		},
		"max_positions": {
			Name:             "max_positions",
			Status:           StatusNormal,
			CooldownDuration: time.Minute * 30,
			MaxTriggers:      5,
			IsEnabled:        true,
		},
		"drawdown_limit": {
			Name:             "drawdown_limit",
			Status:           StatusNormal,
			CooldownDuration: time.Hour * 2,
			MaxTriggers:      2,
			IsEnabled:        true,
		},
		"emergency_stop": {
			Name:             "emergency_stop",
			Status:           StatusNormal,
			CooldownDuration: time.Hour * 24,
			MaxTriggers:      1,
			IsEnabled:        true,
		},
		"pause_trading": {
			Name:             "pause_trading",
			Status:           StatusNormal,
			CooldownDuration: time.Minute * 15,
			MaxTriggers:      10,
			IsEnabled:        true,
		},
		"emergency_close": {
			Name:             "emergency_close",
			Status:           StatusNormal,
			CooldownDuration: time.Hour * 4,
			MaxTriggers:      3,
			IsEnabled:        true,
		},
	}

	for name, breaker := range breakers {
		sg.circuitBreakers[name] = breaker
	}
}

// initializeDefaultRules sets up default safety rules
func (sg *SafetyGuard) initializeDefaultRules() {
	// Add default rules
	sg.rules = append(sg.rules,
		NewDailyLossLimitRule(),
		NewMaxPositionsRule(),
		NewDrawdownLimitRule(),
		NewAccountBalanceRule(),
		NewPositionSizeRule(),
		NewSystemStatusRule(),
		NewMarketVolatilityRule(),
		NewConnectionHealthRule(),
	)
}
