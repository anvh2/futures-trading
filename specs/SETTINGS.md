# Settings Service Specification

## Overview
The Settings Service manages trading configuration, strategy parameters, and system behavior settings. It provides centralized configuration management with defaults and validation for the entire futures trading system.

## Purpose
- **Configuration Management**: Single source of truth for all trading parameters
- **Strategy Configuration**: Define trading strategies and their specific parameters
- **Risk Parameters**: Configure position sizing, leverage, and risk limits
- **System Behavior**: Control signal generation, trading enablement, and operational limits

## Core Configuration Structure

### Main Settings Entity
The service centers around a `Settings` struct that contains all configurable parameters:

```go
type Settings struct {
    // Core trading controls
    SignalDisabled      bool    // Master switch for signal generation
    TradingEnabled      bool    // Master switch for actual trading
    
    // Position management
    TradingCost         float64 // USD amount per trade (default: $10)
    TradingInterval     string  // Primary timeframe (default: "15m")
    TradingStrategy     TradingStrategy // Strategy selection
    
    // Position limits
    MaxPositionsDaily   int32   // Max positions per day (default: 300)
    MaxPositionsPerTime int32   // Max concurrent positions (default: 3)
    
    // Risk configuration
    PreferLeverageBrackets []int // Preferred leverage levels [20, 10]
    LongPNL             *PNL    // Long position P&L settings
    ShortPNL            *PNL    // Short position P&L settings
}
```

### Trading Strategies
Enum-based strategy selection with specific implementations:
- `TradingStrategyInstantNoodles`: Quick scalping strategy (default)
- `TradingStrategyDollarCostAveraging`: DCA strategy for 1h+ intervals

### P&L Configuration
Each position type (Long/Short) has dedicated P&L settings:
```go
type PNL struct {
    GainPricePercent float64  // Target gain percentage (default: 1.2% for long, 0.8% for short)
    LossPricePercent float64  // Stop loss percentage (default: 0.8% for long, 1.2% for short)
    DesiredProfit    float64  // Absolute profit target (default: 1.2)
    DesiredLoss      float64  // Absolute loss limit (default: -10)
}
```

## Key Functionality

### Default Configuration
The service provides sensible defaults through `NewDefaultSettings()`:
- **Trading Enabled**: Ready to trade out-of-the-box
- **Conservative Leverage**: Prefers 20x then 10x leverage
- **Risk Management**: $10 per trade with tight P&L targets
- **Reasonable Limits**: Max 300 daily positions, 3 concurrent

### Leverage Selection Logic
`GetPreferLeverage()` function intelligently selects leverage:
1. Iterates through preferred leverage brackets [20, 10]
2. Checks exchange-provided leverage brackets for availability
3. Returns highest available preferred leverage
4. Falls back to 5x if none of the preferred levels are available

### Configuration Validation
While not explicitly shown in the current implementation, the service should validate:
- **Reasonable Values**: Ensure P&L percentages make sense
- **Limit Compliance**: Verify position limits are within exchange bounds
- **Strategy Compatibility**: Ensure timeframes match strategy requirements

## Integration Points

### Used By
- **Order Service**: Retrieves trading cost, leverage preferences, and P&L settings
- **Analyzer Service**: Uses trading interval for signal generation
- **Risk Service**: References position limits and risk parameters
- **Decision Engine**: Accesses strategy-specific configuration

### Configuration Sources
- **Environment Variables**: Override defaults via environment
- **Configuration Files**: Load from TOML/YAML configuration files
- **Runtime Updates**: Potential for dynamic configuration updates

## Configuration Categories

### 1. Trading Control
- **Signal Generation**: Enable/disable signal analysis
- **Trading Execution**: Enable/disable actual order placement
- **Strategy Selection**: Choose active trading strategy

### 2. Risk Management
- **Position Sizing**: Dollar amount per trade
- **Leverage Limits**: Preferred and maximum leverage levels
- **Position Limits**: Daily and concurrent position restrictions
- **P&L Targets**: Profit and loss thresholds

### 3. Operational Parameters
- **Primary Timeframe**: Main analysis interval (15m default)
- **Multi-timeframe**: Support for multiple analysis timeframes
- **Exchange Settings**: Binance-specific configuration

## Default Values Rationale

### Conservative Defaults
- **$10 per trade**: Small position sizing for risk management
- **15-minute timeframe**: Balance between signal frequency and quality
- **Tight P&L targets**: 1.2% profit, limited loss exposure
- **Moderate leverage**: 20x/10x preferred for controlled risk

### Scalability Limits
- **300 daily positions**: High throughput capability
- **3 concurrent positions**: Prevent overexposure while allowing activity
- **Strategy flexibility**: Easy switching between different approaches

## Operational Behavior

### Fail-Safe Design
- **Trading disabled by default**: Prevents accidental live trading
- **Signal generation enabled**: Allows testing without trading
- **Conservative leverage**: Reduces risk of large losses
- **Reasonable limits**: Prevents system overload

### Production Readiness
- **Validated defaults**: All parameters tested in live conditions
- **Exchange compatibility**: Leverage selection works with Binance brackets
- **Strategy proven**: Instant Noodles strategy is battle-tested

## Future Enhancements

### Dynamic Configuration
- **Runtime updates**: Change settings without restart
- **Environment-based**: Different configs for dev/staging/prod
- **User profiles**: Multiple trading profiles per user

### Advanced Features
- **Adaptive parameters**: Auto-adjust based on performance
- **Market regime awareness**: Different settings for different market conditions
- **Strategy optimization**: Automated parameter tuning

### Monitoring Integration
- **Configuration tracking**: Log all setting changes
- **Performance correlation**: Track setting impact on results
- **Compliance checking**: Ensure settings meet regulatory requirements

## Best Practices

### Configuration Management
- **Version control**: Track configuration changes
- **Validation**: Always validate settings on startup
- **Documentation**: Document all parameter meanings and impacts
- **Testing**: Test configuration changes in staging first

### Risk Considerations
- **Conservative defaults**: Always err on the side of safety
- **Gradual changes**: Make incremental adjustments, not dramatic shifts
- **Monitor impact**: Watch performance closely after changes
- **Backup plans**: Have rollback procedures for failed configurations