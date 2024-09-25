package crawler

import (
	"github.com/anvh2/futures-trading/internal/cache"
	"github.com/anvh2/futures-trading/internal/channel"
	"github.com/anvh2/futures-trading/internal/logger"
	"github.com/anvh2/futures-trading/internal/services/binance"
	"github.com/anvh2/futures-trading/internal/services/telegram"
)

var (
	blacklist = map[string]bool{}
)

type Crawler struct {
	logger        *logger.Logger
	binance       *binance.Binance
	notify        *telegram.TelegramBot
	marketCache   cache.Market
	exchangeCache cache.Exchange
	channel       *channel.Channel
	quitChannel   chan struct{}
}

func New(
	logger *logger.Logger,
	binance *binance.Binance,
	notify *telegram.TelegramBot,
	marketCache cache.Market,
	exchangeCache cache.Exchange,
	channel *channel.Channel,
) *Crawler {
	return &Crawler{
		logger:        logger,
		binance:       binance,
		notify:        notify,
		marketCache:   marketCache,
		exchangeCache: exchangeCache,
		channel:       channel,
		quitChannel:   make(chan struct{}),
	}
}

func (s *Crawler) Stop() {
	close(s.quitChannel)
}
