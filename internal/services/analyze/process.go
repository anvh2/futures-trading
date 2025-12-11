package analyzer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/anvh2/futures-trading/internal/helpers"
	"github.com/anvh2/futures-trading/internal/libs/talib"
	"github.com/anvh2/futures-trading/internal/models"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func validateMessage(message *models.CandleSummary) error {
	if message == nil {
		return errors.New("analyze: message invalid")
	}
	return nil
}

func (s *Analyzer) process(ctx context.Context, data interface{}) error {
	message := &models.CandleSummary{
		Candles: make(map[string]*models.CandlesData),
	}

	if err := json.Unmarshal([]byte(fmt.Sprint(data)), message); err != nil {
		s.logger.Error("[Process] failed to unmarshal message", zap.Error(err))
		return err
	}

	if err := validateMessage(message); err != nil {
		s.logger.Error("[Process] failed to validate message", zap.Error(err))
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

	var lastUpdate int64
	if message.Candles[s.settings.TradingInterval] != nil {
		lastUpdate = message.Candles[s.settings.TradingInterval].UpdateTime
	}

	msg := fmt.Sprintf("#%s\t\t\t [%0.2f(s) ago]\n\t%s\n", message.Symbol, float64((time.Now().UnixMilli()-lastUpdate))/1000.0, helpers.ResolvePositionSide(oscillator.GetRSI(s.settings.TradingInterval)))

	for interval, stoch := range oscillator.Stoch {
		msg += fmt.Sprintf("\t%03s:\t RSI %2.2f | K %02.2f | D %02.2f\n", strings.ToUpper(interval), stoch.RSI, stoch.K, stoch.D)
	}

	lastSent, existed := s.cache.SetEX(fmt.Sprintf("signal.sent.%s-%s", message.Symbol, s.settings.TradingInterval), time.Now().UnixMilli())
	if existed && time.Now().Before(time.UnixMilli(lastSent.(int64)).Add(10*time.Minute)) {
		return errors.New("analyze: signal already sent")
	}

	expiration, _ := time.ParseDuration(s.settings.TradingInterval)
	if err := s.queue.Push(oscillator, expiration); err != nil {
		s.logger.Error("[Process] failed to push queue", zap.Error(err))
	}

	err := s.notify.PushNotify(ctx, viper.GetInt64("notify.channels.futures_announcement"), msg)
	if err != nil {
		s.logger.Error("[Process] failed to push notification", zap.Error(err))
		return err
	}

	s.logger.Info("[Process] analyze success, end process", zap.String("symbol", message.Symbol))
	return nil
}
