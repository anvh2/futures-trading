package notify

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/anvh2/futures-trading/internal/libs/queue"
	"go.uber.org/zap"
)

func (s *Notifier) SendSignal() error {
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
				msg, err := s.queue.Consume(context.Background(), s.config.Topics.SymbolsTradeIntentTopic, "group1")
				if err != nil {
					continue
				}

				if err := s.handleSignal(msg); err != nil {
					s.logger.Error("[SendSignal] failed to handle signal", zap.Error(err))
					continue
				}

				msg.Commit(context.Background())

			case <-s.quitChannel:
				return
			}
		}
	}()

	return nil
}

func (s *Notifier) handleSignal(msg *queue.Message) error {
	// var lastUpdate int64
	// if message.Candles[s.settings.TradingInterval] != nil {
	// 	lastUpdate = message.Candles[s.settings.TradingInterval].UpdateTime
	// }

	// signal := fmt.Sprintf("#%s\t\t\t [%0.2f(s) ago]\n\t%s\n", message.Symbol, float64((time.Now().UnixMilli()-lastUpdate))/1000.0, helpers.ResolvePositionSide(oscillator.GetRSI(s.settings.TradingInterval)))

	// for interval, stoch := range oscillator.Stoch {
	// 	signal += fmt.Sprintf("\t%03s:\t RSI %2.2f | K %02.2f | D %02.2f\n", strings.ToUpper(interval), stoch.RSI, stoch.K, stoch.D)
	// }

	// lastSent, existed := s.cache.SetEX(fmt.Sprintf("signal.sent.%s-%s", message.Symbol, s.settings.TradingInterval), time.Now().UnixMilli())
	// if existed && time.Now().Before(time.UnixMilli(lastSent.(int64)).Add(10*time.Minute)) {
	// 	return errors.New("analyze: signal already sent")
	// }

	// expiration, _ := time.ParseDuration(s.settings.TradingInterval)
	// if err := s.queue.Push(oscillator, expiration); err != nil {
	// 	s.logger.Error("[Process] failed to push queue", zap.Error(err))
	// }

	// err := s.notify.PushNotify(ctx, viper.GetInt64("notify.channels.futures_announcement"), signal)
	// if err != nil {
	// 	s.logger.Error("[Process] failed to push notification", zap.Error(err))
	// 	return err
	// }

	return nil
}
