package notify

import (
	"github.com/anvh2/futures-trading/internal/config"
	"github.com/anvh2/futures-trading/internal/externals/binance"
	"github.com/anvh2/futures-trading/internal/externals/telegram"
	"github.com/anvh2/futures-trading/internal/libs/logger"
	"github.com/anvh2/futures-trading/internal/libs/queue"
)

type Notifier struct {
	config      config.Config
	logger      *logger.Logger
	binance     *binance.Binance
	notify      *telegram.TelegramBot
	queue       *queue.Queue
	quitChannel chan struct{}
}

func New(
	config config.Config,
	logger *logger.Logger,
	binance *binance.Binance,
	notify *telegram.TelegramBot,
	queue *queue.Queue,
) *Notifier {
	return &Notifier{
		config:      config,
		logger:      logger,
		binance:     binance,
		notify:      notify,
		queue:       queue,
		quitChannel: make(chan struct{}),
	}
}

func (s *Notifier) Start() error {
	if err := s.ListenOrder(); err != nil {
		return err
	}

	if err := s.SendSignal(); err != nil {
		return err
	}

	return nil
}

func (s *Notifier) Stop() {
	close(s.quitChannel)
}
