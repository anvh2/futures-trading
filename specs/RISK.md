# Risk Management Service Specification

## Overview
The Risk Management Service provides comprehensive portfolio risk assessment, position sizing calculations, exposure monitoring, and real-time risk controls for the futures trading system. It implements sophisticated algorithms for capital preservation and optimal risk-adjusted returns.

## Architecture
- **Package**: `internal/services/risk`
- **Type**: Real-time monitoring service with calculation utilities
- **Dependencies**: Portfolio data, market data, position information, settings
- **Design Pattern**: Rule-based system with configurable parameters

## Core Components

### 1. Risk Configuration
```go
type Config struct {
    // Position sizing limits
    MaxPositionPercent  float64  // Max % of account per trade (default: 5.0%)
    MaxDailyRisk        float64  // Max daily risk exposure (default: 15.0%)
    MaxPortfolioRisk    float64  // Max total portfolio risk (default: 25.0%)
    
    // Volatility-based adjustments
    StopATRMultiplier   float64  // Stop loss ATR multiplier (default: 1.5)
    TargetATRMultiplier float64  // Take profit ATR multiplier (default: 2.5)
    
    // Minimum risk buffers
    MinSLBuffer         float64  // Min stop loss buffer % (default: 0.2%)
    MinTPBuffer         float64  // Min take profit buffer % (default: 0.4%)
    
    // Leverage management
    MaxLeverage         int      // Maximum allowed leverage (default: 10)
    VolatilityThreshold float64  // ATR threshold for leverage reduction
    
    // Correlation limits
    MaxCorrelatedRisk   float64  // Max risk in correlated positions
    CorrelationLimit    float64  // Correlation coefficient threshold
}

func DefaultConfig() Config {
    return Config{
        MaxPositionPercent:  5.0,
        MaxDailyRisk:       15.0,
        MaxPortfolioRisk:   25.0,
        StopATRMultiplier:  1.5,
        TargetATRMultiplier: 2.5,
        MinSLBuffer:        0.002,
        MinTPBuffer:        0.004,
        MaxLeverage:        10,
        VolatilityThreshold: 3.0,
        MaxCorrelatedRisk:  10.0,
        CorrelationLimit:   0.7,
    }
}
```

### 2. Risk Assessment Models
```go
type RiskAssessment struct {
    Symbol              string
    CurrentPrice        float64
    ATRPercent          float64
    Confidence          int
    
    // Position metrics
    RecommendedSize     float64  // Optimal position size
    MaxAllowedSize      float64  // Maximum permitted size
    RecommendedLeverage int      // Suggested leverage
    MaxAllowedLeverage  int      // Maximum permitted leverage
    
    // Risk metrics
    EstimatedRisk       float64  // Expected risk (stop loss)
    MaxDrawdown         float64  // Potential maximum loss
    RiskRewardRatio     float64  // Expected risk-reward
    
    // Position limits
    StopLoss            float64  // Calculated stop loss level
    TakeProfit          float64  // Calculated take profit level
    
    // Portfolio impact
    PortfolioRiskBefore float64  // Risk before position
    PortfolioRiskAfter  float64  // Risk after position
    
    // Compliance
    WithinLimits        bool     // All limits satisfied
    LimitViolations     []string // List of violated limits
}
```

## Core Risk Functions

### 1. Position Sizing (`PositionSizePercent`)
**Purpose**: Calculate optimal position size based on confidence, volatility, and risk limits

```go
func PositionSizePercent(confidence int, atrPercent float64, cfg Config) float64 {
    if confidence < 60 {
        return 0 // Below minimum confidence threshold
    }
    
    // Base size calculation (60% = 2%, 100% = MaxPositionPercent)
    baseSize := float64(confidence-60)/40.0 * cfg.MaxPositionPercent
    
    // Volatility adjustment (higher volatility = smaller size)
    volAdjustment := 1.0
    switch {
    case atrPercent > 5.0:
        volAdjustment = 0.3 // Reduce to 30% for extreme volatility
    case atrPercent > 3.0:
        volAdjustment = 0.5 // Reduce to 50% for high volatility
    case atrPercent > 2.0:
        volAdjustment = 0.7 // Reduce to 70% for medium volatility
    }
    
    // Market condition adjustment
    marketAdjustment := getMarketRegimeAdjustment()
    
    finalSize := baseSize * volAdjustment * marketAdjustment
    return math.Min(finalSize, cfg.MaxPositionPercent)
}
```

