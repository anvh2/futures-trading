package helpers

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/anvh2/futures-trading/internal/models"
)

func TestCheckCurrentCandle(t *testing.T) {
	data := `{"s":1672414500000,"e":1672414799999,"h":"1.5900","c":"1.5900","l":"1.5900"}`
	candle := &models.Candlestick{}

	json.Unmarshal([]byte(data), candle)

	// candle.OpenTime = time.Now().Add(-4*time.Minute + 59*time.Second).UnixMilli()

	err := CheckCurrentCandle(candle, "5m")
	fmt.Println(err)
}
