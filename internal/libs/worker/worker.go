package worker

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/anvh2/futures-trading/internal/libs/logger"
	"go.uber.org/zap"
)

type Polling func(ctx context.Context, idx int32) error
type Process func(ctx context.Context, message interface{}) error

type PoolConfig struct {
	NumProcess     int32
	NumPolling     int32
	PollingBackoff time.Duration
}

type Worker struct {
	logger  *logger.Logger
	process Process
	polling Polling
	message chan interface{}
	quit    chan struct{}
	wait    *sync.WaitGroup
	config  *PoolConfig
}

func New(logger *logger.Logger, config *PoolConfig) (*Worker, error) {
	if config == nil {
		return nil, errors.New("worker: config invalid")
	}

	if config.NumPolling == 0 && config.NumProcess == 0 {
		return nil, errors.New("worker: no process")
	}

	if config.PollingBackoff == 0 {
		config.PollingBackoff = time.Second
	}

	buffer := config.NumProcess / 2

	return &Worker{
		logger:  logger,
		message: make(chan interface{}, buffer),
		quit:    make(chan struct{}),
		wait:    &sync.WaitGroup{},
		config:  config,
	}, nil
}

func (w *Worker) WithPolling(polling Polling) *Worker {
	w.polling = polling
	return w
}

func (w *Worker) WithProcess(process Process) *Worker {
	w.process = process
	return w
}

func (w *Worker) Start() error {
	// start worker
	go func() {
		for i := int32(0); i < w.config.NumProcess; i++ {
			w.wait.Add(1)

			go func() {
				defer func() {
					if r := recover(); r != nil {
						w.logger.Error("[Worker] process message failed", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
					}
				}()

				defer w.wait.Done()

				for {
					select {
					case msg, ok := <-w.message:
						if ok {
							w.processMessage(msg)
						}

					case <-w.quit:
						if len(w.message) == 0 {
							return
						}
					}
				}
			}()
		}
	}()

	// start poller
	go func() {
		for i := int32(0); i < w.config.NumPolling; i++ {
			w.wait.Add(1)

			go func(idx int32) {
				defer func() {
					if r := recover(); r != nil {
						w.logger.Error("[Worker] failed to process", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
					}
				}()

				defer w.wait.Done()

				ticker := time.NewTicker(w.config.PollingBackoff)

				for {
					select {
					case _, ok := <-ticker.C:
						if ok {
							w.pollingMessage(idx)
						}

					case <-w.quit:
						return
					}
				}
			}(i)
		}
	}()

	return nil
}

func (w *Worker) Stop() {
	close(w.quit)
	w.wait.Wait()
	close(w.message)
}

func (w *Worker) SendJob(ctx context.Context, message interface{}) {
	w.message <- message
}

func (w *Worker) processMessage(message interface{}) {
	if w.process == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := w.process(ctx, message); err != nil {
		fmt.Println(err)
	}
}

func (w *Worker) pollingMessage(idx int32) {
	if w.polling == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	w.polling(ctx, idx)
}
