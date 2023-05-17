package orderer

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/futures-trading/internal/cache/exchange"
	cachemock "github.com/anvh2/futures-trading/internal/cache/mocks"
	"github.com/anvh2/futures-trading/internal/models"
	"github.com/anvh2/futures-trading/internal/settings"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	cases := []*struct {
		desc        string
		symbol      string
		stoch       *models.Stoch
		exchange    *cachemock.ExchangeMock
		expectedErr error
	}{
		{
			desc:   "LONG",
			symbol: "BTCUSDT",
			stoch: &models.Stoch{
				RSI: 15,
				K:   12,
				D:   14,
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
			expectedErr: nil,
		},
	}

	for _, test := range cases {
		t.Run(test.desc, func(t *testing.T) {
			order := &Orderer{
				logger:        _loggerTest,
				binance:       _binanceTestnetInst,
				settings:      settings.DefaultSettings,
				exchangeCache: test.exchange,
			}

			orders, err := order.create(context.Background(), test.symbol, test.stoch)
			assert.Equal(t, test.expectedErr, err)
			b, _ := json.Marshal(orders)
			fmt.Println(string(b))
		})
	}
}
