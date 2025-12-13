# Market Service Specification

## Overview
The Market Service analyzes market structure, identifies support/resistance levels, determines trends, and provides price action insights for trading decisions. It serves as the foundation for technical analysis in the trading system.

## Architecture
- **Package**: `internal/services/market`
- **Type**: Analysis service with caching capabilities
- **Dependencies**: Candlestick data from market cache
- **Design Pattern**: Strategy pattern with multiple analysis algorithms

## Core Components

### 1. Market Structure Analyzer
```go
type StructureAnalyzer struct {
    minSwingSize float64 // Minimum swing size as % of price (default: 1.0%)
    lookback     int     // Periods to look back for structure (default: 50)
}

type MarketStructure struct {
    Trend             TrendType
    TrendStrength     float64   // 0-100 strength indicator
    SupportLevels     []float64 // Array of support price levels
    ResistanceLevels  []float64 // Array of resistance price levels
    KeyLevel          float64   // Most significant level
    BreakoutPotential float64   // 0-100 probability of breakout
    IsRangebound      bool      // True if market is range-bound
    SwingHighs        []SwingPoint
    SwingLows         []SwingPoint
}

type SwingPoint struct {
    Price     float64
    Index     int
    Timestamp int64
    Significance float64 // 0-100, how important this level is
}
```

### 2. Trend Analysis Components
```go
type TrendType int

const (
    UPTREND TrendType = iota
    DOWNTREND
    SIDEWAYS
)

type TrendAnalysis struct {
    Direction      TrendType
    Strength       float64 // 0-100
    Duration       int     // Periods in current trend
    StartPrice     float64
    CurrentSlope   float64 // Price change per period
    Volatility     float64 // ATR-based volatility
}
```

## Core Analysis Functions

### 1. Structure Analysis (`AnalyzeStructure`)
**Purpose**: Comprehensive market structure analysis from candlestick data

**Process**:
1. **Swing Point Detection**: Identify significant highs and lows
2. **Support/Resistance Mapping**: Extract horizontal levels
3. **Trend Determination**: Analyze overall market direction
4. **Range Detection**: Identify consolidation patterns
5. **Breakout Assessment**: Evaluate breakout potential

**Inputs**: Array of candlestick data (minimum 50 periods)
**Output**: Complete MarketStructure with all analysis components

### 2. Swing Point Detection (`findSwingPoints`)
**Algorithm**: 
- Look for local extremes with minimum percentage move
- Validate significance using volume and time
- Filter noise using volatility-adjusted thresholds

```go
func (sa *StructureAnalyzer) findSwingPoints(candles []*models.Candlestick) ([]SwingPoint, []SwingPoint) {
    var swingHighs, swingLows []SwingPoint
    
    for i := 2; i < len(candles)-2; i++ {
        current := candles[i]
        
        // Check for swing high
        if isSwingHigh(candles, i, sa.minSwingSize) {
            swingHighs = append(swingHighs, SwingPoint{
                Price: current.High,
                Index: i,
                Timestamp: current.CloseTime,
                Significance: calculateSignificance(candles, i),
            })
        }
        
        // Check for swing low
        if isSwingLow(candles, i, sa.minSwingSize) {
            swingLows = append(swingLows, SwingPoint{
                Price: current.Low,
                Index: i,
                Timestamp: current.CloseTime,
                Significance: calculateSignificance(candles, i),
            })
        }
    }
    
    return swingHighs, swingLows
}
```

### 3. Support/Resistance Identification (`identifyLevels`)
**Algorithm**:
- Cluster nearby swing points using price proximity
- Weight levels by touch count and volume
- Rank levels by significance and recency

```go
func (sa *StructureAnalyzer) identifyLevels(swingPoints []SwingPoint) []float64 {
    // Group nearby points (within 0.5% price range)
    clusters := clusterByPrice(swingPoints, 0.005)
    
    var levels []float64
    for _, cluster := range clusters {
        // Calculate weighted average price for cluster
        level := calculateClusterLevel(cluster)
        levels = append(levels, level)
    }
    
    // Sort by significance (touch count * volume * recency)
    return rankLevelsBySignificance(levels)
}
```

### 4. Trend Analysis (`analyzeTrend`)
**Multi-method Approach**:
1. **Higher Highs/Higher Lows**: Classic trend definition
2. **Moving Average Slopes**: Multiple timeframe alignment
3. **Linear Regression**: Statistical trend strength
4. **Momentum Analysis**: Rate of change evaluation

```go
func (sa *StructureAnalyzer) analyzeTrend(candles []*models.Candlestick) TrendAnalysis {
    // Method 1: Higher highs/Higher lows analysis
    hhHl := analyzeSwingStructure(candles)
    
    // Method 2: Moving average slope
    maSlope := calculateMASlope(candles, 20) // 20-period MA
    
    // Method 3: Linear regression
    slope, r2 := linearRegression(extractClosePrices(candles))
    
    // Method 4: Momentum analysis
    momentum := calculateMomentum(candles, 10) // 10-period momentum
    
    // Combine methods for final trend determination
    return combineTrendAnalysis(hhHl, maSlope, slope, r2, momentum)
}
```