### 2. Leverage Recommendation (`RecommendLeverage`)
**Purpose**: Determine appropriate leverage based on confidence and market volatility

```go
func RecommendLeverage(atrPercent float64, confidence int, cfg Config) int {
    // Base leverage from confidence
    var baseLeverage int
    switch {
    case confidence >= 90:
        baseLeverage = 8
    case confidence >= 85:
        baseLeverage = 6
    case confidence >= 80:
        baseLeverage = 5
    case confidence >= 75:
        baseLeverage = 4
    case confidence >= 70:
        baseLeverage = 3
    default:
        baseLeverage = 2
    }
    
    // Volatility adjustment
    volAdjustedLeverage := baseLeverage
    switch {
    case atrPercent >= 5.0:
        volAdjustedLeverage = 2 // Maximum 2x for extreme volatility
    case atrPercent >= 3.0:
        volAdjustedLeverage = min(baseLeverage, 3)
    case atrPercent >= 2.0:
        volAdjustedLeverage = min(baseLeverage, 5)
    case atrPercent <= 1.0:
        volAdjustedLeverage = min(baseLeverage+2, cfg.MaxLeverage) // Bonus for low vol
    }
    
    return min(volAdjustedLeverage, cfg.MaxLeverage)
}
```

### 3. Stop Loss Calculation (`CalculateStopLoss`)
**Purpose**: Determine optimal stop loss level using ATR and market structure

```go
func CalculateStopLoss(entryPrice, atrValue float64, isLong bool, cfg Config) float64 {
    // ATR-based stop distance
    atrStop := atrValue * cfg.StopATRMultiplier
    
    // Minimum buffer based on price
    minBuffer := entryPrice * cfg.MinSLBuffer
    
    // Use larger of ATR stop or minimum buffer
    stopDistance := math.Max(atrStop, minBuffer)
    
    // Apply market structure adjustment
    structureAdjustment := getMarketStructureAdjustment(entryPrice, isLong)
    finalStopDistance := stopDistance * structureAdjustment
    
    if isLong {
        return entryPrice - finalStopDistance
    }
    return entryPrice + finalStopDistance
}
```

### 4. Take Profit Calculation (`CalculateTakeProfit`)
**Purpose**: Set profit targets based on risk-reward ratio and market conditions

```go
func CalculateTakeProfit(entryPrice, stopLoss float64, isLong bool, cfg Config) []float64 {
    riskAmount := math.Abs(entryPrice - stopLoss)
    
    // Multiple take profit levels
    targets := []float64{}
    
    // Conservative target (1:1 risk-reward)
    conservativeReward := riskAmount * 1.0
    if isLong {
        targets = append(targets, entryPrice+conservativeReward)
    } else {
        targets = append(targets, entryPrice-conservativeReward)
    }
    
    // Standard target (1:2 risk-reward)
    standardReward := riskAmount * 2.0
    if isLong {
        targets = append(targets, entryPrice+standardReward)
    } else {
        targets = append(targets, entryPrice-standardReward)
    }
    
    // Aggressive target (1:3 risk-reward)
    aggressiveReward := riskAmount * 3.0
    if isLong {
        targets = append(targets, entryPrice+aggressiveReward)
    } else {
        targets = append(targets, entryPrice-aggressiveReward)
    }
    
    return targets
}
```

## Portfolio Risk Management

