package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/futures-trading/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestListPositionRisk(t *testing.T) {
	resp, err := test_binanceInst.GetPositionRisk(context.Background(), "BTCUSDT")
	assert.Nil(t, err)

	for _, position := range resp {
		fmt.Println(position)
	}
}
func TestListOpenOrders(t *testing.T) {
	resp, err := test_binanceInst.GetOpenOrders(context.Background(), "BTCUSDT")
	assert.Nil(t, err)

	for _, order := range resp {
		fmt.Println(order)
	}
}

func TestCreateOrders(t *testing.T) {
	resp, err := test_binanceInst.OpenOrders(context.Background(), []*models.Order{
		{
			Symbol:           "BNXUSDT",
			Side:             futures.SideTypeSell,
			PositionSide:     futures.PositionSideTypeShort,
			OrderType:        futures.OrderTypeLimit,
			TimeInForce:      futures.TimeInForceTypeGTC,
			Quantity:         "0.1",
			Price:            "170",
			WorkingType:      futures.WorkingTypeMarkPrice,
			NewOrderRespType: futures.NewOrderRespTypeRESULT,
		},
		{
			Symbol:           "BNXUSDT",
			Side:             futures.SideTypeSell,
			PositionSide:     futures.PositionSideTypeShort,
			OrderType:        futures.OrderTypeTakeProfit,
			TimeInForce:      futures.TimeInForceTypeGTC,
			Quantity:         "0.1",
			Price:            "170",
			StopPrice:        "170",
			WorkingType:      futures.WorkingTypeMarkPrice,
			NewOrderRespType: futures.NewOrderRespTypeACK,
		},
		{
			Symbol:           "BNXUSDT",
			Side:             futures.SideTypeSell,
			PositionSide:     futures.PositionSideTypeShort,
			OrderType:        futures.OrderTypeStopMarket,
			TimeInForce:      futures.TimeInForceTypeGTC,
			Quantity:         "0.1",
			StopPrice:        "120",
			WorkingType:      futures.WorkingTypeMarkPrice,
			NewOrderRespType: futures.NewOrderRespTypeACK,
		},
	})
	b, _ := json.Marshal(resp)
	fmt.Println(string(b), err)
}
