# Crawl Service Specification

## Overview
The Crawl Service is responsible for collecting real-time market data from Binance Futures, managing WebSocket connections, fetching exchange information, and maintaining market data caches.

## Architecture
- **Package**: `internal/services/crawl`
- **Type**: Real-time data collection service with WebSocket management
- **Dependencies**: Binance API, Market Cache, Exchange Cache, Channel, Telegram notifications

## Core Components
#### Market Data Management
- **`Start() error`**: Main entry point, orchestrates all crawling operations
- **`fetchExchange() error`**: Retrieves and caches exchange information
- **`fetchMarketSummary(ctx) error`**: Fetches historical candlestick data
- **`StartConsumption()`**: Manages real-time WebSocket data streams
- **`StartRetry()`**: Handles connection retry logic
- **`StartNotification()`**: Processes order and position notifications

### 3. Data Collection Operations

#### Exchange Information (`fetchExchange`)
- **Purpose**: Collect trading symbols and their specifications
- **Source**: Binance `/fapi/v1/exchangeInfo` endpoint
- **Filtering**: USDT-margined symbols only, excludes blacklisted symbols
- **Updates**: Every 15 minutes to catch new listings
- **Output**: Cached symbol information with trading filters

#### Historical Data (`fetchMarketSummary`)
- **Purpose**: Populate initial candlestick data for all symbols
- **Concurrency**: Parallel fetching per interval (1m, 5m, 15m, 1h, 4h)
- **Limits**: Configurable candle count via `chart.candles.limit`
- **Error Handling**: Graceful degradation for failed symbols
- **Performance**: Concurrent goroutines with progress tracking

#### Real-time Data (`StartConsumption`)
- **WebSocket Streams**: Candlestick data for all symbols and intervals
- **Connection Management**: Auto-reconnection on failures
- **Data Validation**: Ensures data integrity before caching
- **Error Recovery**: Retry mechanism with exponential backoff

## Data Flow

### Symbol Discovery Pipeline
1. **Exchange Info Fetch**: Get all available symbols from Binance
2. **Symbol Filtering**: Apply USDT-margin and blacklist filters
3. **Filter Parsing**: Extract price/quantity precision and limits
4. **Cache Update**: Store filtered symbols in exchange cache

### Market Data Pipeline
1. **Historical Data Sync**: Fetch initial candles for all intervals
2. **WebSocket Subscription**: Subscribe to real-time candlestick streams
3. **Data Processing**: Validate and normalize incoming data
4. **Cache Management**: Update market cache with latest candles

### Notification Pipeline
1. **Order Events**: Monitor order execution via user data stream
2. **Position Updates**: Track position changes and PnL
3. **Alert Generation**: Send notifications for important events
4. **Error Reporting**: Alert on connection or data issues

## WebSocket Management

### Stream Configuration
```go
// Candlestick streams for all symbols and intervals
streams := []string{
    "btcusdt@kline_1m",
    "btcusdt@kline_5m", 
    "btcusdt@kline_15m",
    "btcusdt@kline_1h",
    "btcusdt@kline_4h",
    // ... for all symbols
}
```

### Connection Handling
- **Auto-reconnection**: Automatic retry on connection loss
- **Heartbeat**: Maintain connection with periodic pings
- **Error Recovery**: Exponential backoff for failed connections
- **Graceful Shutdown**: Clean connection closure on service stop

### Data Validation
- **Timestamp Checks**: Ensure candle data is current
- **Symbol Validation**: Verify symbol exists in exchange cache
- **Completeness**: Check all required OHLC values are present
- **Deduplication**: Prevent processing duplicate candle data

## Caching Strategy

### Exchange Cache
- **Content**: Symbol information, trading filters, precision
- **Update Frequency**: Every 15 minutes
- **Persistence**: In-memory with periodic refresh
- **Access Pattern**: Read-heavy for symbol lookups

### Market Cache
- **Content**: Candlestick data for multiple intervals
- **Update Frequency**: Real-time via WebSocket
- **Retention**: Configurable candle history limit
- **Structure**: Symbol → Interval → Candle array

## Error Handling & Recovery