### 1. Exposure Calculation
```go
type PortfolioExposure struct {
    TotalEquity         float64
    UsedMargin          float64
    FreeMargin          float64
    TotalPositionValue  float64
    
    // Risk metrics
    TotalRiskExposure   float64  // Sum of all position risks
    DailyRiskUsed       float64  // Risk used today
    MaxDrawdownRisk     float64  // Potential maximum loss
    
    // By asset/sector
    ExposureByAsset     map[string]float64
    ExposureBySector    map[string]float64
    
    // Correlation metrics
    CorrelationMatrix   map[string]map[string]float64
    CorrelatedRisk      float64  // Risk in correlated positions
    
    // Limits status
    WithinLimits        bool
    LimitViolations     []string
}

func CalculatePortfolioExposure(positions []Position, cfg Config) *PortfolioExposure {
    exposure := &PortfolioExposure{
        ExposureByAsset:  make(map[string]float64),
        ExposureBySector: make(map[string]float64),
        CorrelationMatrix: make(map[string]map[string]float64),
    }
    
    // Calculate individual position exposures
    for _, position := range positions {
        positionValue := position.Quantity * position.EntryPrice
        riskAmount := math.Abs(position.EntryPrice - position.StopLoss) * position.Quantity
        
        exposure.TotalPositionValue += positionValue
        exposure.TotalRiskExposure += riskAmount
        
        // Asset allocation
        exposure.ExposureByAsset[position.Symbol] = positionValue
        
        // Sector allocation
        sector := getAssetSector(position.Symbol)
        exposure.ExposureBySector[sector] += positionValue
    }
    
    // Calculate correlation-adjusted risk
    exposure.CorrelatedRisk = calculateCorrelatedRisk(positions, cfg)
    
    // Check compliance
    exposure.WithinLimits, exposure.LimitViolations = checkExposureLimits(exposure, cfg)
    
    return exposure
}
```

### 2. Correlation Analysis
```go
func calculateCorrelatedRisk(positions []Position, cfg Config) float64 {
    correlatedRisk := 0.0
    
    for i := 0; i < len(positions); i++ {
        for j := i + 1; j < len(positions); j++ {
            // Get correlation coefficient between assets
            correlation := getAssetCorrelation(positions[i].Symbol, positions[j].Symbol)
            
            if math.Abs(correlation) > cfg.CorrelationLimit {
                // Calculate combined risk for correlated positions
                risk1 := calculatePositionRisk(positions[i])
                risk2 := calculatePositionRisk(positions[j])
                
                // Correlation-adjusted combined risk
                combinedRisk := math.Sqrt(
                    risk1*risk1 + risk2*risk2 + 2*correlation*risk1*risk2,
                )
                
                correlatedRisk += combinedRisk
            }
        }
    }
    
    return correlatedRisk
}
```

## Real-time Risk Monitoring

### 1. Risk Alerts System
```go
type RiskAlert struct {
    Level       AlertLevel  // INFO, WARNING, CRITICAL
    Type        AlertType   // EXPOSURE, CORRELATION, DRAWDOWN, MARGIN
    Symbol      string
    Message     string
    Value       float64
    Threshold   float64
    Timestamp   time.Time
    Action      string      // Required action
}

type RiskMonitor struct {
    cfg         Config
    alerts      chan RiskAlert
    positions   map[string]Position
    metrics     *RiskMetrics
    lastCheck   time.Time
}

func (rm *RiskMonitor) MonitorRisk() {
    ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
    
    for {
        select {
        case <-ticker.C:
            rm.performRiskChecks()
        case <-rm.quitChannel:
            return
        }
    }
}

func (rm *RiskMonitor) performRiskChecks() {
    // 1. Position size limits
    rm.checkPositionSizeLimits()
    
    // 2. Portfolio exposure limits
    rm.checkPortfolioExposure()
    
    // 3. Daily risk usage
    rm.checkDailyRiskUsage()
    
    // 4. Margin requirements
    rm.checkMarginRequirements()
    
    // 5. Correlation limits
    rm.checkCorrelationLimits()
    
    // 6. Drawdown limits
    rm.checkDrawdownLimits()
}
```

### 2. Automatic Risk Controls
```go
type RiskControl struct {
    Type        ControlType  // POSITION_REDUCE, POSITION_CLOSE, TRADING_HALT
    Trigger     float64      // Threshold that triggered control
    Action      string       // Description of action taken
    Timestamp   time.Time
}

func (rm *RiskMonitor) executeRiskControls() {
    exposure := rm.calculateCurrentExposure()
    
    // Critical exposure level - reduce positions
    if exposure.TotalRiskExposure > rm.cfg.MaxPortfolioRisk * 0.9 {
        rm.reducePositions(0.5) // Reduce all positions by 50%
        
        rm.alerts <- RiskAlert{
            Level:   CRITICAL,
            Type:    EXPOSURE,
            Message: "Portfolio risk limit approached - positions reduced",
            Value:   exposure.TotalRiskExposure,
            Threshold: rm.cfg.MaxPortfolioRisk,
        }
    }
    
    // Extreme exposure level - halt trading
    if exposure.TotalRiskExposure > rm.cfg.MaxPortfolioRisk {
        rm.haltTrading()
        rm.closeRiskiestPositions(3) // Close 3 riskiest positions
        
        rm.alerts <- RiskAlert{
            Level:   CRITICAL,
            Type:    EXPOSURE,
            Message: "Portfolio risk limit exceeded - trading halted",
            Value:   exposure.TotalRiskExposure,
            Threshold: rm.cfg.MaxPortfolioRisk,
            Action:  "TRADING_HALTED",
        }
    }
}
```

