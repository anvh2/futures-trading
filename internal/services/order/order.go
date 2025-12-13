package orderer

import (
	"context"
	"log"
	"runtime/debug"
	"time"

	"github.com/anvh2/futures-trading/internal/cache"
	"github.com/anvh2/futures-trading/internal/config"
	"github.com/anvh2/futures-trading/internal/externals/binance"
	"github.com/anvh2/futures-trading/internal/externals/telegram"
	"github.com/anvh2/futures-trading/internal/libs/cache/simple"
	basic "github.com/anvh2/futures-trading/internal/libs/cache/simple"
	"github.com/anvh2/futures-trading/internal/libs/logger"
	"github.com/anvh2/futures-trading/internal/libs/queue"
	"github.com/anvh2/futures-trading/internal/libs/worker"
	"github.com/anvh2/futures-trading/internal/services/settings"
	"go.uber.org/zap"
)

type Orderer struct {
	config        config.Config
	logger        *logger.Logger
	binance       *binance.Binance
	notify        telegram.Notify
	queue         *queue.Queue
	settings      *settings.Settings
	cache         *simple.Cache
	worker        *worker.Worker
	marketCache   cache.Market
	exchangeCache cache.Exchange
	quitChannel   chan struct{}
}

func New(
	config config.Config,
	logger *logger.Logger,
	notify telegram.Notify,
	marketCache cache.Market,
	exchangeCache cache.Exchange,
	queue *queue.Queue,
	settings *settings.Settings,
) *Orderer {
	worker, err := worker.New(logger, &worker.PoolConfig{NumProcess: 1})
	if err != nil {
		log.Fatal("failed to new worker", zap.Error(err))
	}

	orderer := &Orderer{
		config:        config,
		logger:        logger,
		binance:       binance.New(logger, true),
		notify:        notify,
		queue:         queue,
		settings:      settings,
		cache:         basic.NewCache(),
		worker:        worker,
		marketCache:   marketCache,
		exchangeCache: exchangeCache,
		quitChannel:   make(chan struct{}),
	}

	orderer.worker.WithProcess(orderer.open)

	return orderer
}

func (o *Orderer) Start() error {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				o.logger.Error("[Produce] failed to process", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
			}
		}()

		ticker := time.NewTicker(5 * time.Second)

		for {
			select {
			case <-ticker.C:
				msg, err := o.queue.Consume(context.Background(), o.config.Topics.SymbolsTradeIntentTopic, "group1")
				if err != nil {
					continue
				}

				o.worker.SendJob(context.Background(), msg.Data)
				msg.Commit(context.Background())

			case <-o.quitChannel:
				return
			}
		}
	}()

	if err := o.worker.Start(); err != nil {
		return err
	}

	return nil
}

func (o *Orderer) Stop() {
	close(o.quitChannel)
}
