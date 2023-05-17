package models

import (
	"fmt"
	"testing"
)

func TestMarshalChart(t *testing.T) {
	chart := &CandleSummary{
		Symbol: "BTCUSDT",
		Candles: map[string]*CandlesData{
			"1h": {
				Candles: []*Candlestick{
					{
						Low:  "10",
						High: "20",
					},
				},
			},
		},
	}
	fmt.Println(chart.String())
}
