# Guard & Safety System

## Overview

This document explains the comprehensive safety and circuit breaker system that provides continuous monitoring and emergency protection for the trading system. The guard system acts as the final safety layer, capable of immediately terminating operations when critical conditions are detected.

## Architecture

### Core Components

1. **Safety Rules Engine**: Configurable rules for detecting dangerous conditions
2. **Circuit Breaker System**: Automatic trading halts with cooldown periods
3. **Executor Control Interface**: Direct control over trading execution
4. **Emergency Notification System**: Real-time alerts for critical events

### Data Flow

```
Continuous Monitoring → Safety Rules → Violation Detection → Circuit Breakers → Executor Control
                                           ↓
                                  Emergency Notifications
```

## Key Features

### 1. Multi-Rule Safety Framework

The guard system implements comprehensive safety rules:

- **Daily Loss Limit Rule**: Prevents catastrophic daily losses
- **Maximum Positions Rule**: Controls position count limits  
- **Drawdown Limit Rule**: Monitors portfolio drawdown
- **Account Balance Rule**: Protects against account depletion
- **Position Size Rule**: Validates individual position sizes
- **Market Volatility Rule**: Detects extreme market conditions
- **Connection Health Rule**: Monitors system connectivity
- **System Status Rule**: Ensures system state consistency

### 2. Circuit Breaker Management

- **Automatic Triggering**: Rules automatically trigger appropriate breakers
- **Cooldown Periods**: Prevents rapid re-triggering
- **Escalation Levels**: Progressive response to repeated violations
- **Manual Override**: Administrative control over breaker states

### 3. Executor Integration

- **Direct Termination**: Immediate shutdown capability
- **Position Closure**: Emergency liquidation of all positions
- **Trading Pause**: Temporary halt of new position opening
- **Service Coordination**: Integration with all trading services

### 4. Real-Time Monitoring

- **30-Second Monitoring Cycle**: Continuous safety validation
- **Event-Driven Responses**: Immediate reaction to violations
- **Comprehensive Logging**: Detailed audit trail of all safety events
- **Performance Optimized**: Minimal overhead on trading operations

## Safety Rules

### 1. Daily Loss Limit Rule

```go
type DailyLossLimitRule struct {
    Priority int // 100 (Critical)
}
```

**Violation Thresholds:**
- **80% of limit**: HIGH severity → Pause Trading
- **100% of limit**: CRITICAL severity → Emergency Stop

**Response Actions:**
- Immediate position evaluation
- Risk limit tightening
- Emergency position closure if needed

### 2. Maximum Positions Rule

```go
type MaxPositionsRule struct {
    Priority int // 80 (High)
}
```

**Monitoring:**
- Active position count tracking
- 90% utilization warnings
- 100% utilization trading halt

**Escalation:**
- MEDIUM severity at 90% → Warning
- HIGH severity at 100% → Pause Trading

### 3. Drawdown Limit Rule

```go
type DrawdownLimitRule struct {
    Priority     int     // 90 (Critical)
    MaxDrawdown  float64 // 20% maximum drawdown
}
```

**Drawdown Monitoring:**
- Real-time drawdown calculation
- Peak-to-trough measurement
- Historical drawdown tracking

**Response Levels:**
- **16% drawdown**: HIGH severity → Pause Trading  
- **20% drawdown**: CRITICAL severity → Emergency Stop

### 4. Account Balance Rule

```go
type AccountBalanceRule struct {
    Priority        int     // 95 (Critical)
    MinimumBalance  float64 // $100 minimum balance
}
```

**Balance Protection:**
- Estimated balance monitoring
- Critical balance alerts
- Account depletion prevention

**Thresholds:**
- **3x minimum**: HIGH severity → Pause Trading
- **1x minimum**: CRITICAL severity → Emergency Stop

### 5. Position Size Rule

```go
type PositionSizeRule struct {
    Priority              int     // 70 (Medium-High)
    MaxSinglePosition     float64 // $5,000 per position
    MaxTotalExposure      float64 // $15,000 total exposure
}
```

**Size Monitoring:**
- Individual position value tracking
- Total portfolio exposure calculation
- Leverage-adjusted position values

**Violation Handling:**
- Oversized single positions → MEDIUM severity → Warning
- Excessive total exposure → HIGH severity → Pause Trading

### 6. Market Volatility Rule

```go
type MarketVolatilityRule struct {
    Priority              int     // 85 (High)
    VolatilityThreshold   float64 // 5% price movement
}
```

**Volatility Detection:**
- Real-time position P&L monitoring
- Price movement analysis
- Market stress indicators