## Advanced Risk Metrics

### 1. Value at Risk (VaR) Calculation
```go
func CalculateVaR(positions []Position, confidence float64, timeHorizon int) float64 {
    // Historical simulation method
    returns := calculateHistoricalReturns(positions, timeHorizon*4) // 4x time horizon for data
    
    // Sort returns to find percentile
    sort.Float64s(returns)
    
    // Find VaR at specified confidence level
    percentileIndex := int(float64(len(returns)) * (1.0 - confidence))
    if percentileIndex >= len(returns) {
        percentileIndex = len(returns) - 1
    }
    
    return math.Abs(returns[percentileIndex])
}
```

### 2. Expected Shortfall (Conditional VaR)
```go
func CalculateExpectedShortfall(positions []Position, confidence float64, timeHorizon int) float64 {
    var95 := CalculateVaR(positions, confidence, timeHorizon)
    
    // Calculate average loss beyond VaR
    returns := calculateHistoricalReturns(positions, timeHorizon*4)
    sort.Float64s(returns)
    
    var sumBeyondVaR float64
    var countBeyondVaR int
    
    for _, ret := range returns {
        if math.Abs(ret) > var95 {
            sumBeyondVaR += math.Abs(ret)
            countBeyondVaR++
        }
    }
    
    if countBeyondVaR == 0 {
        return var95
    }
    
    return sumBeyondVaR / float64(countBeyondVaR)
}
```

### 3. Sharpe Ratio Calculation
```go
func CalculateSharpeRatio(returns []float64, riskFreeRate float64) float64 {
    if len(returns) < 2 {
        return 0
    }
    
    // Calculate excess returns
    var excessReturns []float64
    for _, ret := range returns {
        excessReturns = append(excessReturns, ret-riskFreeRate)
    }
    
    // Calculate mean and standard deviation
    mean := calculateMean(excessReturns)
    stdDev := calculateStdDev(excessReturns, mean)
    
    if stdDev == 0 {
        return 0
    }
    
    return mean / stdDev
}
```

## Risk-Adjusted Performance

### 1. Performance Attribution
```go
type PerformanceAttribution struct {
    TotalReturn          float64
    RiskAdjustedReturn   float64
    MaxDrawdown          float64
    SharpeRatio          float64
    SortinoRatio         float64
    CalmarRatio          float64
    
    // Attribution by factor
    AlphaReturn          float64  // Skill-based return
    BetaReturn           float64  // Market-based return
    
    // Risk contribution
    VolatilityContrib    float64
    CorrelationContrib   float64
    ConcentrationContrib float64
}

func CalculatePerformanceAttribution(trades []Trade, benchmark []float64) *PerformanceAttribution {
    attribution := &PerformanceAttribution{}
    
    // Calculate returns
    returns := calculateTradeReturns(trades)
    attribution.TotalReturn = calculateCumulativeReturn(returns)
    
    // Risk metrics
    attribution.MaxDrawdown = calculateMaxDrawdown(returns)
    attribution.SharpeRatio = CalculateSharpeRatio(returns, 0.02) // 2% risk-free rate
    
    // Beta calculation against benchmark
    beta, alpha := calculateAlphaBeta(returns, benchmark)
    attribution.AlphaReturn = alpha * float64(len(returns)) // Annualized alpha
    attribution.BetaReturn = attribution.TotalReturn - attribution.AlphaReturn
    
    return attribution
}
```

