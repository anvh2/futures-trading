# Analyze Service Specification

## Overview
The Analyze Service is responsible for technical analysis of market data, calculating oscillators (RSI, KDJ), and generating trading signals based on candlestick patterns and technical indicators.

## Architecture
- **Package**: `internal/services/analyze`
- **Type**: Worker-based service with queue processing
- **Dependencies**: Market Cache, Exchange Cache, Queue, Channel, Settings, Telegram notifications

## Core Components
#### `Start() error`
- Starts the analyzer with continuous processing
- Runs every 10 seconds via ticker
- Processes all symbols from exchangeCache
- Validates candle data completeness before analysis

#### `process(ctx context.Context, data interface{}) error`
- Main processing function for analyzing market data
- Input: CandleSummary with multiple timeframes
- Calculates RSI and KDJ oscillators for all intervals
- Validates signals are within trading range bounds
- Pushes notifications and trading signals to queue

#### Technical Analysis Functions
- **RSI Calculation**: 14-period Relative Strength Index
- **KDJ Calculation**: 9-3-3 period stochastic oscillator with smoothing
- **Range Bound Validation**: Ensures signals are within recommended trading bounds

## Data Flow

### Input Processing
1. **Source**: Market data from crawler service
2. **Format**: CandleSummary containing candles for multiple intervals
3. **Validation**: Ensures candle data completeness and current candle validity

### Technical Analysis Pipeline
1. **Data Extraction**: Extract OHLC data from candles
2. **Oscillator Calculation**: 
   - RSI: Uses closing prices over 14 periods
   - KDJ: Uses high/low/close over 9/3/3 periods
3. **Signal Generation**: Creates Oscillator model with calculated values

### Output Generation
1. **Signal Validation**: Checks if oscillators are within trading bounds
2. **Notification**: Sends formatted signal to Telegram
3. **Queue Push**: Forwards signal to order processing queue

## Signal Format

### Oscillator Model
### Telegram Notification Format
```
#SYMBOL                 [X.XX(s) ago]
    POSITION_RECOMMENDATION

    15M:    RSI XX.XX | K XX.XX | D XX.XX
    1H:     RSI XX.XX | K XX.XX | D XX.XX
    4H:     RSI XX.XX | K XX.XX | D XX.XX
```

## Configuration Parameters

### Trading Intervals
- Primary: Configurable via `settings.TradingInterval` (default: 15m)
- Additional: Multiple intervals from `market.intervals` config

### Range Bounds
- Uses `talib.RangeBoundRecommend` for signal validation
- Only signals within bounds are processed for trading

### Signal Throttling
- Prevents duplicate signals within 10 minutes per symbol-interval
- Uses cache with key format: `signal.sent.{symbol}-{interval}`

## Error Handling

### Data Validation
- **Candle Validation**: Ensures sufficient candle data for analysis
- **Current Candle Check**: Validates candle is from current timeframe
- **Oscillator Validation**: Ensures all required oscillators are calculated

### Recovery Mechanisms
- **Panic Recovery**: Worker pool handles panics gracefully
- **Error Logging**: All errors logged with context (symbol, interval, error)
- **Graceful Degradation**: Skip symbols with invalid data

## Performance Considerations

### Worker Pool Configuration
- **Concurrency**: 2 workers for parallel processing
- **Queue Management**: Automatic work distribution
- **Resource Limits**: Controlled memory usage through worker pools

### Caching Strategy
- **Signal Cache**: Prevents duplicate signal processing
- **Expiration**: Automatic cleanup of old cache entries
- **Memory Efficiency**: Minimal data stored in cache

## Integration Points

### Upstream Services
- **Crawler Service**: Provides market data via market cache
- **Settings Service**: Configuration for intervals and parameters

### Downstream Services
- **Order Service**: Receives signals via queue
- **Telegram Service**: Receives notifications for signal alerts

## Monitoring & Observability

### Logging
- **Info Level**: Successful analysis completion with symbol details
- **Error Level**: Processing failures, queue push errors, notification failures
- **Debug Context**: Symbol, interval, timing information

### Metrics
- Processing time per symbol
- Queue depth monitoring
- Signal generation rate
- Error rate tracking

## Quality Assurance

### Signal Quality
- Range bound validation ensures only high-probability signals
- Multiple timeframe confirmation
- Proper oscillator calculation using proven formulas

### Data Integrity
- Validates candle completeness before processing
- Ensures current timeframe alignment
- Handles edge cases (insufficient data, malformed candles)

## Future Enhancements

### Additional Indicators
- MACD for trend confirmation
- Bollinger Bands for volatility analysis
- Volume-based indicators

### Advanced Signal Logic
- Multi-timeframe confluence
- Pattern recognition (candlestick patterns)
- Volume confirmation requirements

### Performance Optimization
- Parallel oscillator calculation
- Incremental analysis for performance
- Advanced caching strategies