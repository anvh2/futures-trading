package market

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/anvh2/futures-trading/internal/constants"
	"github.com/anvh2/futures-trading/internal/models"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func (s *Market) Retry() {
	for i := 0; i < 4; i++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					s.logger.Error("[Retry] failed to retry", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
				}
			}()

			for {
				select {
				case message := <-s.channel.Get(constants.RetryChannelId):
					symbol, ok := message.(*models.RetryMessage)
					if !ok {
						continue
					}

					if symbol.Counter == nil {
						symbol.Counter = new(int)
					}

					delay(symbol.Counter)

					resp, err := s.binance.GetCandlesticks(context.Background(), symbol.Symbol, symbol.Interval, viper.GetInt("chart.candles.limit"), 0, 0)
					if err != nil {
						s.logger.Error("[Retry] failed to get klines data", zap.String("symbol", symbol.Symbol), zap.String("interval", symbol.Interval), zap.Error(err))
						s.channel.Get(constants.RetryChannelId) <- symbol
						continue
					}

					for _, e := range resp {
						candle := &models.Candlestick{
							OpenTime:  e.OpenTime,
							CloseTime: e.CloseTime,
							Low:       e.Low,
							High:      e.High,
							Close:     e.Close,
						}

						s.marketCache.UpdateSummary(symbol.Symbol).CreateCandle(symbol.Interval, candle)
					}

					s.logger.Info("[Retry] success", zap.String("symbol", symbol.Symbol), zap.String("interval", symbol.Interval), zap.Int("total", len(resp)))

				case <-s.quitChannel:
					return
				}
			}
		}()
	}
}

func delay(counter *int) {
	*counter++
	if *counter%9 == 0 {
		time.Sleep(time.Minute)
	}

	duration := time.Duration(*counter * 100)
	time.Sleep(duration * time.Millisecond)
}
