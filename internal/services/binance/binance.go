package binance

import (
	"net/http"

	"github.com/anvh2/futures-trading/internal/client"
	"github.com/anvh2/futures-trading/internal/logger"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

const (
	_APIURL        = "https://fapi.binance.com"
	_APITestnetURL = "https://testnet.binancefuture.com"
)

type Binance struct {
	limiter *rate.Limiter
	logger  *logger.Logger
	testnet bool
	client  *http.Client
}

func New(logger *logger.Logger, testnet bool) *Binance {
	limiter := rate.NewLimiter(
		rate.Every(viper.GetDuration("binance.rate_limit.duration")),
		viper.GetInt("binance.rate_limit.requests"),
	)
	return &Binance{
		limiter: limiter,
		logger:  logger,
		testnet: testnet,
		client:  client.New(),
	}
}

func (b *Binance) getURL() string {
	if b.testnet {
		return _APITestnetURL
	}
	return _APIURL
}
