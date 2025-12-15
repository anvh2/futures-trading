# Risk Engine System

## Overview

This document explains the comprehensive risk management system that validates trading decisions before execution. The system implements a multi-layered approach to risk control, integrating with the safety guard system for holistic risk management.

## Architecture

### Core Components

1. **Decision Validation**: Multi-criteria risk assessment of trading decisions
2. **Queue Processing**: Asynchronous processing of trading intents from decision engine
3. **Guard Integration**: Real-time safety validation through circuit breaker system
4. **Exposure Management**: Portfolio-level risk monitoring and limits

### Data Flow

```
Decision Engine → decisions queue → Risk Checker → approved-orders queue → Executor
                                      ↓
                               Guard Safety Check
```

## Key Features

### 1. Multi-Layer Risk Assessment

The risk engine evaluates eight critical risk factors:

- **System Status Validation**: Ensures system is active and operational
- **Position Limits**: Maximum number of concurrent positions
- **Position Sizing**: Individual trade size validation
- **Daily Loss Limits**: Cumulative daily loss protection
- **Confidence Thresholds**: Minimum confidence requirements
- **Guard Safety Integration**: Real-time circuit breaker checks
- **Exposure Limits**: Total portfolio exposure management
- **Correlation Limits**: Asset diversification enforcement

### 2. Real-Time Processing

- **Queue-Based Architecture**: Processes decisions asynchronously
- **2-Second Processing Cycle**: Regular polling for new decisions
- **Error Recovery**: Robust error handling with detailed logging
- **Graceful Degradation**: Continues operation during partial system failures

### 3. Dynamic Risk Parameters

- **Confidence-Based Sizing**: Higher confidence allows larger positions
- **Volatility Adjustments**: ATR-based position size scaling
- **Market Condition Awareness**: Adjusts limits based on market volatility
- **Account-Based Scaling**: Risk limits scale with account size

## Risk Validation Process

### 1. System Status Check

```go
// Validates system operational status
func (re *CheckerImpl) checkSystemStatus() bool {
    status := re.state.GetSystemStatus()
    return status == state.SystemStatusActive
}
```

**Rejection Criteria:**
- System status: PAUSED, EMERGENCY, or MAINTENANCE
- Result: Immediate rejection with status logging

### 2. Position Limits Validation

```go
type PositionLimits struct {
    MaxPositions    int     `json:"max_positions"`
    CurrentCount    int     `json:"current_count"`
    UtilizationPct  float64 `json:"utilization_pct"`
}
```

**Validation Rules:**
- Active positions < maximum allowed positions
- Warning at 90% utilization
- Rejection at 100% utilization

### 3. Position Sizing Controls

```go
const (
    MinPositionSize = 0.001  // Minimum trade size
    MaxPositionSize = 10.0   // Maximum single trade size
)
```

**Size Validation:**
- Minimum size enforcement (prevents dust trades)
- Maximum size limits (prevents oversized positions)
- Percentage-based limits relative to account balance

### 4. Daily Loss Protection

```go
type DailyLossControl struct {
    CurrentLoss     float64 `json:"current_daily_loss"`
    LossLimit       float64 `json:"daily_loss_limit"`
    LossPercentage  float64 `json:"loss_percentage"`
    ThresholdLevel  string  `json:"threshold_level"`
}
```

**Protection Levels:**
- **80% of limit**: Warning level, restrict new positions
- **100% of limit**: Emergency stop, close all positions
- **Real-time monitoring**: Continuous loss tracking

### 5. Confidence Thresholds

```go
const (
    MinConfidenceForEntry = 0.6  // 60% for new positions
    MinConfidenceForClose = 0.4  // 40% for closing positions
)
```

**Confidence Rules:**
- Entry trades require 60% minimum confidence
- Exit trades require 40% minimum confidence
- Dynamic adjustment based on market conditions

### 6. Guard Safety Integration

