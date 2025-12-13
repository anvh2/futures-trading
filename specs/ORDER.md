# Order Service Specification

## Overview
The Order Service manages the complete order lifecycle including position assessment, order creation, execution, and monitoring. It implements sophisticated trading strategies, risk management, and position sizing algorithms.

## Architecture
- **Package**: `internal/services/order`
- **Type**: Worker-based execution service with queue processing
- **Dependencies**: Binance API, Market Cache, Exchange Cache, Queue, Settings, Notifications

## Core Components

### 1. Orderer Structure
```go
type Orderer struct {
    logger        *logger.Logger
    binance       *binance.Binance      // Exchange API client
    notify        telegram.Notify       // Notification service
    queue         *queue.Queue         // Signal processing queue
    settings      *settings.Settings    // Trading configuration
    cache         cache.Basic          // Order processing cache
    worker        *worker.Worker       // Worker pool for execution
    marketCache   cache.Market         // Market data access
    exchangeCache cache.Exchange       // Symbol information
    quitChannel   chan struct{}        // Graceful shutdown
}
```

### 2. Order Models
```go
type Order struct {
    Symbol       string
    Side         string              // "BUY" | "SELL"
    Type         string              // "LIMIT" | "MARKET" | "STOP_MARKET"
    Quantity     float64
    Price        float64
    StopPrice    float64             // For stop orders
    TimeInForce  string              // "GTC" | "IOC" | "FOK"
    PositionSide string              // "LONG" | "SHORT"
    WorkingType  string              // "CONTRACT_PRICE" | "MARK_PRICE"
}

type Price struct {
    Entry    float64  // Entry price
    Quantity float64  // Position size
    Profit   float64  // Take profit price
    Loss     float64  // Stop loss price
}

type Position struct {
    Symbol           string
    PositionAmt      string
    EntryPrice       string
    MarkPrice        string
    UnrealizedProfit string
    PositionSide     string
}
```

## Core Functions

### 1. Order Processing Pipeline (`open`)
**Purpose**: Main order processing function triggered by trading signals

**Process Flow**:
1. **Signal Validation**: Verify trading is enabled and signal format
2. **Concurrency Control**: Prevent duplicate processing for same symbol
3. **Position Check**: Verify no existing position exists
4. **Price Appraisal**: Calculate optimal entry, stop, and target prices
5. **Order Creation**: Generate complete order set
6. **Risk Validation**: Final risk checks before execution
7. **Order Submission**: Execute orders via Binance API
8. **Monitoring Setup**: Track order status and fills

```go
func (s *Orderer) open(ctx context.Context, data interface{}) error {
    // 1. Validate trading enabled
    if !s.settings.TradingEnabled {
        return errors.New("trading: trading is disabled")
    }
    
    // 2. Parse oscillator signal
    oscillator := &models.Oscillator{}
    if err := json.Unmarshal([]byte(fmt.Sprint(data)), oscillator); err != nil {
        return err
    }
    
    // 3. Prevent duplicate processing
    if s.cache.Exists(oscillator.Symbol) {
        return nil // Already processing
    }
    
    // 4. Check for existing positions
    positions, err := s.binance.GetPositionRisk(ctx, "")
    if err != nil {
        return err
    }
    
    if positionExisted(positions, oscillator.Symbol) {
        return nil // Position already exists
    }
    
    // 5. Determine position side and appraise prices
    positionSide := helpers.ResolvePositionSide(oscillator.GetRSI(s.settings.TradingInterval))
    priceData, err := s.appraise(ctx, oscillator.Symbol, positionSide)
    if err != nil {
        return err
    }
    
    // 6. Create orders
    orders, err := s.create(ctx, oscillator.Symbol, oscillator.Stoch[s.settings.TradingInterval])
    if err != nil {
        return err
    }
    
    // 7. Execute orders
    return s.executeOrders(ctx, orders)
}
```

### 2. Price Appraisal (`appraise`)
**Purpose**: Calculate optimal entry, stop loss, and take profit prices based on market data and strategy

**Strategy Implementation**:
- **Instant Noodles Strategy**: Quick scalping with tight risk management
- **Dollar Cost Averaging**: (Future implementation for longer timeframes)

