package analyzer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/anvh2/futures-trading/internal/helpers"
	"github.com/anvh2/futures-trading/internal/libs/talib"
	"github.com/anvh2/futures-trading/internal/models"
	"go.uber.org/zap"
)

func (s *Analyzer) process(ctx context.Context, data interface{}) error {
	message := &models.CandleSummary{
		Candles: make(map[string]*models.CandlesData),
	}

	if err := json.Unmarshal([]byte(fmt.Sprint(data)), message); err != nil {
		s.logger.Error("[Process] failed to unmarshal message", zap.Error(err))
		return err
	}

	oscillator := &models.Oscillator{
		Symbol: message.Symbol,
		Stoch:  make(map[string]*models.Stoch),
	}

	for interval, candles := range message.Candles {
		if candles == nil {
			continue
		}

		low := make([]float64, len(candles.Candles))
		high := make([]float64, len(candles.Candles))
		close := make([]float64, len(candles.Candles))

		for idx, candle := range candles.Candles {
			l, _ := strconv.ParseFloat(candle.Low, 64)
			low[idx] = l

			h, _ := strconv.ParseFloat(candle.High, 64)
			high[idx] = h

			c, _ := strconv.ParseFloat(candle.Close, 64)
			close[idx] = c
		}

		_, rsi := talib.RSIPeriod(14, close)
		k, d, _ := talib.KDJ(9, 3, 3, high, low, close)

		stoch := &models.Stoch{
			RSI: rsi[len(rsi)-1],
			K:   k[len(k)-1],
			D:   d[len(d)-1],
		}

		oscillator.Stoch[interval] = stoch
	}

	if oscillator.Stoch[s.settings.TradingInterval] == nil {
		return errors.New("analyze: trading interval notfound")
	}

	if !talib.WithinRangeBound(oscillator.Stoch[s.settings.TradingInterval], talib.RangeBoundRecommend) {
		return errors.New("analyze: not ready to trade")
	}

	signal := &models.Signal{
		Symbol:     oscillator.Symbol,
		Type:       models.SignalTypeEntry,
		Action:     models.SignalAction(helpers.ResolvePositionSide(oscillator.GetRSI(s.settings.TradingInterval))),
		Confidence: s.calculateConfidence(oscillator),
		Strength:   s.calculateStrength(oscillator),
		Interval:   s.settings.TradingInterval,
		Strategy:   "oscillator-analysis",
		Indicators: map[string]float64{
			"rsi": oscillator.GetRSI(s.settings.TradingInterval),
		},
		Metadata: map[string]interface{}{
			"oscillator":    oscillator,
			"position_side": helpers.ResolvePositionSide(oscillator.GetRSI(s.settings.TradingInterval)),
		},
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(10 * time.Minute),
		IsActive:  true,
	}

	s.signal.Ingest(signal)

	if err := s.queue.Push(context.Background(), "signals", signal); err != nil {
		s.logger.Error("[Process] failed to push signal to decision queue", zap.Error(err))
		return err
	}

	return nil
}
