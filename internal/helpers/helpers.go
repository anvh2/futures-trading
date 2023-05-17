package helpers

import (
	"errors"
	"time"

	"github.com/anvh2/futures-trading/internal/models"
)

func ResolvePositionSide(rsi float64) string {
	if rsi >= 70 {
		return "SHORT"
	} else if rsi <= 30 {
		return "LONG"
	}
	return ""
}

func CheckCurrentCandle(candleData interface{}, interval string) error {
	candle, ok := candleData.(*models.Candlestick)
	if !ok {
		return errors.New("candles: invalid data")
	}

	duration, err := time.ParseDuration(interval)
	if err != nil {
		return err
	}

	if time.Now().After(time.UnixMilli(candle.OpenTime).Add(duration)) {
		return errors.New("candles: obsolete")
	}

	return nil
}
