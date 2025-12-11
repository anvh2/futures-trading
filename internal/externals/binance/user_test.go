package binance

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetListenKey(t *testing.T) {
	resp, err := test_binanceInst.GetListenKey(context.Background())
	assert.Nil(t, err)

	fmt.Println(resp)
}