**Response:**
- >50% positions showing 5%+ moves → HIGH severity → Pause Trading
- Automatic position size reductions
- Enhanced monitoring activation

### 7. Connection Health Rule

```go
type ConnectionHealthRule struct {
    Priority         int           // 75 (Medium-High)
    MaxStaleTime     time.Duration // 5 minutes
    MaxPendingTime   time.Duration // 30 minutes
}
```

**Health Monitoring:**
- Data freshness validation
- Stuck order detection
- System responsiveness checks

**Escalation:**
- **5-15 minutes stale**: MEDIUM severity → Warning
- **>15 minutes stale**: HIGH severity → Pause Trading
- **Stuck orders detected**: MEDIUM severity → Warning

### 8. System Status Rule

```go
type SystemStatusRule struct {
    Priority int // 60 (Medium)
}
```

**Status Consistency:**
- Emergency status duration monitoring
- Active positions vs system status validation
- Status transition validation

**Validation:**
- Emergency status >24 hours → Warning
- Active positions while paused → Warning
- Invalid status transitions → Warning

## Circuit Breakers

### Circuit Breaker Configuration

```go
type CircuitBreaker struct {
    Name             string
    Status           CircuitBreakerStatus
    TriggerCount     int
    CooldownDuration time.Duration
    MaxTriggers      int
    IsEnabled        bool
}
```

### Default Circuit Breakers

#### 1. Daily Loss Limit Breaker
- **Cooldown**: 1 hour
- **Max Triggers**: 3
- **Action**: Emergency stop after 3rd trigger

#### 2. Max Positions Breaker  
- **Cooldown**: 30 minutes
- **Max Triggers**: 5
- **Action**: Pause trading, allow position closure only

#### 3. Drawdown Limit Breaker
- **Cooldown**: 2 hours
- **Max Triggers**: 2
- **Action**: Emergency stop after 2nd trigger

#### 4. Emergency Stop Breaker
- **Cooldown**: 24 hours
- **Max Triggers**: 1
- **Action**: Complete system shutdown

#### 5. Pause Trading Breaker
- **Cooldown**: 15 minutes
- **Max Triggers**: 10
- **Action**: Temporary halt, auto-resume after cooldown

#### 6. Emergency Close Breaker
- **Cooldown**: 4 hours
- **Max Triggers**: 3
- **Action**: Close all positions immediately

### Circuit Breaker States

```go
type CircuitBreakerStatus string

const (
    StatusNormal    CircuitBreakerStatus = "NORMAL"
    StatusTriggered CircuitBreakerStatus = "TRIGGERED"
    StatusCooldown  CircuitBreakerStatus = "COOLDOWN"
    StatusDisabled  CircuitBreakerStatus = "DISABLED"
)
```

**State Transitions:**
1. **NORMAL** → **TRIGGERED** (on violation)
2. **TRIGGERED** → **COOLDOWN** (after cooldown period)
3. **COOLDOWN** → **NORMAL** (after additional cooldown)
4. **Any State** → **DISABLED** (manual override or max triggers)

## Executor Control Interface

### ExecutorController Interface

```go
type ExecutorController interface {
    Terminate() error        // Complete shutdown
    Pause() error           // Pause new positions
    Resume() error          // Resume operations
    CloseAllPositions() error // Emergency liquidation
}
```

### Action Mapping

| Violation Action | Executor Response | System Status |
|-----------------|-------------------|---------------|
| `ActionWarn` | Log only | No change |
| `ActionPauseTrading` | `Pause()` | PAUSED |
| `ActionClosePositions` | `CloseAllPositions()` | EMERGENCY |
| `ActionEmergencyStop` | `Terminate()` | EMERGENCY |

### Emergency Protocols

#### 1. Pause Trading Protocol
```go
func (sg *SafetyGuard) executePauseTrading(violation *SafetyViolation) {
    // Update system status
    sg.stateManager.SetSystemStatus(state.SystemStatusPaused)
    
    // Trigger circuit breaker
    sg.triggerCircuitBreaker("pause_trading", violation)
    
    // Pause executor
    if sg.executorCtrl != nil {
        sg.executorCtrl.Pause()
    }
}
```

#### 2. Emergency Close Protocol
```go
func (sg *SafetyGuard) executeEmergencyClose(violation *SafetyViolation) {
    // Set emergency status
    sg.stateManager.SetSystemStatus(state.SystemStatusEmergency)
    
    // Close all positions
    if sg.executorCtrl != nil {
        sg.executorCtrl.CloseAllPositions()
    }
    
    // Send emergency notifications
    sg.sendEmergencyNotification("emergency_close", violation)
}
```