### 2. Risk Budgeting
```go
type RiskBudget struct {
    TotalRiskBudget    float64
    AllocatedRisk      map[string]float64  // Symbol -> risk allocation
    UsedRisk           map[string]float64  // Symbol -> actual risk usage
    AvailableRisk      float64
    
    // Sector/strategy allocation
    SectorRiskBudget   map[string]float64
    StrategyRiskBudget map[string]float64
}

func AllocateRiskBudget(totalCapital float64, maxRiskPercent float64, symbols []string) *RiskBudget {
    budget := &RiskBudget{
        TotalRiskBudget:    totalCapital * maxRiskPercent / 100,
        AllocatedRisk:      make(map[string]float64),
        UsedRisk:           make(map[string]float64),
        SectorRiskBudget:   make(map[string]float64),
        StrategyRiskBudget: make(map[string]float64),
    }
    
    // Equal allocation to start (can be optimized)
    riskPerSymbol := budget.TotalRiskBudget / float64(len(symbols))
    
    for _, symbol := range symbols {
        budget.AllocatedRisk[symbol] = riskPerSymbol
    }
    
    budget.AvailableRisk = budget.TotalRiskBudget
    
    return budget
}
```

## Integration Points

### 1. Input Sources
- **Order Service**: Position information and trade requests
- **Market Service**: Current prices and volatility data
- **Portfolio Service**: Account equity and margin information
- **Settings Service**: Risk parameters and limits configuration

### 2. Output Consumers
- **Order Service**: Position sizing and leverage recommendations
- **Decision Engine**: Risk-adjusted trade signals
- **Monitoring Service**: Risk alerts and limit violations
- **Reporting Service**: Risk metrics and performance attribution

## Configuration & Tuning

### 1. Risk Parameter Optimization
```go
type OptimizationResult struct {
    Parameters      Config
    BacktestResults BacktestMetrics
    SharpeRatio     float64
    MaxDrawdown     float64
    WinRate         float64
}

func OptimizeRiskParameters(historicalData []Trade, paramRanges map[string][]float64) *OptimizationResult {
    bestResult := &OptimizationResult{SharpeRatio: -999}
    
    // Grid search over parameter space
    for _, maxPos := range paramRanges["MaxPositionPercent"] {
        for _, stopMult := range paramRanges["StopATRMultiplier"] {
            for _, targetMult := range paramRanges["TargetATRMultiplier"] {
                config := Config{
                    MaxPositionPercent:  maxPos,
                    StopATRMultiplier:   stopMult,
                    TargetATRMultiplier: targetMult,
                }
                
                // Backtest with these parameters
                results := backTestWithConfig(historicalData, config)
                
                // Update best if improved
                if results.SharpeRatio > bestResult.SharpeRatio {
                    bestResult = &OptimizationResult{
                        Parameters:      config,
                        BacktestResults: results,
                        SharpeRatio:     results.SharpeRatio,
                        MaxDrawdown:     results.MaxDrawdown,
                        WinRate:         results.WinRate,
                    }
                }
            }
        }
    }
    
    return bestResult
}
```

## Monitoring & Reporting

### 1. Risk Dashboard Metrics
```go
type RiskDashboard struct {
    CurrentExposure     PortfolioExposure
    RiskMetrics         RiskMetrics
    RecentAlerts        []RiskAlert
    PerformanceMetrics  PerformanceAttribution
    
    // Real-time indicators
    RiskUtilization     float64  // % of risk budget used
    LeverageRatio       float64  // Current portfolio leverage
    VolatilityIndex     float64  // Portfolio volatility measure
    
    // Health indicators
    SystemStatus        string   // OK, WARNING, CRITICAL
    LastUpdate          time.Time
    
    // Charts data
    EquityCurve         []float64
    DrawdownCurve       []float64
    RiskUsageHistory    []float64
}
```

### 2. Compliance Reporting
```go
type ComplianceReport struct {
    Period              string
    TotalTrades         int
    LimitViolations     []LimitViolation
    RiskExceedances     []RiskExceedance
    MaxDrawdownReached  float64
    
    // Risk limit compliance
    PositionLimitBreaches int
    LeverageLimitBreaches int
    ExposureLimitBreaches int
    
    // Performance vs risk budget
    PlannedRisk         float64
    ActualRisk          float64
    RiskEfficiency      float64  // Return per unit risk
}
```

## Future Enhancements

### 1. Machine Learning Risk Models
- **Volatility Prediction**: ML models for ATR forecasting
- **Correlation Modeling**: Dynamic correlation prediction
- **Regime Detection**: Market state identification
- **Optimal Sizing**: Reinforcement learning for position sizing

### 2. Advanced Risk Metrics
- **Tail Risk Measures**: Extreme value theory applications
- **Regime-Conditional VaR**: VaR adjusted for market regimes
- **Options-Based Risk**: Implied volatility and skew metrics
- **Liquidity Risk**: Market impact and slippage modeling