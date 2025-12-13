# Decision Engine Service Specification

## Overview
The Decision Engine Service implements a sophisticated quantitative trading decision system based on multi-factor analysis. It processes market data, applies weighted scoring across 7 categories, and generates structured trade recommendations with confidence levels.

## Architecture
- **Package**: `internal/services/decision`
- **Type**: Stateless calculation engine
- **Dependencies**: Risk management service
- **Design Pattern**: Functional scoring system with weighted aggregation

## Core Components

### 1. Decision Input Structure
```go
type DecisionInput struct {
    // Technical Indicators
    RSI                 float64
    KDJ_K               float64
    KDJ_D               float64
    KDJ_J               float64
    ATR_Percent         float64
    
    // Price Data
    Price               float64
    RecentHigh          float64
    RecentLow           float64
    SupportLevel        float64
    ResistanceLevel     float64
    SwingHighBroken     bool
    SwingLowBroken      bool
    
    // Volume & Order Flow
    SpotVolumeChange    float64
    FuturesVolumeChange float64
    OpenInterestChange  float64
    OrderBookImbalance  float64
    
    // Market Sentiment
    FundingRate         float64
    LongShortRatio      float64
    ExchangeInflows     float64
    WhaleTransactions   int
    SpotFuturesPremium  float64
    
    // Macro & Sentiment
    MacroSentimentScore float64
    NewsSentimentScore  float64
    FearGreedIndex      float64
    
    // Risk Parameters
    Capital             float64
    CurrentPosition     float64
    MaxPositionSize     float64
}
```

### 2. Decision Output Structure
```go
type DecisionOutput struct {
    Prediction          string  // "Bump" | "Dump" | "Sideways"
    Confidence          int     // 0-100
    Bias                string  // "Bullish" | "Bearish" | "Neutral"
    TotalScore          float64
    CategoryScores      CategoryScores
    Action              string  // "Long" | "Short" | "Hold"
    PositionSizePercent float64
    Leverage            float64
    EntryPrice          float64
    StopLoss            float64
    TakeProfit          float64
    ScaleInPlan         string
    ScaleOutPlan        string
    Reasoning           string
}

type CategoryScores struct {
    MarketStructure     float64
    VolumeOrderFlow     float64
    FundingLongShort    float64
    OnChain             float64
    MacroSentiment      float64
    QuantModels         float64
    RiskManagement      float64
}
```

## Scoring System

### 1. Category Weights (TRADE.md Alignment)
```go
const (
    weightMarketStructure  = 0.25  // 25%
    weightVolumeOrderFlow  = 0.20  // 20%
    weightFundingLongShort = 0.15  // 15%
    weightOnChain          = 0.10  // 10%
    weightMacroSentiment   = 0.10  // 10%
    weightQuantModels      = 0.15  // 15%
    weightRiskManagement   = 0.05  // 5%
)
```

### 2. Individual Scoring Functions

#### Market Structure & Price Action (25%)
```go
func scoreMarketStructure(in *DecisionInput) float64
```
**Analysis Points:**
- **Breakout/Breakdown Detection**: ±2.0 for strong price breaks above/below recent highs/lows
- **Support/Resistance Proximity**: ±0.8 for prices near key levels
- **Range Position Analysis**: Position within trading range affects bias
- **Swing Level Confirmation**: Additional weight for broken swing points

#### Volume & Order Flow (20%)
```go
func scoreVolumeOrderFlow(in *DecisionInput) float64
```
**Analysis Points:**
- **Volume Divergence**: Compare spot vs futures volume changes
- **Open Interest Analysis**: Rising OI + price direction = trend confirmation
- **Order Book Imbalance**: Positive = buy pressure, Negative = sell pressure
- **Volume Confirmation**: High volume supports directional moves

