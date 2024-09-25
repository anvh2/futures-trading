package analyzer

import (
	"log"

	"github.com/anvh2/futures-trading/internal/cache"
	"github.com/anvh2/futures-trading/internal/cache/basic"
	"github.com/anvh2/futures-trading/internal/channel"
	"github.com/anvh2/futures-trading/internal/libs/queue"
	"github.com/anvh2/futures-trading/internal/logger"
	"github.com/anvh2/futures-trading/internal/services/telegram"
	"github.com/anvh2/futures-trading/internal/settings"
	"github.com/anvh2/futures-trading/internal/worker"
	"go.uber.org/zap"
)

type Analyzer struct {
	logger        *logger.Logger
	cache         cache.Basic
	worker        *worker.Worker
	marketCache   cache.Market
	exchangeCache cache.Exchange
	queue         *queue.Queue
	channel       *channel.Channel
	settings      *settings.Settings
	notify        *telegram.TelegramBot
	quitChannel   chan struct{}
}

func New(
	logger *logger.Logger,
	notify *telegram.TelegramBot,
	marketCache cache.Market,
	exchangeCache cache.Exchange,
	queue *queue.Queue,
	channel *channel.Channel,
	settings *settings.Settings,
) *Analyzer {
	worker, err := worker.New(logger, &worker.PoolConfig{NumProcess: 8})
	if err != nil {
		log.Fatal("failed to new worker", zap.Error(err))
	}

	analyzer := &Analyzer{
		logger:        logger,
		notify:        notify,
		worker:        worker,
		cache:         basic.NewCache(),
		marketCache:   marketCache,
		exchangeCache: exchangeCache,
		channel:       channel,
		queue:         queue,
		settings:      settings,
		quitChannel:   make(chan struct{}),
	}

	analyzer.worker.WithProcess(analyzer.process)
	return analyzer
}

func (s *Analyzer) Stop() {
	close(s.quitChannel)
	s.worker.Stop()
}
