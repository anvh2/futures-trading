package risk

import (
	"github.com/anvh2/futures-trading/internal/libs/logger"
	"github.com/anvh2/futures-trading/internal/models"
	"github.com/anvh2/futures-trading/internal/services/guard"
	"github.com/anvh2/futures-trading/internal/services/state"
	"go.uber.org/zap"
)

// Checker interface defines the risk checking capabilities
type Checker interface {
	CheckDecision(decision *models.TradingDecision) bool
}

// Checker implementation
type CheckerImpl struct {
	logger *logger.Logger
	config *Config
	state  *state.StateManager
	guard  *guard.SafetyGuard
}

// NewChecker creates a new risk checker
func NewChecker(
	logger *logger.Logger,
	config *Config,
	stateManager *state.StateManager,
	safetyGuard *guard.SafetyGuard,
) Checker {
	if config == nil {
		defaultConfig := DefaultConfig()
		config = &defaultConfig
	}

	return &CheckerImpl{
		logger: logger,
		config: config,
		state:  stateManager,
		guard:  safetyGuard,
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

	re.logger.Info("Decision approved by risk engine",
		zap.String("symbol", decision.Symbol),
		zap.String("action", decision.Action))

	return true
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
