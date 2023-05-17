package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/anvh2/futures-trading/internal/cache/exchange"
	"github.com/stretchr/testify/assert"
)

func TestGetExchangeInfo(t *testing.T) {
	resp, err := test_binanceInst.GetExchangeInfo(context.Background())
	assert.Nil(t, err)

	// fmt.Println(resp)

	for _, symbol := range resp.Symbols {
		if symbol.Symbol == "BTCUSDT" {
			filters := &exchange.Filters{}
			filters.Parse(symbol.Filters)
			b, _ := json.Marshal(filters)
			fmt.Println(string(b))
		}
	}
}

func TestGetCurrentPrice(t *testing.T) {
	resp, err := test_binanceInst.GetCurrentPrice(context.Background(), "BTCUSDT")
	assert.Nil(t, err)
	fmt.Println(resp)
}

func TestListCandlesticks(t *testing.T) {
	resp, err := test_binanceInst.GetCandlesticks(context.TODO(), "BTCUSDT", "1h", 10, 0, 0)
	assert.Nil(t, err)

	fmt.Println(resp)
}