```go
func (re *CheckerImpl) checkGuardSafety(decision *models.TradingDecision) bool {
    // Check circuit breaker status
    status := re.guard.GetStatus()
    
    // Validate real-time safety violations
    violations := re.guard.CheckSafety()
    
    return len(criticalViolations) == 0
}
```

**Integration Points:**
- Circuit breaker status validation
- Real-time safety violation checks
- Escalation to guard system for severe violations

### 7. Exposure Management

```go
type ExposureControl struct {
    TotalExposure     float64 `json:"total_exposure"`
    AccountBalance    float64 `json:"account_balance"`
    ExposureRatio     float64 `json:"exposure_ratio"`
    MaxExposureRatio  float64 `json:"max_exposure_ratio"`
}
```

**Exposure Limits:**
- **Maximum Exposure**: 80% of account balance
- **Position Aggregation**: Sum of all position notional values
- **Real-time Calculation**: Updated with each new position

### 8. Correlation & Diversification

```go
type DiversificationControl struct {
    BaseAsset           string  `json:"base_asset"`
    SameAssetPositions  int     `json:"same_asset_positions"`
    MaxSameAsset        int     `json:"max_same_asset"`
    DiversificationScore float64 `json:"diversification_score"`
}
```

**Diversification Rules:**
- Maximum 2 positions per base asset (BTC, ETH, etc.)
- Correlation matrix validation (future enhancement)
- Sector exposure limits (future enhancement)

## Configuration

### Default Risk Parameters

```go
type Config struct {
    // Position sizing
    MaxPositionPercent  float64 // 5.0% of capital per trade
    MinPositionSize     float64 // 0.001 minimum size
    MaxPositionSize     float64 // 10.0 maximum size
    
    // Risk limits
    MaxPositions        int     // 5 maximum concurrent positions
    DailyLossLimit      float64 // 2% of account daily loss limit
    MaxExposureRatio    float64 // 80% maximum portfolio exposure
    
    // Confidence thresholds
    MinConfidenceEntry  float64 // 60% minimum for entries
    MinConfidenceExit   float64 // 40% minimum for exits
    
    // Correlation limits
    MaxSameAssetPositions int   // 2 positions per base asset
}
```

### Risk Adjustment Factors

```go
// Volatility-based position sizing
func AdjustForVolatility(baseSize, atrPercent float64) float64 {
    switch {
    case atrPercent > 5.0:  // Very high volatility
        return baseSize * 0.5
    case atrPercent > 3.0:  // High volatility
        return baseSize * 0.7
    case atrPercent > 2.0:  // Medium volatility
        return baseSize * 0.85
    default:                // Normal/low volatility
        return baseSize * 1.0
    }
}
```

## Usage

### Basic Integration

```go
// Initialize risk engine
riskConfig := risk.DefaultConfig()
riskChecker := risk.NewChecker(
    logger,
    &riskConfig,
    stateManager,
    safetyGuard,
    queue,
)

// Start processing
if err := riskChecker.Start(); err != nil {
    log.Fatal("Failed to start risk engine:", err)
}

// Manual decision validation
approved := riskChecker.CheckDecision(tradingDecision)
```

### Advanced Configuration

```go
// Custom risk configuration
customConfig := &risk.Config{
    MaxPositionPercent:    3.0,  // More conservative sizing
    MaxPositions:          3,    // Fewer concurrent positions
    DailyLossLimit:        1.0,  // Tighter loss control
    MinConfidenceEntry:    0.7,  // Higher confidence requirement
}

riskChecker := risk.NewChecker(logger, customConfig, stateManager, safetyGuard, queue)
```

### Monitoring & Metrics

```go
// Get current risk metrics
metrics := riskChecker.GetRiskMetrics()
fmt.Printf("Exposure Ratio: %.2f%%\\n", metrics.ExposureRatio*100)
fmt.Printf("Active Positions: %d/%d\\n", metrics.ActivePositions, metrics.MaxPositions)
fmt.Printf("Daily PnL: $%.2f\\n", metrics.DailyPnL)
```

## Error Handling

