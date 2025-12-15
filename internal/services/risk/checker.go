package risk

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/anvh2/futures-trading/internal/libs/logger"
	"github.com/anvh2/futures-trading/internal/libs/queue"
	"github.com/anvh2/futures-trading/internal/models"
	"github.com/anvh2/futures-trading/internal/services/guard"
	"github.com/anvh2/futures-trading/internal/services/state"
	"go.uber.org/zap"
)

// Checker interface defines the risk checking capabilities
type Checker interface {
	CheckDecision(decision *models.TradingDecision) bool
	Start() error // Start queue processing
	Stop()        // Stop processing
}

// Checker implementation
type CheckerImpl struct {
	logger      *logger.Logger
	config      *Config
	state       *state.StateManager
	guard       *guard.SafetyGuard
	queue       *queue.Queue
	quitChannel chan struct{}
}

// NewChecker creates a new risk checker
func NewChecker(
	logger *logger.Logger,
	config *Config,
	stateManager *state.StateManager,
	safetyGuard *guard.SafetyGuard,
	queue *queue.Queue,
) Checker {
	if config == nil {
		defaultConfig := DefaultConfig()
		config = &defaultConfig
	}

	return &CheckerImpl{
		logger:      logger,
		config:      config,
		state:       stateManager,
		guard:       safetyGuard,
		queue:       queue,
		quitChannel: make(chan struct{}),
	}
}

// CheckDecision evaluates a trading decision against risk parameters
func (re *CheckerImpl) CheckDecision(decision *models.TradingDecision) bool {
	re.logger.Debug("Checking trading decision",
		zap.String("symbol", decision.Symbol),
		zap.String("action", decision.Action))

	// 1. Check system status
	systemStatus := re.state.GetSystemStatus()
	if systemStatus != state.SystemStatusActive {
		re.logger.Info("Decision rejected - system not active",
			zap.String("status", string(systemStatus)))
		return false
	}

	// 2. Check position limits
	if !re.checkPositionLimits(decision) {
		return false
	}

	// 3. Check position sizing
	if !re.checkPositionSizing(decision) {
		return false
	}

	// 4. Check daily loss limits
	if !re.checkDailyLimits(decision) {
		return false
	}

	// 5. Check confidence thresholds
	if !re.checkConfidenceThresholds(decision) {
		return false
	}

	// 6. Check guard safety violations
	if !re.checkGuardSafety(decision) {
		return false
	}

	// 7. Check exposure limits
	if !re.checkExposureLimits(decision) {
		return false
	}

	// 8. Check correlations and diversification
	if !re.checkCorrelationLimits(decision) {
		return false
	}

	re.logger.Info("Decision approved by risk engine",
		zap.String("symbol", decision.Symbol),
		zap.String("action", decision.Action))

	return true
}

// Start begins the queue processing goroutine
func (re *CheckerImpl) Start() error {
	if err := re.handleDecisions(); err != nil {
		return err
	}

	return nil
}

// Stop stops the queue processing
func (re *CheckerImpl) Stop() {
	close(re.quitChannel)
}

// handleDecisions processes trading decisions from the queue
func (re *CheckerImpl) handleDecisions() error {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				re.logger.Error("[handleDecisions] failed to process", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
			}
		}()

		ticker := time.NewTicker(2 * time.Second)

		for {
			select {
			case <-ticker.C:
				msg, err := re.queue.Consume(context.Background(), "decisions", "risk-checker")
				if err != nil {
					continue
				}

				if err := re.handleDecision(msg); err != nil {
					re.logger.Error("[handleDecisions] failed to handle decision", zap.Error(err))
					msg.Commit(context.Background())
					continue
				}

				msg.Commit(context.Background())

			case <-re.quitChannel:
				return
			}
		}
	}()

	return nil
}

// handleDecision processes a single trading decision message
func (re *CheckerImpl) handleDecision(msg *queue.Message) error {
	decision, ok := msg.Data.(*models.TradingDecision)
	if !ok {
		return nil
	}

	// Check if decision passes risk validation
	if !re.CheckDecision(decision) {
		re.logger.Info("Decision rejected by risk engine",
			zap.String("symbol", decision.Symbol),
			zap.String("action", decision.Action))
		return nil
	}

	// Decision approved - forward to executor queue
	if err := re.queue.Push(context.Background(), "approved-orders", decision); err != nil {
		re.logger.Error("Failed to push approved decision to executor queue", zap.Error(err))
		return err
	}

	re.logger.Info("Decision approved and forwarded to executor",
		zap.String("symbol", decision.Symbol),
		zap.String("action", decision.Action),
		zap.Float64("size", decision.Size))

	return nil
}

// checkPositionLimits checks if the decision violates position limits
func (re *CheckerImpl) checkPositionLimits(decision *models.TradingDecision) bool {
	riskMetrics := re.state.GetRiskMetrics()
	if riskMetrics == nil {
		return true // No limits set
	}

	positions := re.state.GetPositions()
	activePositions := len(positions)

	// Don't open new positions if at limit
	if decision.Action == "BUY" || decision.Action == "SELL" {
		if activePositions >= riskMetrics.MaxPositions {
			re.logger.Info("Decision rejected - max positions reached",
				zap.Int("active_positions", activePositions),
				zap.Int("max_positions", riskMetrics.MaxPositions))
			return false
		}
	}

	return true
}

