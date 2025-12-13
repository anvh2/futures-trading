package orchestrator

import (
	"github.com/anvh2/futures-trading/internal/libs/logger"
	"github.com/anvh2/futures-trading/internal/services/guard"
	"github.com/anvh2/futures-trading/internal/services/notify"
	"github.com/anvh2/futures-trading/internal/services/state"
	"go.uber.org/zap"
)

// SafetyEventHandler handles safety guard events
type SafetyEventHandler struct {
	logger       *logger.Logger
	stateManager *state.StateManager
	notifier     *notify.Notifier
}

// OnViolationDetected handles safety violations
func (h *SafetyEventHandler) OnViolationDetected(violation *guard.SafetyViolation) {
	h.logger.Warn("Safety violation detected",
		zap.String("rule", violation.RuleName),
		zap.String("severity", string(violation.Severity)),
		zap.String("message", violation.Message),
		zap.String("action", string(violation.Action)))

	// Additional logging or notification logic can be added here
}

// OnCircuitBreakerTriggered handles circuit breaker triggers
func (h *SafetyEventHandler) OnCircuitBreakerTriggered(breakerName string, violation *guard.SafetyViolation) {
	h.logger.Error("Circuit breaker triggered",
		zap.String("breaker", breakerName),
		zap.String("violation_rule", violation.RuleName),
		zap.String("severity", string(violation.Severity)))

	// Handle emergency actions based on breaker type
	switch breakerName {
	case "emergency_stop":
		h.stateManager.SetSystemStatus(state.SystemStatusEmergency)
	case "pause_trading":
		h.stateManager.SetSystemStatus(state.SystemStatusPaused)
	case "emergency_close":
		h.stateManager.SetSystemStatus(state.SystemStatusEmergency)
		// TODO: Trigger position closure
	}
}

// OnCircuitBreakerReset handles circuit breaker resets
func (h *SafetyEventHandler) OnCircuitBreakerReset(breakerName string) {
	h.logger.Info("Circuit breaker reset", zap.String("breaker", breakerName))

	// Potentially reset system status if all breakers are normal
	// This would require checking all breaker statuses
}

// StateEventHandler handles state change events
type StateEventHandler struct {
	logger   *logger.Logger
	notifier *notify.Notifier
	guard    *guard.SafetyGuard
}

// OnStateChanged handles general state changes
func (h *StateEventHandler) OnStateChanged(state *state.TradingState) {
	h.logger.Debug("Trading state updated",
		zap.String("system_status", string(state.SystemStatus)),
		zap.Int("active_positions", len(state.Positions)),
		zap.Int("pending_orders", len(state.PendingOrders)),
		zap.Float64("total_pnl", state.TotalPnL),
		zap.Float64("daily_pnl", state.DailyPnL))

	// Trigger safety check when state changes
	if h.guard != nil {
		h.guard.CheckSafety()
	}
}

// OnPositionUpdated handles position updates
func (h *StateEventHandler) OnPositionUpdated(position *state.Position) {
	h.logger.Info("Position updated",
		zap.String("symbol", position.Symbol),
		zap.String("side", position.Side),
		zap.Float64("size", position.Size),
		zap.Float64("entry_price", position.EntryPrice),
		zap.Float64("current_price", position.CurrentPrice),
		zap.Float64("unrealized_pnl", position.UnrealizedPnL),
		zap.Bool("is_active", position.IsActive))

	// TODO: Send notification through notifier if needed
}

// OnOrderUpdated handles order updates
func (h *StateEventHandler) OnOrderUpdated(order *state.PendingOrder) {
	h.logger.Info("Order updated",
		zap.String("order_id", order.OrderID),
		zap.String("symbol", order.Symbol),
		zap.String("side", order.Side),
		zap.String("status", string(order.Status)),
		zap.Float64("size", order.Size),
		zap.Float64("price", order.Price))

	// TODO: Send notification through notifier if needed
}

// OnTradeCompleted handles completed trades
func (h *StateEventHandler) OnTradeCompleted(trade *state.TradeRecord) {
	h.logger.Info("Trade completed",
		zap.String("trade_id", trade.ID),
		zap.String("symbol", trade.Symbol),
		zap.String("side", trade.Side),
		zap.Float64("entry_price", trade.EntryPrice),
		zap.Float64("exit_price", trade.ExitPrice),
		zap.Float64("realized_pnl", trade.RealizedPnL),
		zap.String("win_loss", trade.WinLoss),
		zap.Int64("duration_seconds", trade.Duration))

	// TODO: Send notification through notifier
}