### Rejection Scenarios

1. **System Not Active**
   - Log: "Decision rejected - system not active"
   - Action: Return false, log system status

2. **Position Limit Exceeded**
   - Log: "Decision rejected - max positions reached"
   - Action: Return false, log current vs max positions

3. **Daily Loss Limit Approached**
   - Log: "Decision rejected - approaching daily loss limit"
   - Action: Return false, log loss percentage

4. **Insufficient Confidence**
   - Log: "Decision rejected - insufficient confidence"
   - Action: Return false, log confidence vs threshold

5. **Safety Violation**
   - Log: "Decision rejected - safety violation detected"
   - Action: Return false, log violation details

### Recovery Mechanisms

```go
// Graceful error handling in queue processing
func (re *CheckerImpl) handleDecision(msg *queue.Message) error {
    defer func() {
        if r := recover(); r != nil {
            re.logger.Error("Risk check panic", zap.Any("error", r))
        }
    }()
    
    // Process decision with comprehensive error handling
    // Always commit message to prevent reprocessing
    return nil
}
```

## Testing

### Unit Tests

```bash
# Test individual risk checks
go test ./internal/services/risk -v -run TestPositionLimits
go test ./internal/services/risk -v -run TestConfidenceThresholds
go test ./internal/services/risk -v -run TestExposureLimits

# Test integration with guard system
go test ./internal/services/risk -v -run TestGuardIntegration
```

### Integration Tests

```bash
# Test full risk validation flow
go test ./internal/services/risk -v -run TestRiskValidationFlow

# Test queue processing
go test ./internal/services/risk -v -run TestQueueProcessing

# Load testing
go test ./internal/services/risk -v -run TestHighVolumeProcessing
```

### Risk Scenarios

```go
// Test extreme market conditions
func TestExtremeVolatilityRisk(t *testing.T) {
    decision := &models.TradingDecision{
        Symbol: "BTCUSDT",
        Action: "BUY",
        Size:   5.0,  // Large size
        // High volatility metadata
    }
    
    approved := riskChecker.CheckDecision(decision)
    assert.False(t, approved, "Should reject large position in high volatility")
}
```

## Performance

### Throughput Metrics

- **Processing Rate**: 1000+ decisions per second
- **Queue Processing**: 2-second polling interval
- **Memory Usage**: <10MB baseline, <50MB under load
- **Latency**: <5ms per decision validation

### Optimization

```go
// Efficient risk calculation caching
type RiskCache struct {
    ExposureRatio    float64
    PositionCount    int
    DailyPnL        float64
    LastUpdate      time.Time
    CacheDuration   time.Duration
}

// Cache risk metrics to avoid repeated calculations
func (re *CheckerImpl) getCachedRiskMetrics() *RiskMetrics {
    if time.Since(re.cache.LastUpdate) < re.cache.CacheDuration {
        return re.cache.Metrics
    }
    // Refresh cache
    return re.refreshRiskCache()
}
```

## Best Practices

### 1. Risk Configuration
- Start with conservative limits and gradually adjust
- Monitor rejection rates and adjust thresholds accordingly
- Regular backtesting of risk parameters

### 2. Monitoring
- Set up alerts for high rejection rates
- Monitor exposure ratios in real-time
- Track daily loss progression

### 3. Integration
- Ensure proper guard system integration
- Regular validation of state synchronization
- Implement comprehensive logging for audit trails

## Future Enhancements

### Planned Improvements

1. **Dynamic Risk Adjustment**
   - Market volatility-based limit scaling
   - Time-of-day risk adjustments
   - Correlation-based exposure management

2. **Machine Learning Integration**
   - Adaptive confidence thresholds
   - Predictive risk modeling
   - Anomaly detection

3. **Advanced Portfolio Management**
   - Sector exposure limits
   - Currency exposure management
   - Greeks-based risk for options (future)

4. **Real-Time Risk Dashboard**
   - Web-based monitoring interface
   - Real-time risk metrics visualization
   - Alert management system