#### Funding & Long/Short Ratio (15%)
```go
func scoreFundingLongShort(in *DecisionInput) float64
```
**Analysis Points:**
- **Funding Rate Analysis**: Extreme rates indicate potential squeezes
- **Long/Short Ratio**: Extreme imbalances suggest reversal opportunities
- **Leverage Positioning**: Over-leveraged positions create instability
- **Contrarian Signals**: High funding = potential opposite direction move

#### On-Chain Data (10%)
```go
func scoreOnChain(in *DecisionInput) float64
```
**Analysis Points:**
- **Exchange Flows**: Inflows = selling pressure, Outflows = accumulation
- **Whale Activity**: Large transaction counts indicate major moves
- **Spot-Futures Premium**: Extreme premiums suggest overheated sentiment
- **Network Activity**: Transaction volume and fees analysis

#### Macro & Sentiment (10%)
```go
func scoreMacroSentiment(in *DecisionInput) float64
```
**Analysis Points:**
- **Global Risk Appetite**: Correlation with traditional markets
- **News Impact**: Regulatory, adoption, and ecosystem developments
- **Fear & Greed Index**: Extreme readings for contrarian opportunities
- **Market Correlation**: Bitcoin dominance and altcoin relationships

#### Quantitative Models (15%)
```go
func scoreQuantModels(in *DecisionInput) float64
```
**Analysis Points:**
- **RSI Analysis**: Overbought/oversold conditions and divergences
- **KDJ Stochastic**: K/D crossovers and J-line extremes
- **Mean Reversion vs Momentum**: Determine market regime
- **Volatility Analysis**: ATR-based regime detection

#### Risk Management (5%)
```go
func scoreRiskManagement(in *DecisionInput) float64
```
**Analysis Points:**
- **Position Sizing**: Adherence to max position limits
- **Portfolio Exposure**: Current risk vs maximum allowable
- **Volatility Adjustment**: ATR-based position size scaling
- **Capital Preservation**: Drawdown protection mechanisms

## Decision Logic

### 1. Score Aggregation
```go
func calculateTotalScore(scores CategoryScores) float64 {
    return scores.MarketStructure*weightMarketStructure +
           scores.VolumeOrderFlow*weightVolumeOrderFlow +
           scores.FundingLongShort*weightFundingLongShort +
           scores.OnChain*weightOnChain +
           scores.MacroSentiment*weightMacroSentiment +
           scores.QuantModels*weightQuantModels +
           scores.RiskManagement*weightRiskManagement
}
```

### 2. Decision Thresholds
```go
const (
    longThreshold  = 0.8   // ≥ 0.8 = Bullish bias (Long)
    shortThreshold = -0.8  // ≤ -0.8 = Bearish bias (Short)
)
```

### 3. Confidence Calculation
- **Base Confidence**: Absolute value of total score * 50
- **Category Agreement**: Bonus for aligned category scores
- **Signal Strength**: Additional confidence for extreme readings
- **Risk Adjustment**: Reduction for high volatility environments

### 4. Position Sizing Algorithm
```go
func calculatePositionSize(totalScore, capital, atr float64) (float64, float64) {
    // Base size: 2-5% of capital based on confidence
    baseSize := math.Abs(totalScore) * 0.025 * capital
    
    // Volatility adjustment
    volAdjustment := math.Max(0.5, math.Min(2.0, 1.0/atr))
    
    // Final position size
    positionSize := baseSize * volAdjustment
    
    // Leverage calculation (2x-10x based on confidence)
    leverage := math.Max(2.0, math.Min(10.0, math.Abs(totalScore)*5))
    
    return positionSize, leverage
}
```

## Risk Controls

### 1. Entry Validation
- **Minimum Confidence**: 60% threshold for trade execution
- **Maximum Position**: 5% of total capital per trade
- **Volatility Limits**: No trades during extreme volatility
- **Market Hours**: Avoid low-liquidity periods