#### 3. Emergency Stop Protocol
```go
func (sg *SafetyGuard) executeEmergencyStop(violation *SafetyViolation) {
    // Set emergency status
    sg.stateManager.SetSystemStatus(state.SystemStatusEmergency)
    
    // Terminate all operations
    if sg.executorCtrl != nil {
        sg.executorCtrl.Terminate()
    }
    
    // Send critical alerts
    sg.sendCriticalAlert("emergency_stop", violation)
}
```

## Usage

### Basic Setup

```go
// Initialize safety guard
logger := logger.New()
stateManager := state.NewStateManager(logger, db)
safetyGuard := guard.New(logger, stateManager)

// Set executor control
executorController := &MyExecutorController{}
safetyGuard.SetExecutorController(executorController)

// Set emergency notification queue
emergencyQueue := queue.NewQueue()
safetyGuard.SetQueue(emergencyQueue)

// Start monitoring
if err := safetyGuard.Start(); err != nil {
    log.Fatal("Failed to start safety guard:", err)
}
```

### Custom Rules

```go
// Add custom safety rule
type CustomVolatilityRule struct{}

func (r *CustomVolatilityRule) Name() string { return "custom_volatility" }
func (r *CustomVolatilityRule) Priority() int { return 90 }

func (r *CustomVolatilityRule) Check(state *state.TradingState) *guard.SafetyViolation {
    // Custom volatility logic
    if volatilityScore > threshold {
        return &guard.SafetyViolation{
            RuleName:  r.Name(),
            Severity:  guard.SeverityHigh,
            Message:   "Custom volatility threshold exceeded",
            Action:    guard.ActionPauseTrading,
            Timestamp: time.Now(),
        }
    }
    return nil
}

// Register custom rule
safetyGuard.AddRule(&CustomVolatilityRule{})
```

### Event Handling

```go
// Implement safety event handler
type SafetyEventHandler struct {
    logger *logger.Logger
}

func (h *SafetyEventHandler) OnViolationDetected(violation *guard.SafetyViolation) {
    h.logger.Warn("Safety violation", zap.String("rule", violation.RuleName))
}

func (h *SafetyEventHandler) OnCircuitBreakerTriggered(name string, violation *guard.SafetyViolation) {
    h.logger.Error("Circuit breaker triggered", zap.String("breaker", name))
}

func (h *SafetyEventHandler) OnCircuitBreakerReset(name string) {
    h.logger.Info("Circuit breaker reset", zap.String("breaker", name))
}

// Register event handler
safetyGuard.AddListener(&SafetyEventHandler{logger: logger})
```

### Manual Controls

```go
// Manual circuit breaker control
safetyGuard.EnableCircuitBreaker("emergency_stop")
safetyGuard.DisableCircuitBreaker("daily_loss_limit")
safetyGuard.ResetCircuitBreaker("pause_trading")

// Get system status
status := safetyGuard.GetStatus()
for name, breaker := range status {
    fmt.Printf("Breaker %s: %s\\n", name, breaker.Status)
}

// Get violation history
violations := safetyGuard.GetViolationHistory()
fmt.Printf("Recent violations: %d\\n", len(violations))
```

## Configuration

### Safety Parameters

```go
type SafetyConfig struct {
    // Monitoring
    CheckInterval       time.Duration // 30 seconds
    MaxViolationHistory int          // 1000 violations
    
    // Daily loss limits
    DailyLossLimit      float64      // $500
    DailyLossWarning    float64      // 80% of limit
    
    // Position limits
    MaxPositions        int          // 5 positions
    MaxPositionSize     float64      // $5,000
    MaxTotalExposure    float64      // $15,000
    
    // Drawdown limits
    MaxDrawdown         float64      // 20%
    DrawdownWarning     float64      // 16%
    
    // Volatility thresholds
    VolatilityThreshold float64      // 5% price movement
    VolatilityWindow    time.Duration // 10 minutes
}
```

### Circuit Breaker Tuning

```go
// Conservative configuration
conservativeConfig := map[string]*guard.CircuitBreaker{
    "daily_loss_limit": {
        CooldownDuration: time.Hour * 2,    // Longer cooldown
        MaxTriggers:      2,                // Fewer triggers allowed
    },
    "max_positions": {
        CooldownDuration: time.Hour,        // Longer cooldown
        MaxTriggers:      3,                // Fewer triggers
    },
}

// Apply configuration
for name, config := range conservativeConfig {
    safetyGuard.UpdateCircuitBreaker(name, config)
}
```

## Monitoring & Alerts

