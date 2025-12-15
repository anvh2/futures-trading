package analyzer

import (
	"log"

	"github.com/anvh2/futures-trading/internal/cache"
	"github.com/anvh2/futures-trading/internal/config"
	"github.com/anvh2/futures-trading/internal/externals/telegram"
	"github.com/anvh2/futures-trading/internal/libs/cache/simple"
	"github.com/anvh2/futures-trading/internal/libs/channel"
	"github.com/anvh2/futures-trading/internal/libs/logger"
	"github.com/anvh2/futures-trading/internal/libs/queue"
	"github.com/anvh2/futures-trading/internal/libs/worker"
	"github.com/anvh2/futures-trading/internal/services/settings"
	"github.com/anvh2/futures-trading/internal/services/signal"
	"go.uber.org/zap"
)

type Analyzer struct {
	config        config.Config
	logger        *logger.Logger
	cache         *simple.Cache
	worker        *worker.Worker
	marketCache   cache.Market
	exchangeCache cache.Exchange
	queue         *queue.Queue
	channel       *channel.Channel
	signal        signal.Service
	settings      *settings.Settings
	notify        *telegram.TelegramBot
	quitChannel   chan struct{}
}

func New(
	config config.Config,
	logger *logger.Logger,
	notify *telegram.TelegramBot,
	marketCache cache.Market,
	exchangeCache cache.Exchange,
	queue *queue.Queue,
	channel *channel.Channel,
	signal signal.Service,
	settings *settings.Settings,
) *Analyzer {
	worker, err := worker.New(logger, &worker.PoolConfig{NumProcess: 2})
	if err != nil {
		log.Fatal("failed to new worker", zap.Error(err))
	}

	analyzer := &Analyzer{
		config:        config,
		logger:        logger,
		signal:        signal,
		notify:        notify,
		worker:        worker,
		cache:         simple.NewCache(),
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