## Key Analysis Algorithms

### 1. Swing Point Validation
```go
func isSwingHigh(candles []*models.Candlestick, index int, minMove float64) bool {
    current := candles[index]
    left1, left2 := candles[index-1], candles[index-2]
    right1, right2 := candles[index+1], candles[index+2]
    
    // Must be higher than surrounding candles
    if current.High <= left1.High || current.High <= left2.High {
        return false
    }
    if current.High <= right1.High || current.High <= right2.High {
        return false
    }
    
    // Must meet minimum move threshold
    avgPrice := (left2.High + left1.High) / 2
    movePercent := (current.High - avgPrice) / avgPrice
    
    return movePercent >= minMove/100
}
```

### 2. Level Clustering Algorithm
```go
func clusterByPrice(points []SwingPoint, tolerance float64) [][]SwingPoint {
    var clusters [][]SwingPoint
    used := make(map[int]bool)
    
    for i, point := range points {
        if used[i] {
            continue
        }
        
        cluster := []SwingPoint{point}
        used[i] = true
        
        // Find nearby points
        for j := i + 1; j < len(points); j++ {
            if used[j] {
                continue
            }
            
            // Check if within tolerance range
            if math.Abs(point.Price-points[j].Price)/point.Price <= tolerance {
                cluster = append(cluster, points[j])
                used[j] = true
            }
        }
        
        clusters = append(clusters, cluster)
    }
    
    return clusters
}
```

### 3. Breakout Potential Assessment
```go
func (sa *StructureAnalyzer) assessBreakoutPotential(structure *MarketStructure, currentPrice float64) float64 {
    score := 0.0
    
    // 1. Distance to key levels
    if len(structure.ResistanceLevels) > 0 {
        nearestResistance := structure.ResistanceLevels[0]
        distanceToResistance := (nearestResistance - currentPrice) / currentPrice
        
        // Closer to resistance = higher breakout potential
        if distanceToResistance < 0.02 { // Within 2%
            score += 40
        } else if distanceToResistance < 0.05 { // Within 5%
            score += 20
        }
    }
    
    // 2. Trend strength contribution
    if structure.Trend == UPTREND {
        score += structure.TrendStrength * 0.3
    }
    
    // 3. Range compression
    if structure.IsRangebound {
        rangeSize := structure.ResistanceLevels[0] - structure.SupportLevels[0]
        compressionRatio := rangeSize / currentPrice
        
        // Smaller range = higher compression = higher breakout potential
        if compressionRatio < 0.05 { // Range < 5% of price
            score += 30
        } else if compressionRatio < 0.10 { // Range < 10% of price
            score += 15
        }
    }
    
    return math.Min(100, score)
}
```

## Market State Detection

### 1. Range-bound Market Identification
```go
func (sa *StructureAnalyzer) isRangebound(swingHighs, swingLows []SwingPoint) bool {
    if len(swingHighs) < 2 || len(swingLows) < 2 {
        return false
    }
    
    // Check if recent highs are roughly equal
    recentHighs := getRecentPoints(swingHighs, 3)
    highsEqual := arePointsEqual(recentHighs, 0.02) // 2% tolerance
    
    // Check if recent lows are roughly equal
    recentLows := getRecentPoints(swingLows, 3)
    lowsEqual := arePointsEqual(recentLows, 0.02)
    
    // Must have clear range (highs above lows with margin)
    avgHigh := averagePrice(recentHighs)
    avgLow := averagePrice(recentLows)
    rangeMargin := (avgHigh - avgLow) / avgLow
    
    return highsEqual && lowsEqual && rangeMargin > 0.03 // 3% minimum range
}
```

### 2. Trend Strength Calculation
```go
func calculateTrendStrength(candles []*models.Candlestick, trend TrendType) float64 {
    // Multiple strength indicators
    var strength float64
    
    // 1. Consistency of higher highs/higher lows (or lower lows/lower highs)
    consistency := calculateTrendConsistency(candles, trend)
    strength += consistency * 0.4
    
    // 2. Moving average alignment
    maAlignment := calculateMAAlignment(candles, []int{10, 20, 50})
    strength += maAlignment * 0.3
    
    // 3. Volume confirmation
    volumeConfirmation := calculateVolumeConfirmation(candles, trend)
    strength += volumeConfirmation * 0.2
    
    // 4. Price momentum
    momentum := calculatePriceMomentum(candles, 14)
    strength += momentum * 0.1
    
    return math.Min(100, strength)
}
```

## Integration with Trading System

### 1. Support/Resistance for Entry/Exit
```go
func (ms *MarketStructure) GetNearestSupport(currentPrice float64) float64 {
    for i := len(ms.SupportLevels) - 1; i >= 0; i-- {
        if ms.SupportLevels[i] < currentPrice {
            return ms.SupportLevels[i]
        }
    }
    return 0
}

func (ms *MarketStructure) GetNearestResistance(currentPrice float64) float64 {
    for _, level := range ms.ResistanceLevels {
        if level > currentPrice {
            return level
        }
    }
    return 0
}
```