```go
func (s *Orderer) appraise(ctx context.Context, symbol string, positionSide futures.PositionSideType) (*models.Price, error) {
    // 1. Get leverage bracket information
    leverageBrackets, err := s.binance.GetLeverageBracket(ctx, symbol)
    if err != nil {
        return nil, err
    }
    
    leverage := s.settings.GetPreferLeverage(leverageBrackets)
    
    // 2. Get current market price
    symbolPrice, err := s.binance.GetCurrentPrice(ctx, symbol)
    if err != nil {
        return nil, err
    }
    
    // 3. Get recent candlestick data for structure analysis
    candles, err := s.binance.GetCandlesticks(ctx, symbol, s.settings.TradingInterval, 2, 0, 0)
    if err != nil {
        return nil, err
    }
    
    if len(candles) < 2 {
        return nil, errors.New("orders: insufficient candle data")
    }
    
    price := &models.Price{}
    current := helpers.StringToFloat(symbolPrice.Price)
    
    // 4. Apply strategy-specific pricing logic
    switch s.settings.TradingStrategy {
    case settings.TradingStrategyInstantNoodles:
        switch positionSide {
        case futures.PositionSideTypeShort:
            // Short entry: High of recent candles with buffer
            price.Entry = helpers.MaxFloat(candles[0].High, candles[1].High)
            if price.Entry < current {
                price.Entry = current * 1.01 // 1% above current for limit order
            }
            
            // Position sizing based on capital and leverage
            price.Quantity = s.settings.TradingCost * float64(leverage) / price.Entry
            
            // Profit/Loss targets
            price.Profit = price.Entry - s.settings.ShortPNL.DesiredProfit/price.Quantity
            price.Loss = price.Entry - s.settings.ShortPNL.DesiredLoss/price.Quantity
            
        case futures.PositionSideTypeLong:
            // Long entry: Low of recent candles with buffer
            price.Entry = helpers.MinFloat(candles[0].Low, candles[1].Low)
            if price.Entry > current {
                price.Entry = current * 0.99 // 1% below current for limit order
            }
            
            price.Quantity = s.settings.TradingCost * float64(leverage) / price.Entry
            price.Profit = s.settings.LongPNL.DesiredProfit/price.Quantity + price.Entry
            price.Loss = s.settings.LongPNL.DesiredLoss/price.Quantity + price.Entry
        }
    }
    
    return price, nil
}
```

### 3. Order Creation (`create`)
**Purpose**: Generate complete order set including entry, stop loss, and take profit orders

```go
func (s *Orderer) create(ctx context.Context, symbol string, stoch *models.Stoch) ([]*models.Order, error) {
    var orders []*models.Order
    
    // Determine position side from RSI
    positionSide := futures.PositionSideTypeLong
    if stoch.RSI > 70 {
        positionSide = futures.PositionSideTypeShort
    }
    
    // Get price calculations
    priceData, err := s.appraise(ctx, symbol, positionSide)
    if err != nil {
        return nil, err
    }
    
    // 1. Entry Order (Limit Order)
    entryOrder := &models.Order{
        Symbol:       symbol,
        Side:         getSide(positionSide, "entry"),
        Type:         "LIMIT",
        Quantity:     priceData.Quantity,
        Price:        priceData.Entry,
        TimeInForce:  "GTC",
        PositionSide: string(positionSide),
    }
    orders = append(orders, entryOrder)
    
    // 2. Stop Loss Order (Stop Market Order)
    stopOrder := &models.Order{
        Symbol:       symbol,
        Side:         getSide(positionSide, "close"),
        Type:         "STOP_MARKET",
        Quantity:     priceData.Quantity,
        StopPrice:    priceData.Loss,
        TimeInForce:  "GTC",
        PositionSide: string(positionSide),
        WorkingType:  "CONTRACT_PRICE",
    }
    orders = append(orders, stopOrder)
    
    // 3. Take Profit Order (Limit Order)
    profitOrder := &models.Order{
        Symbol:       symbol,
        Side:         getSide(positionSide, "close"),
        Type:         "LIMIT",
        Quantity:     priceData.Quantity,
        Price:        priceData.Profit,
        TimeInForce:  "GTC",
        PositionSide: string(positionSide),
    }
    orders = append(orders, profitOrder)
    
    return orders, nil
}
```

### 4. Position Management Functions