// checkPositionSizing checks if the position size is appropriate
func (re *CheckerImpl) checkPositionSizing(decision *models.TradingDecision) bool {
	// Check minimum and maximum position sizes
	const minSize = 0.001
	const maxSize = 10.0

	if decision.Size < minSize {
		re.logger.Info("Decision rejected - position size too small",
			zap.Float64("size", decision.Size),
			zap.Float64("min_size", minSize))
		return false
	}

	if decision.Size > maxSize {
		re.logger.Info("Decision rejected - position size too large",
			zap.Float64("size", decision.Size),
			zap.Float64("max_size", maxSize))
		return false
	}

	return true
}

// checkDailyLimits checks if the decision would violate daily limits
func (re *CheckerImpl) checkDailyLimits(decision *models.TradingDecision) bool {
	state := re.state.GetState()
	riskMetrics := state.RiskMetrics

	if riskMetrics == nil || riskMetrics.DailyLossLimit <= 0 {
		return true // No limits set
	}

	// Check current daily loss
	currentDailyLoss := -state.DailyPnL // Negative PnL is loss
	if currentDailyLoss > 0 {
		lossPercentage := (currentDailyLoss / riskMetrics.DailyLossLimit) * 100

		// Don't allow new positions if we're close to daily loss limit
		if lossPercentage > 80 && (decision.Action == "BUY" || decision.Action == "SELL") {
			re.logger.Info("Decision rejected - approaching daily loss limit",
				zap.Float64("current_loss", currentDailyLoss),
				zap.Float64("loss_limit", riskMetrics.DailyLossLimit),
				zap.Float64("loss_percentage", lossPercentage))
			return false
		}
	}

	return true
}

// checkConfidenceThresholds checks if the decision meets confidence requirements
func (re *CheckerImpl) checkConfidenceThresholds(decision *models.TradingDecision) bool {
	// Set minimum confidence thresholds
	const minConfidenceForEntry = 0.6 // 60% confidence for new positions
	const minConfidenceForClose = 0.4 // 40% confidence for closing positions

	var minRequired float64
	if decision.Action == "CLOSE" {
		minRequired = minConfidenceForClose
	} else {
		minRequired = minConfidenceForEntry
	}

	if decision.Confidence < minRequired {
		re.logger.Info("Decision rejected - insufficient confidence",
			zap.Float64("confidence", decision.Confidence),
			zap.Float64("min_required", minRequired),
			zap.String("action", decision.Action))
		return false
	}

	return true
}

// checkGuardSafety checks if the guard system allows trading
func (re *CheckerImpl) checkGuardSafety(decision *models.TradingDecision) bool {
	// Check if any circuit breakers are triggered that would prevent trading
	status := re.guard.GetStatus()

	for name, breaker := range status {
		if breaker.Status == guard.CircuitBreakerStatus("TRIGGERED") &&
			time.Now().Before(breaker.CooldownUntil) {
			re.logger.Info("Decision rejected - circuit breaker active",
				zap.String("breaker", name),
				zap.Time("cooldown_until", breaker.CooldownUntil))
			return false
		}
	}

	// Perform real-time safety check
	violations := re.guard.CheckSafety()
	for _, violation := range violations {
		if violation.Severity == guard.ViolationSeverity("CRITICAL") ||
			violation.Severity == guard.ViolationSeverity("HIGH") {
			if violation.Action == guard.RecommendedAction("PAUSE_TRADING") ||
				violation.Action == guard.RecommendedAction("EMERGENCY_STOP") {
				re.logger.Info("Decision rejected - safety violation detected",
					zap.String("rule", violation.RuleName),
					zap.String("severity", string(violation.Severity)),
					zap.String("action", string(violation.Action)))
				return false
			}
		}
	}

	return true
}

// checkExposureLimits checks portfolio exposure limits
func (re *CheckerImpl) checkExposureLimits(decision *models.TradingDecision) bool {
	riskMetrics := re.state.GetRiskMetrics()
	if riskMetrics == nil {
		return true // No limits set
	}

	// Check total exposure ratio
	positions := re.state.GetPositions()
	totalExposure := 0.0

	for _, position := range positions {
		totalExposure += position.Size * position.EntryPrice
	}

	// Add potential exposure from this decision
	if decision.Action == "BUY" || decision.Action == "SELL" {
		totalExposure += decision.Size * decision.Price
	}

	// Check against account balance (assuming 100k for now - should be configurable)
	accountBalance := 100000.0 // Should come from account service
	exposureRatio := totalExposure / accountBalance

	maxExposureRatio := 0.8 // 80% max exposure
	if exposureRatio > maxExposureRatio {
		re.logger.Info("Decision rejected - exposure limit exceeded",
			zap.Float64("exposure_ratio", exposureRatio),
			zap.Float64("max_exposure_ratio", maxExposureRatio))
		return false
	}

	return true
}

// checkCorrelationLimits checks position correlation and diversification
func (re *CheckerImpl) checkCorrelationLimits(decision *models.TradingDecision) bool {
	if decision.Action != "BUY" && decision.Action != "SELL" {
		return true // Only check for new positions
	}

	positions := re.state.GetPositions()

	// Simple correlation check: limit similar assets
	// In production, this should use actual correlation matrices
	symbol := decision.Symbol
	baseAsset := symbol[:3] // Extract base asset (e.g., BTC from BTCUSDT)

	sameAssetPositions := 0
	for _, position := range positions {
		if position.Symbol[:3] == baseAsset {
			sameAssetPositions++
		}
	}

	// Limit to 2 positions per base asset
	maxSameAsset := 2
	if sameAssetPositions >= maxSameAsset {
		re.logger.Info("Decision rejected - too many positions in same asset",
			zap.String("base_asset", baseAsset),
			zap.Int("current_positions", sameAssetPositions),
			zap.Int("max_allowed", maxSameAsset))
		return false
	}

	return true
}