### 2. Breakout Confirmation
```go
func (ms *MarketStructure) IsBreakout(currentPrice, previousPrice float64) (bool, string) {
    // Check resistance breakout
    for _, resistance := range ms.ResistanceLevels {
        if previousPrice <= resistance && currentPrice > resistance {
            return true, "bullish_breakout"
        }
    }
    
    // Check support breakdown
    for _, support := range ms.SupportLevels {
        if previousPrice >= support && currentPrice < support {
            return true, "bearish_breakdown"
        }
    }
    
    return false, ""
}
```

## Performance Optimizations

### 1. Caching Strategy
```go
type MarketCache struct {
    structures map[string]*MarketStructure // symbol -> structure
    lastUpdate map[string]int64            // symbol -> timestamp
    ttl        time.Duration               // cache time-to-live
}

func (mc *MarketCache) GetStructure(symbol string, candles []*models.Candlestick) *MarketStructure {
    // Check cache first
    if structure, exists := mc.structures[symbol]; exists {
        if time.Now().Unix()-mc.lastUpdate[symbol] < int64(mc.ttl.Seconds()) {
            return structure
        }
    }
    
    // Calculate and cache
    analyzer := NewStructureAnalyzer()
    structure := analyzer.AnalyzeStructure(candles)
    
    mc.structures[symbol] = structure
    mc.lastUpdate[symbol] = time.Now().Unix()
    
    return structure
}
```

### 2. Incremental Updates
```go
func (sa *StructureAnalyzer) UpdateStructure(existing *MarketStructure, newCandle *models.Candlestick) *MarketStructure {
    // Only recalculate if new candle could affect structure
    if isSignificantCandle(newCandle, existing) {
        return sa.AnalyzeStructure(getRecentCandles(100)) // Re-analyze recent data
    }
    
    // Otherwise, update incrementally
    updated := *existing
    updated = sa.updateLevelsIncremental(&updated, newCandle)
    updated = sa.updateTrendIncremental(&updated, newCandle)
    
    return &updated
}
```

## Quality Assurance

### 1. Level Validation
- **Minimum Touch Count**: Levels must be touched at least 2 times
- **Volume Confirmation**: Significant volume at level touches
- **Time Validation**: Levels must span reasonable time periods
- **Significance Threshold**: Only statistically significant levels

### 2. Trend Validation
- **Multiple Confirmation**: Require agreement from multiple methods
- **Statistical Significance**: R-squared thresholds for trend lines
- **Duration Requirements**: Minimum time in trend before confirmation
- **Strength Thresholds**: Minimum strength scores for trend validity

## Error Handling

### 1. Data Quality Checks
```go
func validateCandleData(candles []*models.Candlestick) error {
    for i, candle := range candles {
        // Check for valid OHLC relationships
        if candle.High < candle.Low {
            return fmt.Errorf("invalid OHLC at index %d", i)
        }
        if candle.Open < 0 || candle.Close < 0 {
            return fmt.Errorf("negative prices at index %d", i)
        }
        // Additional validations...
    }
    return nil
}
```

### 2. Graceful Degradation
- **Insufficient Data**: Return basic structure with available data
- **Calculation Errors**: Use fallback algorithms
- **Edge Cases**: Handle extreme market conditions gracefully
- **Memory Limits**: Implement data size limits and cleanup

## Integration Points

### 1. Input Sources
- **Crawl Service**: Real-time and historical candlestick data
- **External APIs**: Additional market data for validation

### 2. Output Consumers
- **Decision Engine**: Uses structure analysis for trade decisions
- **Risk Service**: Incorporates support/resistance in risk calculations
- **Alert Service**: Breakout and level touch notifications

## Monitoring & Metrics

### 1. Analysis Quality Metrics
- **Level Accuracy**: How often identified levels hold
- **Trend Accuracy**: Percentage of correct trend identifications
- **Breakout Success**: Success rate of identified breakouts
- **Performance Timing**: Analysis computation times

### 2. System Health Monitoring
- **Cache Hit Rates**: Efficiency of structure caching
- **Memory Usage**: Track memory consumption patterns
- **Update Frequency**: Monitor analysis update rates
- **Error Rates**: Track calculation failures and edge cases

## Future Enhancements

### 1. Advanced Pattern Recognition
- **Candlestick Patterns**: Doji, hammer, engulfing patterns
- **Chart Patterns**: Head and shoulders, triangles, flags
- **Elliott Wave**: Wave pattern identification
- **Harmonic Patterns**: Fibonacci-based patterns

### 2. Multi-Timeframe Analysis
- **Cross-timeframe Confirmation**: Align multiple timeframes
- **Fractal Analysis**: Identify self-similar patterns
- **Context Awareness**: Consider broader market context
- **Adaptive Periods**: Dynamic period selection based on volatility