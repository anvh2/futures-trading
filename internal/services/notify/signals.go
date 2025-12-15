package notify

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/anvh2/futures-trading/internal/libs/queue"
	"github.com/anvh2/futures-trading/internal/models"
	"github.com/spf13/viper"
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
				msg, err := s.queue.Consume(context.Background(), "signals", "notify-service")
				if err != nil {
					continue
				}

				if err := s.handleSignal(msg); err != nil {
					s.logger.Error("[SendSignal] failed to handle signal", zap.Error(err))
					msg.Commit(context.Background())
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
	// Try to unmarshal the signal
	signal, ok := msg.Data.(*models.Signal)
	if !ok {
		s.logger.Error("[handleSignal] invalid signal data type")
		return errors.New("invalid signal data type")
	}

	// Validate signal
	if !signal.IsValid() {
		s.logger.Debug("[handleSignal] signal is not valid or expired",
			zap.String("symbol", signal.Symbol))
		return nil
	}

	// Get position side from metadata
	positionSideRaw, _ := signal.GetMetadata("position_side")
	positionSide, _ := positionSideRaw.(string)

	// Calculate time ago from creation
	timeAgo := time.Since(signal.CreatedAt).Seconds()

	// Build signal message
	signalMsg := fmt.Sprintf("#%s\t\t\t [%0.2f(s) ago]\n\t%s\n",
		signal.Symbol,
		timeAgo,
		positionSide)

	// Add indicator values to message
	if rsi, exists := signal.GetIndicatorValue("rsi"); exists {
		signalMsg += fmt.Sprintf("\t%03s:\t RSI %2.2f",
			strings.ToUpper(signal.Interval), rsi)
	}

	if k, exists := signal.GetIndicatorValue("k"); exists {
		signalMsg += fmt.Sprintf(" | K %02.2f", k)
	}

	if d, exists := signal.GetIndicatorValue("d"); exists {
		signalMsg += fmt.Sprintf(" | D %02.2f", d)
	}

	signalMsg += "\n"

	// Add confidence and strength info
	signalMsg += fmt.Sprintf("\tConfidence: %02.1f%% | Strength: %02.1f%% | Action: %s\n",
		signal.Confidence*100,
		signal.Strength*100,
		signal.Action)

	// Check rate limiting to avoid spam (if cache is available)
	// Note: Remove this section if Notifier doesn't have cache field
	cacheKey := fmt.Sprintf("signal.sent.%s-%s", signal.Symbol, signal.Interval)
	lastSent, existed := s.cache.SetEX(cacheKey, time.Now().UnixMilli())
	if existed && time.Now().Before(time.UnixMilli(lastSent.(int64)).Add(10*time.Minute)) {
		s.logger.Debug("[handleSignal] signal already sent recently, skipping",
			zap.String("symbol", signal.Symbol))
		return nil
	}

	// Send notification
	err := s.notify.PushNotify(context.Background(),
		viper.GetInt64("notify.channels.futures_announcement"),
		signalMsg)
	if err != nil {
		s.logger.Error("[handleSignal] failed to push notification",
			zap.Error(err),
			zap.String("symbol", signal.Symbol))
		return err
	}

	s.logger.Info("[handleSignal] signal notification sent successfully",
		zap.String("symbol", signal.Symbol),
		zap.String("action", string(signal.Action)),
		zap.Float64("confidence", signal.Confidence))

	return nil
}