#### Position Validation (`positionExisted`)
```go
func positionExisted(positions []*Position, symbol string) bool {
    for _, position := range positions {
        if position.Symbol == symbol {
            positionAmt := helpers.StringToFloat(position.PositionAmt)
            if positionAmt != 0 {
                return true
            }
        }
    }
    return false
}
```

#### Quantity Calculations
```go
func calculateQuantity(price, amount float64) float64 {
    return amount / price
}

func calculateStopQuantity(price float64, totalAmount float64) float64 {
    return totalAmount / price
}
```

## Trading Strategies

### 1. Instant Noodles Strategy
**Characteristics**:
- Quick entry/exit scalping strategy
- Tight stop losses and profit targets
- High frequency trading approach
- Based on recent candle extremes

**Entry Logic**:
- **Long**: Enter below recent lows for better fills
- **Short**: Enter above recent highs for better fills
- Use limit orders with small buffer from current price

**Risk Management**:
- **Position Size**: Based on fixed USD amount and leverage
- **Stop Loss**: Calculated from PNL settings
- **Take Profit**: Risk-reward ratio based targets

### 2. Dollar Cost Averaging (DCA)
**Status**: Planned implementation
**Use Case**: Longer timeframe positions (1h intervals)
**Approach**: Gradual position building with multiple entries

## Risk Management Integration

### 1. Position Sizing
```go
type PositionSizing struct {
    MaxPositionValue float64  // Maximum USD value per position
    MaxLeverage      int      // Maximum allowed leverage
    RiskPerTrade     float64  // Percentage of capital at risk
}

func (ps *PositionSizing) CalculateSize(capital, entryPrice, stopPrice float64, leverage int) float64 {
    // Risk-based sizing
    riskAmount := capital * ps.RiskPerTrade
    priceRisk := math.Abs(entryPrice - stopPrice)
    maxQuantity := riskAmount / priceRisk
    
    // Leverage-based sizing
    leverageQuantity := ps.MaxPositionValue * float64(leverage) / entryPrice
    
    // Return the smaller of the two
    return math.Min(maxQuantity, leverageQuantity)
}
```

### 2. Exposure Limits
- **Maximum Positions**: Configured daily and concurrent limits
- **Symbol Exposure**: Prevent oversized positions in single assets
- **Correlation Limits**: Consider correlated positions
- **Leverage Management**: Dynamic leverage based on volatility

## Error Handling & Recovery

### 1. Order Execution Errors
```go
func (s *Orderer) handleOrderError(order *models.Order, err error) error {
    switch {
    case strings.Contains(err.Error(), "insufficient balance"):
        // Reduce position size and retry
        return s.retryWithReducedSize(order)
        
    case strings.Contains(err.Error(), "price filter"):
        // Adjust price to meet exchange requirements
        return s.adjustPriceAndRetry(order)
        
    case strings.Contains(err.Error(), "lot size"):
        // Round quantity to valid precision
        return s.adjustQuantityAndRetry(order)
        
    default:
        // Log error and alert
        s.logger.Error("Order execution failed", zap.Error(err))
        return s.notify.AlertOrderError(order, err)
    }
}
```

### 2. Recovery Mechanisms
- **Partial Fills**: Handle partially filled orders gracefully
- **Order Rejection**: Automatic retry with adjusted parameters
- **Connection Loss**: Queue orders for retry when connection restored
- **Rate Limiting**: Exponential backoff for rate limit errors

## Performance Optimizations

### 1. Concurrent Processing
- **Worker Pool**: 8 worker goroutines for order processing
- **Symbol Isolation**: Prevent blocking between different symbols
- **Queue Management**: Efficient signal processing queue
- **Cache Management**: Symbol-based caching to prevent duplicates

### 2. Order Batching
```go
func (s *Orderer) batchOrders(orders []*models.Order) error {
    // Group orders by symbol for efficient execution
    ordersBySymbol := make(map[string][]*models.Order)
    for _, order := range orders {
        ordersBySymbol[order.Symbol] = append(ordersBySymbol[order.Symbol], order)
    }
    
    // Execute orders in parallel by symbol
    var wg sync.WaitGroup
    errors := make(chan error, len(ordersBySymbol))
    
    for symbol, symbolOrders := range ordersBySymbol {
        wg.Add(1)
        go func(symbol string, orders []*models.Order) {
            defer wg.Done()
            errors <- s.executeSymbolOrders(symbol, orders)
        }(symbol, symbolOrders)
    }
    
    wg.Wait()
    close(errors)
    
    // Check for errors
    for err := range errors {
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

## Monitoring & Observability

### 1. Order Tracking
```go
type OrderStatus struct {
    OrderID      string
    Symbol       string
    Status       string    // "NEW", "FILLED", "CANCELED", "REJECTED"
    FilledQty    float64
    AvgPrice     float64
    Commission   float64
    Timestamp    time.Time
}

