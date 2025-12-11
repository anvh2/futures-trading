package orderer

import (
	"context"
	"errors"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/futures-trading/internal/helpers"
	"github.com/anvh2/futures-trading/internal/libs/talib"
	"github.com/anvh2/futures-trading/internal/models"
	"github.com/anvh2/futures-trading/internal/services/settings"
)

func (s *Orderer) create(ctx context.Context, symbol string, stoch *models.Stoch) ([]*models.Order, error) {
	if stoch == nil {
		return nil, errors.New("orders: empty stoch")
	}

	if !talib.WithinRangeBound(stoch, talib.RangeBoundReadyTrade) {
		return nil, errors.New("orders: indicator not ready to trade")
	}

	var (
		sideType  futures.SideType
		closeSide futures.SideType
	)

	positionSide, err := talib.ResolvePositionSide(stoch, talib.RangeBoundReadyTrade)
	if err != nil {
		return nil, err
	}

	switch positionSide {
	case futures.PositionSideTypeShort:
		sideType = futures.SideTypeSell
		closeSide = futures.SideTypeBuy
	case futures.PositionSideTypeLong:
		sideType = futures.SideTypeBuy
		closeSide = futures.SideTypeSell
	}

	price, err := s.appraise(ctx, symbol, positionSide)

	exchange, err := s.exchangeCache.Get(symbol)
	if err != nil {
		return nil, err
	}

	priceFilter, err := exchange.GetPriceFilter()
	if err != nil {
		return nil, err
	}

	lotFilter, err := exchange.GetLotSizeFilter()
	if err != nil {
		return nil, err
	}

	var orders = []*models.Order{}

	switch s.settings.TradingStrategy {
	case settings.TradingStrategyInstantNoodles:
		orders = []*models.Order{
			{
				Symbol:           symbol,
				Side:             sideType,
				PositionSide:     positionSide,
				OrderType:        futures.OrderTypeLimit,
				TimeInForce:      futures.TimeInForceTypeGTC,
				Quantity:         helpers.AlignQuantityToString(price.Quantity, lotFilter.StepSize),
				Price:            helpers.AlignPriceToString(price.Entry, priceFilter.TickSize),
				WorkingType:      futures.WorkingTypeMarkPrice,
				NewOrderRespType: futures.NewOrderRespTypeRESULT,
			},
			// take profile
			{
				Symbol:           symbol,
				Side:             closeSide,
				PositionSide:     positionSide,
				OrderType:        futures.OrderTypeTakeProfitMarket,
				TimeInForce:      futures.TimeInForceTypeGTC,
				Quantity:         helpers.AlignQuantityToString(price.Quantity, lotFilter.StepSize),
				StopPrice:        helpers.AlignPriceToString(price.Profit, priceFilter.TickSize),
				WorkingType:      futures.WorkingTypeMarkPrice,
				NewOrderRespType: futures.NewOrderRespTypeRESULT,
			},
			// // stop loss
			// {
			// 	Symbol:           symbol,
			// 	Side:             closeSide,
			// 	PositionSide:     positionSide,
			// 	OrderType:        futures.OrderTypeStopMarket,
			// 	TimeInForce:      futures.TimeInForceTypeGTC,
			// 	Quantity:         helpers.AlignQuantityToString(price.Quantity, lotFilter.StepSize),
			// 	StopPrice:        helpers.AlignPriceToString(price.Loss, priceFilter.TickSize),
			// 	WorkingType:      futures.WorkingTypeMarkPrice,
			// 	NewOrderRespType: futures.NewOrderRespTypeRESULT,
			// },
		}

	case settings.TradingStrategyDollarCostAveraging:
		orders = []*models.Order{
			{
				Symbol:           symbol,
				Side:             sideType,
				PositionSide:     positionSide,
				OrderType:        futures.OrderTypeLimit,
				TimeInForce:      futures.TimeInForceTypeGTC,
				Quantity:         helpers.AlignQuantityToString(calculateQuantity(price.Entry, 30), lotFilter.StepSize),
				Price:            helpers.AlignPriceToString(price.Entry, priceFilter.TickSize),
				WorkingType:      futures.WorkingTypeMarkPrice,
				NewOrderRespType: futures.NewOrderRespTypeRESULT,
			},
			{
				Symbol:           symbol,
				Side:             sideType,
				PositionSide:     positionSide,
				OrderType:        futures.OrderTypeLimit,
				TimeInForce:      futures.TimeInForceTypeGTC,
				Quantity:         helpers.AlignQuantityToString(calculateQuantity(price.Entry*1.03, 40), lotFilter.StepSize),
				Price:            helpers.AlignPriceToString(price.Entry*1.03, priceFilter.TickSize),
				WorkingType:      futures.WorkingTypeMarkPrice,
				NewOrderRespType: futures.NewOrderRespTypeRESULT,
			},
			{
				Symbol:           symbol,
				Side:             sideType,
				PositionSide:     positionSide,
				OrderType:        futures.OrderTypeLimit,
				TimeInForce:      futures.TimeInForceTypeGTC,
				Quantity:         helpers.AlignQuantityToString(calculateQuantity(price.Entry*1.03*1.03, 50), lotFilter.StepSize),
				Price:            helpers.AlignPriceToString(price.Entry*1.03*1.03, priceFilter.TickSize),
				WorkingType:      futures.WorkingTypeMarkPrice,
				NewOrderRespType: futures.NewOrderRespTypeRESULT,
			},
			// take profile
			{
				Symbol:           symbol,
				Side:             closeSide,
				PositionSide:     positionSide,
				OrderType:        futures.OrderTypeTakeProfitMarket,
				TimeInForce:      futures.TimeInForceTypeGTC,
				Quantity:         helpers.AlignQuantityToString(calculateStopQuantity(price.Entry, 120), lotFilter.StepSize),
				StopPrice:        helpers.AlignPriceToString(price.Profit, priceFilter.TickSize),
				WorkingType:      futures.WorkingTypeMarkPrice,
				NewOrderRespType: futures.NewOrderRespTypeRESULT,
			},
			// stop loss
			{
				Symbol:           symbol,
				Side:             closeSide,
				PositionSide:     positionSide,
				OrderType:        futures.OrderTypeStopMarket,
				TimeInForce:      futures.TimeInForceTypeGTC,
				Quantity:         helpers.AlignQuantityToString(calculateStopQuantity(price.Entry, 120), lotFilter.StepSize),
				StopPrice:        helpers.AlignPriceToString(price.Loss, priceFilter.TickSize),
				WorkingType:      futures.WorkingTypeMarkPrice,
				NewOrderRespType: futures.NewOrderRespTypeRESULT,
			},
		}
	}

	return orders, nil
}

func calculateQuantity(price, amount float64) float64 {
	return amount / price
}

func calculateStopQuantity(price float64, totalAmount float64) float64 {
	return totalAmount / price
}
