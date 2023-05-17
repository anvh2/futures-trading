package talib

import (
	"errors"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/futures-trading/internal/models"
	"github.com/spf13/viper"
)

type RangeBound struct {
	RSI *Bound
	K   *Bound
	D   *Bound
}

type Bound struct {
	Lower float64
	Upper float64
}

var (
	RangeBoundRecommend = &RangeBound{
		RSI: &Bound{30, 70},
		K:   &Bound{20, 80},
		D:   &Bound{20, 80},
	}

	RangeBoundReadyTrade = &RangeBound{
		RSI: &Bound{20, 80},
		K:   &Bound{15, 85},
		D:   &Bound{15, 85},
	}
)

func SetUp() {
	switch viper.GetString("server.env") {
	case "dev":
		RangeBoundRecommend = &RangeBound{
			RSI: &Bound{40, 60},
			K:   &Bound{40, 60},
			D:   &Bound{40, 60},
		}

		RangeBoundReadyTrade = &RangeBound{
			RSI: &Bound{40, 60},
			K:   &Bound{40, 60},
			D:   &Bound{40, 60},
		}

	case "prod":
		RangeBoundRecommend = &RangeBound{
			RSI: &Bound{30, 70},
			K:   &Bound{20, 80},
			D:   &Bound{20, 80},
		}

		RangeBoundReadyTrade = &RangeBound{
			RSI: &Bound{20, 80},
			K:   &Bound{15, 85},
			D:   &Bound{15, 85},
		}
	}
}

func WithinRangeBound(stoch *models.Stoch, bound *RangeBound) bool {
	if stoch == nil || bound == nil {
		return false
	}

	if (stoch.RSI >= bound.RSI.Upper || stoch.RSI <= bound.RSI.Lower) &&
		(stoch.K >= bound.K.Upper || stoch.K <= bound.K.Lower) &&
		(stoch.D >= bound.D.Upper || stoch.D <= bound.D.Lower) {

		return true
	}

	return false
}

func ResolvePositionSide(stoch *models.Stoch, bound *RangeBound) (futures.PositionSideType, error) {
	if stoch == nil || bound == nil {
		return "", errors.New("indicator: stoch or bound invalid")
	}

	if (stoch.RSI >= bound.RSI.Upper) && (stoch.K >= bound.K.Upper) && (stoch.D >= bound.D.Upper) {
		return futures.PositionSideTypeShort, nil
	}

	if (stoch.RSI <= bound.RSI.Lower) && (stoch.K <= bound.K.Lower) && (stoch.D <= bound.D.Lower) {
		return futures.PositionSideTypeLong, nil
	}

	return "", errors.New("indicator: not ready to trade")
}
