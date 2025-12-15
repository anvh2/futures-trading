package decision

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/anvh2/futures-trading/internal/libs/queue"
	"github.com/anvh2/futures-trading/internal/models"
	"go.uber.org/zap"
)

func (de *Maker) HandleSignals() error {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				de.logger.Error("[handleSignals] failed to process", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
			}
		}()

		ticker := time.NewTicker(5 * time.Second)

		for {
			select {
			case <-ticker.C:
				msg, err := de.queue.Consume(context.Background(), "signals", "decision-maker")
				if err != nil {
					continue
				}

				if err := de.handleSignal(msg); err != nil {
					de.logger.Error("[handleSignals] failed to handle signal", zap.Error(err))
					msg.Commit(context.Background())
					continue
				}

				msg.Commit(context.Background())

			case <-de.quitChannel:
				return
			}
		}
	}()

	return nil
}

func (de *Maker) handleSignal(msg *queue.Message) error {
	signal, ok := msg.Data.(*models.Signal)
	if !ok {
		return nil
	}

	stats := de.signal.Stats(signal.Interval)
	if stats[signal.Interval] < 3 { // minimum 3 signals to make a decision
		return nil
	}

	decision := de.signal.Pop(signal.Interval)
	if decision == nil {
		return nil
	}

	sig := models.Signal(*decision)

	final := de.MakeDecision(&sig)
	if final == nil {
		return nil
	}

	if err := de.queue.Push(context.Background(), "decisions", final); err != nil {
		return err
	}

	return nil
}