### Safety Metrics

```go
type SafetyMetrics struct {
    ViolationsToday     int                    `json:"violations_today"`
    ActiveBreakers      []string               `json:"active_breakers"`
    SystemStatus        string                 `json:"system_status"`
    LastViolation       time.Time              `json:"last_violation"`
    RuleEffectiveness   map[string]int         `json:"rule_effectiveness"`
    BreakerTriggerCount map[string]int         `json:"breaker_trigger_count"`
}
```

### Alert Levels

| Severity | Response Time | Notification |
|----------|---------------|--------------|
| LOW | 5 minutes | Log entry |
| MEDIUM | 1 minute | Email alert |
| HIGH | 30 seconds | SMS + Email |
| CRITICAL | Immediate | Phone call + SMS + Email |

### Dashboard Integration

```go
// Safety dashboard endpoint
func (sg *SafetyGuard) GetDashboardData() *DashboardData {
    return &DashboardData{
        SystemStatus:      sg.stateManager.GetSystemStatus(),
        CircuitBreakers:   sg.GetStatus(),
        RecentViolations:  sg.GetViolationHistory()[0:10],
        SafetyMetrics:     sg.calculateSafetyMetrics(),
        RealTimeStatus:    sg.getRealTimeStatus(),
    }
}
```

## Testing

### Safety Rule Testing

```bash
# Test individual safety rules
go test ./internal/services/guard -v -run TestDailyLossLimitRule
go test ./internal/services/guard -v -run TestMaxPositionsRule
go test ./internal/services/guard -v -run TestDrawdownLimitRule

# Test circuit breaker logic
go test ./internal/services/guard -v -run TestCircuitBreakerTrigger
go test ./internal/services/guard -v -run TestCircuitBreakerCooldown
```

### Stress Testing

```go
// Test emergency scenarios
func TestEmergencyStopScenario(t *testing.T) {
    // Simulate critical loss scenario
    state := &state.TradingState{
        DailyPnL: -1000.0, // Exceeds daily loss limit
        RiskMetrics: &state.RiskMetrics{
            DailyLossLimit: 500.0,
        },
    }
    
    violation := rule.Check(state)
    assert.Equal(t, guard.ActionEmergencyStop, violation.Action)
}
```

### Integration Testing

```bash
# Test full safety system integration
go test ./internal/services/guard -v -run TestFullSafetyIntegration

# Test executor control integration  
go test ./internal/services/guard -v -run TestExecutorControlIntegration

# Load testing
go test ./internal/services/guard -v -run TestHighFrequencyViolations
```

## Performance

### Monitoring Overhead

- **CPU Usage**: <1% under normal conditions
- **Memory Usage**: <5MB baseline, <20MB under stress
- **Check Latency**: <1ms per safety rule evaluation
- **Response Time**: <100ms from violation to action

### Optimization

```go
// Efficient rule evaluation
func (sg *SafetyGuard) optimizedRuleCheck() {
    // Skip expensive rules if system is paused
    if sg.stateManager.GetSystemStatus() != state.SystemStatusActive {
        return
    }
    
    // Parallel rule evaluation for independent rules
    violations := make(chan *SafetyViolation, len(sg.rules))
    
    for _, rule := range sg.rules {
        go func(r SafetyRule) {
            if violation := r.Check(state); violation != nil {
                violations <- violation
            }
        }(rule)
    }
}
```

## Best Practices

### 1. Rule Design
- Keep rules simple and focused
- Avoid overlapping rule responsibilities
- Test rules with historical data

### 2. Circuit Breaker Tuning
- Start with conservative settings
- Monitor false positive rates
- Adjust based on market conditions

### 3. Emergency Procedures
- Document all emergency protocols
- Regular emergency drill testing
- Clear escalation procedures

### 4. Monitoring
- 24/7 monitoring setup
- Automated alert systems
- Regular safety metric reviews

## Future Enhancements

### Planned Improvements

1. **Machine Learning Integration**
   - Adaptive threshold adjustment
   - Anomaly detection rules
   - Predictive safety modeling

2. **Advanced Risk Models**
   - Real-time VaR calculation
   - Stress testing integration
   - Scenario-based safety rules

3. **Enhanced Monitoring**
   - Web-based safety dashboard
   - Real-time safety metrics
   - Mobile alert applications

4. **Integration Improvements**
   - Exchange API health monitoring
   - Network connectivity rules
   - Third-party data validation

### Experimental Features

- **Quantum-resistant safety protocols**
- **Blockchain-based audit trails**  
- **AI-powered safety rule generation**
- **Cross-exchange safety coordination**