### 2. Stop Loss Calculation
```go
func calculateStopLoss(entryPrice, atr float64, isLong bool) float64 {
    stopDistance := atr * 2.0 // 2x ATR stop distance
    
    if isLong {
        return entryPrice - stopDistance
    }
    return entryPrice + stopDistance
}
```

### 3. Take Profit Levels
```go
func calculateTakeProfit(entryPrice, stopLoss float64, isLong bool) float64 {
    riskAmount := math.Abs(entryPrice - stopLoss)
    rewardMultiple := 2.0 // 2:1 reward-to-risk ratio
    
    if isLong {
        return entryPrice + (riskAmount * rewardMultiple)
    }
    return entryPrice - (riskAmount * rewardMultiple)
}
```

## Scaling Strategies

### 1. Scale-In Plan
- **Initial Position**: 50% of calculated size
- **Scale-In 1**: 30% if price moves favorably by 0.5% 
- **Scale-In 2**: 20% if price moves favorably by 1.0%
- **Condition**: Only scale-in if original thesis strengthens

### 2. Scale-Out Plan
- **Partial Profit**: 50% at 1:1 risk-reward
- **Runner Position**: 50% targeting 1:3 risk-reward
- **Trailing Stop**: Move stop to breakeven after 1:1 target
- **Final Exit**: Trail stop using ATR-based levels

## Error Handling

### 1. Input Validation
```go
func validateInput(in *DecisionInput) error {
    if in.Price <= 0 {
        return errors.New("invalid price")
    }
    if in.Capital <= 0 {
        return errors.New("invalid capital")
    }
    // Additional validations...
}
```

### 2. Boundary Checks
- **Score Clamping**: All scores limited to [-2, 2] range
- **Confidence Limits**: Bounded to [0, 100]
- **Position Sizing**: Maximum 5% of capital enforcement
- **Leverage Limits**: Capped at 10x maximum

## Integration Points

### 1. Input Sources
- **Analyze Service**: Technical indicator calculations (RSI, KDJ)
- **Market Service**: Price action and structure analysis
- **Risk Service**: Portfolio and exposure calculations
- **External APIs**: Sentiment and macro data feeds

### 2. Output Consumers
- **Order Service**: Trade execution based on decisions
- **Risk Service**: Portfolio updates and monitoring
- **Notification Service**: Alert generation and reporting
- **Analytics Service**: Performance tracking and optimization

## Performance Characteristics

### 1. Computational Complexity
- **O(1)**: All scoring functions are constant time
- **Memory**: Minimal allocation, stateless operation
- **Throughput**: >1000 decisions per second capability
- **Latency**: <10ms average decision time

### 2. Accuracy Metrics
- **Win Rate Target**: >55% profitable trades
- **Risk-Reward**: Maintain >1.5:1 average
- **Drawdown Limit**: <15% maximum portfolio drawdown
- **Sharpe Ratio**: Target >2.0 annually

## Monitoring & Validation

### 1. Decision Quality Metrics
- **Score Distribution**: Monitor category score patterns
- **Confidence vs Outcome**: Validate confidence accuracy
- **Category Performance**: Track individual category effectiveness
- **Threshold Optimization**: Periodic review of decision thresholds

### 2. Backtesting Framework
- **Historical Validation**: Test decisions against past market data
- **Paper Trading**: Live validation without real money
- **Performance Attribution**: Identify contributing factors
- **Strategy Evolution**: Continuous improvement based on results

## Future Enhancements

### 1. Machine Learning Integration
- **Pattern Recognition**: Deep learning for market patterns
- **Adaptive Weights**: Dynamic category weight adjustment
- **Sentiment Analysis**: NLP for news and social media
- **Regime Detection**: Automatic market state identification

### 2. Advanced Risk Management
- **Portfolio Optimization**: Modern Portfolio Theory integration
- **Correlation Analysis**: Cross-asset risk assessment
- **Dynamic Hedging**: Automatic hedging strategies
- **Stress Testing**: Scenario-based risk evaluation