func (s *Orderer) trackOrderExecution(orderID string) {
    // Monitor order status via WebSocket
    // Update internal tracking
    // Send notifications on fills
    // Calculate performance metrics
}
```

### 2. Performance Metrics
- **Fill Rate**: Percentage of orders successfully filled
- **Slippage**: Difference between expected and actual fill prices
- **Execution Time**: Time from signal to order placement
- **Success Rate**: Profitable vs unprofitable trades
- **Risk Metrics**: Actual vs planned risk exposure

### 3. Health Monitoring
```go
type ServiceHealth struct {
    OrdersProcessed    int64
    OrdersSuccessful   int64
    OrdersFailed       int64
    AverageLatency     time.Duration
    QueueDepth         int
    ActivePositions    int
    TotalExposure      float64
}

func (s *Orderer) getHealthMetrics() *ServiceHealth {
    return &ServiceHealth{
        OrdersProcessed:  s.metrics.ordersProcessed,
        OrdersSuccessful: s.metrics.ordersSuccessful,
        OrdersFailed:     s.metrics.ordersFailed,
        AverageLatency:   s.calculateAverageLatency(),
        QueueDepth:       s.queue.Depth(),
        ActivePositions:  s.countActivePositions(),
        TotalExposure:    s.calculateTotalExposure(),
    }
}
```

## Integration Points

### 1. Upstream Dependencies
- **Analyze Service**: Receives trading signals via queue
- **Decision Engine**: Implements trade decisions
- **Settings Service**: Configuration for strategies and limits

### 2. Downstream Dependencies
- **Binance API**: Order execution and position management
- **Risk Service**: Real-time risk monitoring
- **Notification Service**: Trade alerts and error reporting

## Configuration Management

### 1. Trading Settings
```go
type TradingConfig struct {
    Enabled            bool
    MaxDailyPositions  int
    MaxConcurrentPos   int
    TradingCost        float64  // USD per trade
    TradingInterval    string   // Primary timeframe
    TradingStrategy    TradingStrategy
    
    // Risk settings per strategy
    LongPNL  *PNL
    ShortPNL *PNL
    
    // Leverage preferences
    PreferLeverageBrackets []int
}
```

### 2. Strategy Parameters
```go
type PNL struct {
    GainPricePercent float64  // Target gain percentage
    LossPricePercent float64  // Maximum loss percentage  
    DesiredProfit    float64  // Absolute profit target
    DesiredLoss      float64  // Absolute loss limit
}
```

## Quality Assurance

### 1. Pre-execution Validation
- **Balance Checks**: Ensure sufficient margin
- **Position Limits**: Validate against exposure limits
- **Price Validation**: Check prices against market data
- **Order Format**: Validate order parameters

### 2. Post-execution Monitoring
- **Fill Confirmation**: Verify order execution
- **Position Reconciliation**: Match expected vs actual positions
- **Risk Compliance**: Monitor position risk metrics
- **Performance Tracking**: Calculate trade P&L

## Future Enhancements

### 1. Advanced Order Types
- **Iceberg Orders**: Large orders with hidden quantity
- **TWAP Orders**: Time-weighted average price execution
- **Conditional Orders**: Complex trigger conditions
- **Multi-leg Strategies**: Spread and arbitrage orders

### 2. Smart Execution
- **Market Impact Modeling**: Minimize price impact
- **Liquidity Analysis**: Choose optimal execution venues
- **Timing Optimization**: Best execution time selection
- **Adaptive Algorithms**: Machine learning for execution

### 3. Risk Management Evolution
- **Dynamic Sizing**: Real-time position size adjustment
- **Correlation Hedging**: Automatic hedge implementation
- **Scenario Analysis**: Stress testing and what-if analysis
- **Real-time Monitoring**: Continuous risk assessment