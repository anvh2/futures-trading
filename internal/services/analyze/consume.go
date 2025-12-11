package analyzer

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/anvh2/futures-trading/internal/constants"
	"github.com/anvh2/futures-trading/internal/helpers"
	"github.com/anvh2/futures-trading/internal/models"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func (s *Analyzer) Start() error {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("[Produce] failed to process", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
			}
		}()

		ticker := time.NewTicker(5 * time.Second)

		for {
			select {
			case <-ticker.C:
				for _, symbol := range s.exchangeCache.Symbols() {
					summary, err := s.marketCache.CandleSummary(symbol)
					if err != nil {
						continue
					}

					message := &models.CandleSummary{
						Symbol:  symbol,
						Candles: make(map[string]*models.CandlesData),
					}

					for _, interval := range viper.GetStringSlice("market.intervals") {
						candles, err := summary.Candles(interval)
						if err != nil {
							break
						}

						lastCandles, _ := candles.Tail()
						if err := helpers.CheckCurrentCandle(lastCandles, interval); err != nil {
							s.channel.Get(constants.RetryChannelId) <- &models.RetryMessage{Symbol: symbol, Interval: interval}
							s.logger.Error("[Produce] the last candle is not current candle", zap.String("interval", interval), zap.Any("lastCandle", lastCandles), zap.Error(err))
							break
						}

						candleData := candles.Sorted()
						candlesticks := make([]*models.Candlestick, len(candleData))

						for idx, candle := range candleData {
							result, ok := candle.(*models.Candlestick)
							if ok {
								candlesticks[idx] = result
							}
						}

						if len(candlesticks) > 0 {
							data := summary.SummaryData(interval)
							message.Candles[interval] = &models.CandlesData{
								Candles:    candlesticks,
								CreateTime: data.CreateTime,
								UpdateTime: data.UpdateTime,
							}
						}
					}

					s.worker.SendJob(context.Background(), message)
				}

			case <-s.quitChannel:
				return
			}
		}
	}()

	if err := s.worker.Start(); err != nil {
		return err
	}

	return nil
}