### Connection Failures
- **Retry Logic**: Exponential backoff with maximum retry count
- **Fallback Strategy**: REST API fallback for critical data
- **Circuit Breaker**: Temporary suspension of failing streams
- **Alert System**: Notification on persistent failures

### Data Quality Issues
- **Validation Errors**: Skip malformed data, log for investigation
- **Missing Data**: Backfill via REST API when gaps detected
- **Stale Data**: Timeout mechanism for outdated information
- **Inconsistency**: Cross-validation between streams and REST

## Performance Optimizations

### Concurrent Processing
- **Parallel Fetching**: Simultaneous requests for different intervals
- **Goroutine Pooling**: Controlled concurrency to prevent resource exhaustion
- **Batch Processing**: Group operations for efficiency
- **Memory Management**: Efficient data structures and cleanup

### Network Efficiency
- **Connection Pooling**: Reuse HTTP connections
- **Compression**: Enable gzip for REST API calls
- **Rate Limiting**: Respect Binance API limits
- **Bandwidth Control**: Optimize WebSocket payload size

## Configuration Parameters

### Market Data Settings
```toml
[market]
intervals = ["1m", "5m", "15m", "1h", "4h"]

[chart.candles]
limit = 1000  # Historical candles per symbol

[crawl]
retry_max = 5
retry_interval = "30s"
websocket_timeout = "60s"
```

### Symbol Filtering
```go
// Blacklisted symbols (in crawler.go)
blacklist = map[string]bool{
    "BTCDOMUSDT": true,  // Dominance tokens
    "DEFIUSDT":   true,  // Index tokens
}

// Filter criteria
- MarginAsset == "USDT"  // USDT-margined only
- !contains(Symbol, "_") // Exclude leveraged tokens
```

## Integration Points

### Upstream Dependencies
- **Binance API**: Real-time and historical market data
- **Configuration**: Trading intervals and data limits

### Downstream Consumers
- **Analyze Service**: Uses cached market data for analysis
- **Order Service**: Accesses symbol information for trading
- **Risk Service**: Uses market data for risk calculations

## Monitoring & Observability

### Performance Metrics
- **Connection Status**: WebSocket connection health
- **Data Throughput**: Candles processed per second
- **Cache Hit Rate**: Efficiency of market data access
- **API Rate Limits**: Usage vs. available limits

### Health Checks
- **Symbol Count**: Verify expected number of symbols cached
- **Data Freshness**: Check latest candle timestamps
- **Connection Stability**: Monitor reconnection frequency
- **Error Rates**: Track processing failures

### Logging
- **Info Level**: Successful operations, cache updates, symbol counts
- **Warning Level**: Connection retries, missing data
- **Error Level**: API failures, connection losses, data corruption

## Quality Assurance

### Data Integrity
- **Timestamp Validation**: Ensure chronological order
- **Price Validation**: Check for reasonable price ranges
- **Volume Validation**: Verify non-negative volumes
- **Completeness**: Ensure all OHLC values present

### System Reliability
- **Graceful Degradation**: Continue operation with partial data
- **Auto-recovery**: Automatic restart of failed components
- **Resource Protection**: Prevent memory leaks and goroutine leaks
- **Controlled Shutdown**: Clean termination of all connections

## Security Considerations

### API Security
- **Credential Management**: Secure storage of API keys
- **Rate Limit Protection**: Prevent API limit violations
- **Error Exposure**: Avoid leaking sensitive information in logs
- **Network Security**: Use secure connections (HTTPS/WSS)

### Data Privacy
- **Log Sanitization**: Remove sensitive data from logs
- **Access Control**: Limit access to market data
- **Audit Trail**: Track data access and modifications

## Future Enhancements

### Data Sources
- **Multi-exchange Support**: Integrate additional exchanges
- **On-chain Data**: Blockchain-based market indicators
- **News Integration**: Sentiment analysis from news feeds
- **Social Signals**: Twitter/Reddit sentiment tracking

### Advanced Features
- **Data Compression**: Efficient storage of historical data
- **Real-time Analytics**: Edge processing of market data
- **Predictive Caching**: Pre-fetch likely needed data
- **Quality Scoring**: Rate data quality for different sources