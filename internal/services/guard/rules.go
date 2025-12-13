package guard

import (
	"fmt"
	"time"

	"github.com/anvh2/futures-trading/internal/services/state"
)

// DailyLossLimitRule checks if daily losses exceed the limit
type DailyLossLimitRule struct{}

func NewDailyLossLimitRule() *DailyLossLimitRule {
	return &DailyLossLimitRule{}
}

func (r *DailyLossLimitRule) Name() string {
	return "daily_loss_limit"
}

func (r *DailyLossLimitRule) Priority() int {
	return 100 // High priority
}

func (r *DailyLossLimitRule) Check(state *state.TradingState) *SafetyViolation {
	if state.RiskMetrics == nil {
		return nil
	}

	dailyLoss := -state.DailyPnL // Negative PnL means loss
	if dailyLoss <= 0 {
		return nil // No loss or profitable
	}

	limit := state.RiskMetrics.DailyLossLimit
	if limit <= 0 {
		return nil // No limit set
	}

	lossPercentage := (dailyLoss / limit) * 100

	if dailyLoss >= limit {
		return &SafetyViolation{
			RuleName:  r.Name(),
			Severity:  SeverityCritical,
			Message:   fmt.Sprintf("Daily loss limit exceeded: %.2f / %.2f (%.1f%%)", dailyLoss, limit, lossPercentage),
			Action:    ActionEmergencyStop,
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"daily_loss":      dailyLoss,
				"loss_limit":      limit,
				"loss_percentage": lossPercentage,
			},
		}
	}

	if lossPercentage >= 80 {
		return &SafetyViolation{
			RuleName:  r.Name(),
			Severity:  SeverityHigh,
			Message:   fmt.Sprintf("Daily loss approaching limit: %.2f / %.2f (%.1f%%)", dailyLoss, limit, lossPercentage),
			Action:    ActionPauseTrading,
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"daily_loss":      dailyLoss,
				"loss_limit":      limit,
				"loss_percentage": lossPercentage,
			},
		}
	}

	return nil
}

// MaxPositionsRule checks if the number of positions exceeds the limit
type MaxPositionsRule struct{}

func NewMaxPositionsRule() *MaxPositionsRule {
	return &MaxPositionsRule{}
}

func (r *MaxPositionsRule) Name() string {
	return "max_positions"
}

func (r *MaxPositionsRule) Priority() int {
	return 80
}

func (r *MaxPositionsRule) Check(state *state.TradingState) *SafetyViolation {
	if state.RiskMetrics == nil {
		return nil
	}

	activePositions := 0
	for _, position := range state.Positions {
		if position.IsActive {
			activePositions++
		}
	}

	maxPositions := state.RiskMetrics.MaxPositions
	if maxPositions <= 0 {
		return nil // No limit set
	}

	if activePositions >= maxPositions {
		return &SafetyViolation{
			RuleName:  r.Name(),
			Severity:  SeverityHigh,
			Message:   fmt.Sprintf("Maximum positions exceeded: %d / %d", activePositions, maxPositions),
			Action:    ActionPauseTrading,
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"active_positions": activePositions,
				"max_positions":    maxPositions,
			},
		}
	}

	if float64(activePositions)/float64(maxPositions) >= 0.9 {
		return &SafetyViolation{
			RuleName:  r.Name(),
			Severity:  SeverityMedium,
			Message:   fmt.Sprintf("Position count near limit: %d / %d", activePositions, maxPositions),
			Action:    ActionWarn,
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"active_positions": activePositions,
				"max_positions":    maxPositions,
			},
		}
	}

	return nil
}

// DrawdownLimitRule checks if drawdown exceeds acceptable levels
type DrawdownLimitRule struct{}

func NewDrawdownLimitRule() *DrawdownLimitRule {
	return &DrawdownLimitRule{}
}

func (r *DrawdownLimitRule) Name() string {
	return "drawdown_limit"
}

func (r *DrawdownLimitRule) Priority() int {
	return 90
}

func (r *DrawdownLimitRule) Check(state *state.TradingState) *SafetyViolation {
	const maxDrawdownPercentage = 20.0 // 20% max drawdown

	if state.DrawDown <= 0 {
		return nil // No drawdown
	}

	if state.DrawDown >= maxDrawdownPercentage {
		return &SafetyViolation{
			RuleName:  r.Name(),
			Severity:  SeverityCritical,
			Message:   fmt.Sprintf("Drawdown limit exceeded: %.2f%% (max: %.2f%%)", state.DrawDown, maxDrawdownPercentage),
			Action:    ActionEmergencyStop,
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"current_drawdown": state.DrawDown,
				"max_drawdown":     state.MaxDrawDown,
				"drawdown_limit":   maxDrawdownPercentage,
			},
		}
	}

	if state.DrawDown >= maxDrawdownPercentage*0.8 {
		return &SafetyViolation{
			RuleName:  r.Name(),
			Severity:  SeverityHigh,
			Message:   fmt.Sprintf("Drawdown approaching limit: %.2f%% (max: %.2f%%)", state.DrawDown, maxDrawdownPercentage),
			Action:    ActionPauseTrading,
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"current_drawdown": state.DrawDown,
				"max_drawdown":     state.MaxDrawDown,
				"drawdown_limit":   maxDrawdownPercentage,
			},
		}
	}

	return nil
}

