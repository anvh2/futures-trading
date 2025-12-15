# Combined Decision Making System

## Overview

This document explains the integrated decision-making system that combines signal-based trading with a comprehensive scoring engine for futures trading. The system provides sophisticated, multi-factor analysis for trading decisions.

## Architecture

### Core Components

1. **Signal Processing**: Converts trading signals with technical indicators into structured decision inputs
2. **Scoring Engine**: Applies weighted scoring across 7 categories of market analysis
3. **Decision Output**: Generates comprehensive trading decisions with risk management

### Data Flow

```
Trading Signal → DecisionInput → Scoring Engine → DecisionOutput → TradingDecision
```

## Key Features

### 1. Multi-Category Scoring System

The engine evaluates seven weighted categories:

- **Market Structure (25%)**: Breakouts, support/resistance, trend analysis
- **Volume & Order Flow (20%)**: Spot/futures volume, order book, VWAP analysis  
- **Funding & Long/Short (15%)**: Funding rates, positioning sentiment
- **On-Chain Analysis (10%)**: Exchange flows, whale activity
- **Macro Sentiment (10%)**: News sentiment, fear/greed index
- **Quantitative Models (15%)**: Multi-timeframe RSI, KDJ oscillators
- **Risk Management (5%)**: ATR volatility, position conflicts

### 2. Multi-Timeframe Analysis

- **5-minute**: Short-term momentum
- **15-minute**: Medium-term trend
- **1-hour**: Primary trend direction

### 3. Sophisticated Risk Management

- **Position Sizing**: ATR-adjusted, confidence-based sizing
- **Leverage Calculation**: Dynamic based on volatility and confidence
- **Stop Loss/Take Profit**: ATR-based levels
- **Scale-in/Scale-out**: Progressive position management

### 4. Decision Confidence

Confidence is calculated based on:
- Signal magnitude across categories
- Agreement between different analysis methods
- Market volatility conditions

## Usage

### Basic Decision Making

```go
// Create a signal with comprehensive indicator data
signal := &models.Signal{
    Symbol:     "BTCUSDT",
    Price:      45000.0,
    Interval:   "1h",
    Strategy:   "comprehensive",
    Indicators: map[string]float64{
        "rsi":              35.0,
        "k":                25.0,
        "d":                20.0,
        "atr_percent":      2.5,
        "vwap":            44800.0,
        "funding_rate":     0.01,
        "long_short_ratio": 1.3,
        // ... more indicators
    },
    // ... metadata for trends, swing breaks, etc.
}

// Generate decision using the integrated system
decision := decisionMaker.MakeDecision(signal)
```

### Advanced Usage

```go
// Get detailed engine output for analysis
engineOutput := decisionMaker.GetEngineDecision(signal)

fmt.Printf("Prediction: %s\\n", engineOutput.Prediction)
fmt.Printf("Confidence: %d%%\\n", engineOutput.Confidence) 
fmt.Printf("Total Score: %.2f\\n", engineOutput.TotalScore)
fmt.Printf("Reasoning: %s\\n", engineOutput.Reasoning)
```

## Signal Requirements

### Required Indicators

**Technical Analysis:**
- `rsi`: Primary RSI (0-100)
- `k`, `d`, `j`: KDJ oscillator values
- `atr_percent`: ATR as percentage of price
- `vwap`: Volume Weighted Average Price

**Market Structure:**
- `recent_high`, `recent_low`: Recent price extremes
- `support_level`, `resistance_level`: Key levels
- `trend_strength`: Trend strength indicator (0-100)

**Volume Analysis:**
- `relative_volume`: Volume vs average (1.0 = normal)
- `volume_ratio`: Buy/sell volume ratio
- `spot_volume_change`, `futures_volume_change`: Volume changes

**Funding & Positioning:**
- `funding_rate`: Current funding rate
- `long_short_ratio`: Long vs short positioning
- `oi_change`: Open interest change

### Required Metadata

**Multi-timeframe Trends:**
- `trend_5m`, `trend_15m`, `trend_1h`: "UP", "DOWN", or "SIDEWAYS"

**Market Structure:**
- `swing_high_broken`, `swing_low_broken`: Boolean swing break flags

## Decision Output

### TradingDecision Structure

```go
type TradingDecision struct {
    Symbol     string    // Trading symbol
    Action     string    // "BUY", "SELL", "HOLD"  
    Size       float64   // Position size in base currency
    Price      float64   // Entry price
    Confidence float64   // Confidence (0-1)
    Timestamp  time.Time // Decision timestamp
    Metadata   map[string]interface{} // Rich decision data
}
```

### Rich Metadata

The decision includes comprehensive metadata:

- **prediction**: "Bump", "Dump", or "Sideways"
- **bias**: "Bullish", "Bearish", or "Neutral"
- **total_score**: Composite score (-2 to 2)
- **category_scores**: Individual category contributions
- **position_size_pct**: Position size as percentage of capital
- **leverage**: Recommended leverage (1-10)
- **stop_loss**: Calculated stop loss level
- **take_profit**: Calculated take profit level
- **scale_in_plan**: Position scaling strategy
- **scale_out_plan**: Exit scaling strategy  
- **reasoning**: Human-readable explanation

## Configuration

### Risk Parameters

Default risk configuration (configurable):

```go
- Max Position Size: 10% of capital
- Max Leverage: 10x (adjusted by volatility)
- Min Confidence: 60% for trade execution
- ATR Stop Loss: 2x ATR
- ATR Take Profit: 3x ATR
```

### Scoring Thresholds

- **Long Threshold**: 0.8 (composite score)
- **Short Threshold**: -0.8 (composite score)
- **Neutral Zone**: Between -0.8 and 0.8

## Testing

The system includes comprehensive tests:

```bash
# Run integration tests
go test ./internal/services/decision -v -run TestIntegratedDecisionMaking

# Run edge case tests  
go test ./internal/services/decision -v -run TestEdgeCases

# Run benchmarks
go test ./internal/services/decision -bench=BenchmarkDecisionMaking
```

## Best Practices

### 1. Signal Quality
- Ensure all critical indicators are populated
- Validate signal data before processing
- Use consistent timeframe analysis

### 2. Risk Management
- Never ignore engine confidence levels
- Respect position sizing recommendations
- Monitor ATR volatility conditions

### 3. Performance
- Cache frequently accessed data
- Use appropriate timeframe for signal generation
- Consider market conditions in signal creation

## Error Handling

The system gracefully handles:
- Missing indicator values (defaults to 0)
- Invalid signal data (returns nil)
- Extreme market conditions (adjusts confidence)
- Position conflicts (reduces scoring)

## Future Enhancements

Potential improvements:
- Machine learning model integration
- Dynamic parameter optimization
- Enhanced multi-asset correlation
- Real-time sentiment analysis integration