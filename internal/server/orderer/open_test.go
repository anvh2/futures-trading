package orderer

import (
	"context"
	"testing"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/futures-trading/internal/cache/exchange"
	cachemock "github.com/anvh2/futures-trading/internal/cache/mocks"
	"github.com/anvh2/futures-trading/internal/models"
	telemock "github.com/anvh2/futures-trading/internal/services/telegram/mocks"
	"github.com/anvh2/futures-trading/internal/settings"
	"github.com/stretchr/testify/assert"
)

func TestOpen(t *testing.T) {
	cases := []*struct {
		desc        string
		message     *models.Oscillator
		cache       *cachemock.BasicMock
		exchange    *cachemock.ExchangeMock
		notify      *telemock.NotifyMock
		expectedErr error
	}{
		{
			desc: "LONG",
			message: &models.Oscillator{
				Symbol: "BTCUSDT",
				Stoch: map[string]*models.Stoch{
					"5m": {
						RSI: 15,
						K:   12,
						D:   14,
					},
				},
			},
			cache: &cachemock.BasicMock{
				SetFunc: func(key string, value interface{}) {},
				ExsFunc: func(key string) bool { return false },
			},
			exchange: &cachemock.ExchangeMock{
				GetFunc: func(symbol string) (*exchange.Symbol, error) {
					return &exchange.Symbol{
						Filters: &exchange.Filters{
							{
								MinPrice:   "282.70",
								MaxPrice:   "876544.80",
								FilterType: futures.SymbolFilterTypePrice,
								TickSize:   "0.10",
							},
							{
								FilterType: "LOT_SIZE",
								StepSize:   "0.001",
							},
							{
								FilterType: futures.SymbolFilterTypeMarketLotSize,
								StepSize:   "0.001",
							},
							{
								FilterType: futures.SymbolFilterTypeMinNotional,
								Notional:   "5.0",
							},
							{
								FilterType:     futures.SymbolFilterTypePercentPrice,
								MultiplierDown: "0.5454",
								MultiplierUp:   "1.1000",
							},
						},
					}, nil
				},
			},
			notify: &telemock.NotifyMock{
				PushNotifyFunc: func(ctx context.Context, chatId int64, message string) error {
					return nil
				},
			},
			expectedErr: nil,
		},
	}

	for _, test := range cases {
		t.Run(test.desc, func(t *testing.T) {
			settings := settings.DefaultSettings
			settings.TradingEnabled = true

			order := &Orderer{
				logger:        _loggerTest,
				binance:       _binanceTestnetInst,
				settings:      settings,
				notify:        test.notify,
				cache:         test.cache,
				exchangeCache: test.exchange,
			}

			err := order.open(context.Background(), test.message)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}
