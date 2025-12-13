package market

import (
	"context"
	"runtime/debug"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/futures-trading/internal/cache/exchange"
	"github.com/anvh2/futures-trading/internal/cache/market"
	"github.com/anvh2/futures-trading/internal/models"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func (s *Market) Watch() error {
	ready := make(chan bool)

	go func() {
		for _, interval := range viper.GetStringSlice("market.intervals") {
			pair := make(map[string]string, len(s.exchangeCache.Symbols()))
			for _, symbol := range s.exchangeCache.Symbols() {
				pair[symbol] = interval
			}

			go func() {
				defer func() {
					if r := recover(); r != nil {
						s.logger.Error("[CandlesConsumption] failed to start, recovered", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
					}
				}()

				s.processCandlesConsumption(context.Background(), pair)
			}()
		}

		ready <- true
	}()

	<-ready

	return nil
}

func (s *Market) processCandlesConsumption(ctx context.Context, pair map[string]string) (done chan struct{}, stop chan struct{}) {
	done, stop, err := futures.WsCombinedKlineServe(pair, s.handleCandlesConsumption, s.handleConsumeError)
	if err != nil {
		s.logger.Fatal("[CandlesConsumption] failed to connect to klines stream data", zap.Error(err))
		return
	}

	s.logger.Info("[CandlesConsumption] start consume data from websocket")

	select {
	case <-done:
		s.logger.Error("[CandlesConsumption] resume failed connection from done channel")
	case <-stop:
		s.logger.Error("[CandlesConsumption] resume failed connection from stop channel")
	case <-ctx.Done():
		s.logger.Info("[CandlesConsumption] consume finished, quit process")
		return
	}

	s.processCandlesConsumption(ctx, pair)
	return
}

func (s *Market) handleCandlesConsumption(event *futures.WsKlineEvent) {
	_, err := s.exchangeCache.Get(event.Symbol)
	if err == exchange.ErrorSymbolNotFound {
		s.logger.Info("[CandlesConsumption] no need to handle this symbol", zap.String("symbol", event.Symbol))
		return
	}

	chart, err := s.marketCache.CandleSummary(event.Symbol)
	if err == market.ErrorChartNotFound {
		chart = s.marketCache.CreateSummary(event.Symbol)
	}

	candles, err := chart.Candles(event.Kline.Interval)
	if err == market.ErrorCandlesNotFound {
		return
	}

	last, idx := candles.Tail()
	if idx < 0 {
		return
	}

	lastCandle, ok := last.(*models.Candlestick)
	if !ok {
		return
	}

	// update the last candle
	if lastCandle.OpenTime == event.Kline.StartTime &&
		lastCandle.CloseTime == event.Kline.EndTime {

		lastCandle.Close = event.Kline.Close
		lastCandle.High = event.Kline.High
		lastCandle.Low = event.Kline.Low

		chart.UpdateCandle(event.Kline.Interval, idx, lastCandle)
		return
	}

	// create new candle
	candle := &models.Candlestick{
		OpenTime:  event.Kline.StartTime,
		CloseTime: event.Kline.EndTime,
		Low:       event.Kline.Low,
		High:      event.Kline.High,
		Close:     event.Kline.Close,
	}

	chart.CreateCandle(event.Kline.Interval, candle)
}

func (s *Market) handleConsumeError(err error) {
	s.logger.Error("[CandlesConsumption] failed to recieve data", zap.Error(err))
}
