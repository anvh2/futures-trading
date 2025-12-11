package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLeverageBracket(t *testing.T) {
	resp, err := test_binanceInst.GetLeverageBracket(context.Background(), "BTCUSDT")
	assert.Nil(t, err)
	b, _ := json.Marshal(resp)
	fmt.Println(string(b))
}