// AccountBalanceRule checks for unusually low account balance
type AccountBalanceRule struct{}

func NewAccountBalanceRule() *AccountBalanceRule {
	return &AccountBalanceRule{}
}

func (r *AccountBalanceRule) Name() string {
	return "account_balance"
}

func (r *AccountBalanceRule) Priority() int {
	return 95
}

func (r *AccountBalanceRule) Check(state *state.TradingState) *SafetyViolation {
	// This is a placeholder - in a real implementation, you would get
	// the actual account balance from the exchange
	const minimumBalanceUSD = 100.0

	// For now, we'll estimate based on PnL
	// In practice, you'd query the actual balance from Binance
	estimatedBalance := 1000.0 + state.TotalPnL // Assuming starting balance of $1000

	if estimatedBalance <= minimumBalanceUSD {
		return &SafetyViolation{
			RuleName:  r.Name(),
			Severity:  SeverityCritical,
			Message:   fmt.Sprintf("Account balance critically low: $%.2f", estimatedBalance),
			Action:    ActionEmergencyStop,
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"estimated_balance": estimatedBalance,
				"minimum_balance":   minimumBalanceUSD,
				"total_pnl":         state.TotalPnL,
			},
		}
	}

	if estimatedBalance <= minimumBalanceUSD*3 {
		return &SafetyViolation{
			RuleName:  r.Name(),
			Severity:  SeverityHigh,
			Message:   fmt.Sprintf("Account balance low: $%.2f", estimatedBalance),
			Action:    ActionPauseTrading,
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"estimated_balance": estimatedBalance,
				"minimum_balance":   minimumBalanceUSD,
				"total_pnl":         state.TotalPnL,
			},
		}
	}

	return nil
}

// PositionSizeRule checks for excessive position sizes
type PositionSizeRule struct{}

func NewPositionSizeRule() *PositionSizeRule {
	return &PositionSizeRule{}
}

func (r *PositionSizeRule) Name() string {
	return "position_size"
}

func (r *PositionSizeRule) Priority() int {
	return 70
}

func (r *PositionSizeRule) Check(state *state.TradingState) *SafetyViolation {
	const maxSinglePositionValueUSD = 5000.0
	const maxTotalExposureUSD = 15000.0

	var totalExposure float64
	var largestPosition float64
	var largestPositionSymbol string

	for symbol, position := range state.Positions {
		if !position.IsActive {
			continue
		}

		positionValue := position.Size * position.CurrentPrice * position.Leverage
		totalExposure += positionValue

		if positionValue > largestPosition {
			largestPosition = positionValue
			largestPositionSymbol = symbol
		}
	}

	// Check total exposure
	if totalExposure > maxTotalExposureUSD {
		return &SafetyViolation{
			RuleName:  r.Name(),
			Severity:  SeverityHigh,
			Message:   fmt.Sprintf("Total exposure too high: $%.2f (max: $%.2f)", totalExposure, maxTotalExposureUSD),
			Action:    ActionPauseTrading,
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"total_exposure":     totalExposure,
				"max_total_exposure": maxTotalExposureUSD,
			},
		}
	}

	// Check single position size
	if largestPosition > maxSinglePositionValueUSD {
		return &SafetyViolation{
			RuleName:  r.Name(),
			Severity:  SeverityMedium,
			Message:   fmt.Sprintf("Single position too large: %s $%.2f (max: $%.2f)", largestPositionSymbol, largestPosition, maxSinglePositionValueUSD),
			Action:    ActionWarn,
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"largest_position":        largestPosition,
				"largest_position_symbol": largestPositionSymbol,
				"max_single_position":     maxSinglePositionValueUSD,
			},
		}
	}

	return nil
}

// SystemStatusRule checks system status consistency
type SystemStatusRule struct{}

func NewSystemStatusRule() *SystemStatusRule {
	return &SystemStatusRule{}
}

func (r *SystemStatusRule) Name() string {
	return "system_status"
}

func (r *SystemStatusRule) Priority() int {
	return 60
}

func (r *SystemStatusRule) Check(state *state.TradingState) *SafetyViolation {
	// Check if system has been in emergency status too long
	if state.SystemStatus == "EMERGENCY" {
		timeSinceUpdate := time.Since(state.LastUpdated)
		if timeSinceUpdate > time.Hour*24 {
			return &SafetyViolation{
				RuleName:  r.Name(),
				Severity:  SeverityMedium,
				Message:   fmt.Sprintf("System in emergency status for too long: %v", timeSinceUpdate),
				Action:    ActionWarn,
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"system_status":     string(state.SystemStatus),
					"time_since_update": timeSinceUpdate.String(),
				},
			}
		}
	}

	// Check if there are active positions but system is paused
	if state.SystemStatus == "PAUSED" {
		activePositions := 0
		for _, position := range state.Positions {
			if position.IsActive {
				activePositions++
			}
		}

		if activePositions > 0 {
			return &SafetyViolation{
				RuleName:  r.Name(),
				Severity:  SeverityMedium,
				Message:   fmt.Sprintf("System paused but %d active positions remain", activePositions),
				Action:    ActionWarn,
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"system_status":    string(state.SystemStatus),
					"active_positions": activePositions,
				},
			}
		}
	}

	return nil
}
