package orderer

import (
	"context"
	"fmt"
	"testing"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/futures-trading/internal/settings"
	"github.com/stretchr/testify/assert"
)

func TestAppraise(t *testing.T) {
	cases := []*struct {
		desc         string
		symbol       string
		positionSide futures.PositionSideType
		expectedErr  error
	}{
		{
			desc:         "LONG",
			symbol:       "BTCUSDT",
			positionSide: futures.PositionSideTypeLong,
			expectedErr:  nil,
		},
		{
			desc:         "SHORT",
			symbol:       "BTCUSDT",
			positionSide: futures.PositionSideTypeShort,
			expectedErr:  nil,
		},
	}

	for _, test := range cases {
		t.Run(test.desc, func(t *testing.T) {
			order := &Orderer{
				logger:   _loggerTest,
				binance:  _binanceTestnetInst,
				settings: settings.DefaultSettings,
			}
			price, err := order.appraise(context.Background(), test.symbol, test.positionSide)
			assert.Equal(t, test.expectedErr, err)
			fmt.Println(price.String())
		})
	}
}
