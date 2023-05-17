package crawler

import (
	"context"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/anvh2/futures-trading/internal/cache/exchange"
	"github.com/anvh2/futures-trading/internal/constants"
	"github.com/anvh2/futures-trading/internal/models"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func (s *Crawler) Start() error {
	if err := s.fetchExchange(); err != nil {
		return err
	}

	// watch exchange info for the new coin listed
	go func() {
		ticker := time.NewTicker(15 * time.Minute)

		for {
			select {
			case <-ticker.C:
				_ = s.fetchExchange()

			case <-s.quitChannel:
				return
			}
		}
	}()

	if err := s.fetchMarketSummary(); err != nil {
		return err
	}

	s.retrying()
	s.consuming()
	s.notifying()

	return nil
}

func (s *Crawler) fetchExchange() error {
	resp, err := s.binance.GetExchangeInfo(context.Background())
	if err != nil {
		s.logger.Error("[Crawling] failed to get exchange info", zap.Error(err))
		return err
	}

	selected := []*exchange.Symbol{}

	for _, symbol := range resp.Symbols {
		if strings.Contains(symbol.Symbol, "_") {
			continue
		}

		if symbol.MarginAsset == "USDT" {
			if blacklist[symbol.Symbol] {
				continue
			}

			filters := &exchange.Filters{}
			filters.Parse(symbol.Filters)

			selected = append(selected,
				&exchange.Symbol{
					Symbol:      symbol.Symbol,
					Pair:        symbol.Pair,
					Filters:     filters,
					MarginAsset: symbol.MarginAsset,
					BaseAsset:   symbol.BaseAsset,
				},
			)
		}
	}

	s.exchangeCache.Set(selected)
	s.logger.Info("[Crawling] cache symbols success", zap.Int("total", len(selected)))
	return nil
}

func (s *Crawler) fetchMarketSummary() error {
	var (
		wg    = &sync.WaitGroup{}
		total = int32(0)
		start = time.Now()
	)

	for _, interval := range viper.GetStringSlice("market.intervals") {
		wg.Add(1)

		go func(interval string) {
			defer func() {
				if r := recover(); r != nil {
					s.logger.Error("[Crawling] failed to sync market", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
				}
			}()

			defer wg.Done()

			for _, symbol := range s.exchangeCache.Symbols() {
				ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
				defer cancel()

				resp, err := s.binance.GetCandlesticks(ctx, symbol, interval, viper.GetInt("chart.candles.limit"), 0, 0)
				if err != nil {
					s.logger.Error("[Crawling] failed to get klines data", zap.String("symbol", symbol), zap.String("interval", interval), zap.Error(err))
					s.channel.Get(constants.RetryChannelId) <- &models.RetryMessage{Symbol: symbol, Interval: interval}
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

					s.marketCache.UpdateSummary(symbol).CreateCandle(interval, candle)
				}

				atomic.AddInt32(&total, 1)
				s.logger.Info("[Crawling] cache market success", zap.String("symbol", symbol), zap.String("interval", interval), zap.Int("total", len(resp)))
			}

		}(interval)
	}

	wg.Wait()

	s.logger.Info("[Crawling] success to crawl data", zap.Int32("total", total), zap.Float64("take(s)", time.Since(start).Seconds()))
	return nil
}
