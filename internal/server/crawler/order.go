package crawler

import (
	"context"
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func (s *Crawler) notifying() error {
	ready := make(chan error)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("[OrderConsumption] failed", zap.Any("error", r), zap.Any("debug", debug.Stack()))
			}
		}()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if err := s.processOrderConsumption(ctx); err != nil {
			ready <- err
			return
		}

		ready <- nil
	}()

	return nil
}

func (s *Crawler) processOrderConsumption(ctx context.Context) error {
	listenKey, err := s.binance.GetListenKey(ctx)
	if err != nil {
		s.logger.Error("[OrderConsumption] failed to get listen key", zap.Error(err))
		return err
	}

	s.logger.Info("[OrderConsumption] start consume data from websocket")

	done, stop, err := futures.WsUserDataServe(listenKey, func(event *futures.WsUserDataEvent) {
		switch event.Event {
		case futures.UserDataEventTypeOrderTradeUpdate:
			s.handleOrderConsumption(ctx, event)

		case futures.UserDataEventTypeListenKeyExpired:
			s.logger.Info("[OrderConsumption] reconsume data", zap.Int("goroutines", runtime.NumGoroutine()))
			s.processOrderConsumption(ctx)
		}

	}, func(err error) {
		s.logger.Error("[OrderConsumption] failed to consume user data", zap.Error(err))
	})

	if err != nil {
		s.logger.Error("[OrderConsumption] failed to new user data stream", zap.Error(err))
		return err
	}

	select {
	case <-done:
		s.logger.Error("[OrderConsumption] resume failed connection from done channel")
	case <-stop:
		s.logger.Error("[OrderConsumption] resume failed connection from stop channel")
	case <-ctx.Done():
		s.logger.Info("[OrderConsumption] consume finished, quit process")
		return nil
	}

	s.processOrderConsumption(ctx)
	return nil
}

func (s *Crawler) handleOrderConsumption(ctx context.Context, event *futures.WsUserDataEvent) {
	order := event.OrderTradeUpdate

	msg := fmt.Sprintf("%s #%s: %s | Price: %s | Quantity: %s | Status: %s", order.PositionSide, order.Symbol, order.Side, order.StopPrice, order.OriginalQty, order.Status)
	err := s.notify.PushNotify(ctx, viper.GetInt64("notify.channels.futures_announcement"), msg)
	if err != nil {
		s.logger.Error("[OrderConsumption] failed to push notification", zap.String("message", msg), zap.Error(err))
		return
	}

	s.logger.Info("[OrderConsumption] success", zap.String("msg", msg))
}
