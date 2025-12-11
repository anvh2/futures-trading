package orderer

import (
	"context"
	"errors"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/futures-trading/internal/helpers"
	"github.com/anvh2/futures-trading/internal/models"
	"github.com/anvh2/futures-trading/internal/services/settings"
	"go.uber.org/zap"
)

func (s *Orderer) appraise(ctx context.Context, symbol string, positionSide futures.PositionSideType) (*models.Price, error) {
	leverageBrackets, err := s.binance.GetLeverageBracket(ctx, symbol)
	if err != nil {
		s.logger.Error("[Appraise] faile to get leverage bracket", zap.String("symbol", symbol), zap.Error(err))
		return nil, err
	}

	leverage := s.settings.GetPreferLeverage(leverageBrackets)

	symbolPrice, err := s.binance.GetCurrentPrice(ctx, symbol)
	if err != nil {
		s.logger.Error("[Appraise] failed to get current symbol price", zap.String("symbol", symbol), zap.Error(err))
		return nil, err
	}

	candles, err := s.binance.GetCandlesticks(ctx, symbol, s.settings.TradingInterval, 2, 0, 0)
	if err != nil {
		s.logger.Error("[Appraise] failed to get candles", zap.String("symbol", symbol), zap.Error(err))
		return nil, err
	}

	if len(candles) < 2 {
		return nil, errors.New("orders: len of candles not enough")
	}

	price := &models.Price{}

	// switch trading_strategy
	switch s.settings.TradingStrategy {
	case settings.TradingStrategyInstantNoodles:
		switch positionSide {
		case futures.PositionSideTypeShort:
			price.Entry = helpers.MinFloat(candles[0].High, candles[1].High)

			current := helpers.StringToFloat(symbolPrice.Price)
			if price.Entry < current {
				price.Entry = current * 1.01
			}

			price.Quantity = s.settings.TradingCost * float64(leverage) / price.Entry
			price.Profit = price.Entry - s.settings.ShortPNL.DesiredProfit/price.Quantity
			price.Loss = price.Entry - s.settings.ShortPNL.DesiredLoss/price.Quantity

		case futures.PositionSideTypeLong:
			price.Entry = helpers.MinFloat(candles[0].Low, candles[1].Low)

			current := helpers.StringToFloat(symbolPrice.Price)
			if price.Entry > current {
				price.Entry = current * 0.99
			}

			price.Quantity = s.settings.TradingCost * float64(leverage) / price.Entry
			price.Profit = s.settings.LongPNL.DesiredProfit/price.Quantity + price.Entry
			price.Loss = s.settings.LongPNL.DesiredLoss/price.Quantity + price.Entry
		}

	case settings.TradingStrategyDollarCostAveraging:
		return nil, errors.New("orders: not implement for dca strategy")
	}

	return price, nil
}
