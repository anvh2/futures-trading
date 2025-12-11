package orderer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/adshao/go-binance/v2/futures"
	binancew "github.com/anvh2/futures-trading/internal/externals/binance"
	"github.com/anvh2/futures-trading/internal/helpers"
	"github.com/anvh2/futures-trading/internal/models"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func validateOscillator(message *models.Oscillator) error {
	if message == nil {
		return errors.New("trading: message invalid")
	}
	return nil
}

func (s *Orderer) open(ctx context.Context, data interface{}) error {
	if !s.settings.TradingEnabled {
		return errors.New("trading: trading is disabled")
	}

	oscillator := &models.Oscillator{}

	if err := json.Unmarshal([]byte(fmt.Sprint(data)), oscillator); err != nil {
		s.logger.Error("[OpenOrders] failed to unmarshal oscillator", zap.Error(err))
		return err
	}

	if err := validateOscillator(oscillator); err != nil {
		return err
	}

	if s.cache.Exs(oscillator.Symbol) {
		s.logger.Info("[OpenOrders] symbol is processing", zap.String("symbol", oscillator.Symbol))
		return nil
	}

	openPositions, err := s.binance.GetPositionRisk(ctx, "")
	if err != nil {
		s.logger.Error("[OpenOrders] failed to get positions", zap.String("symbol", oscillator.Symbol), zap.Error(err))
		return err
	}

	if positionExisted(openPositions, oscillator.Symbol) {
		s.logger.Info("[OpenOrders] position existed", zap.String("symbol", oscillator.Symbol), zap.Any("openPositions", openPositions))
		return nil
	}

	openOrders, err := s.binance.GetOpenOrders(ctx, "")
	if err != nil {
		s.logger.Error("[OpenOrders] failed to get orders", zap.String("symbol", oscillator.Symbol), zap.Error(err))
		return err
	}

	if orderExisted(openOrders, oscillator.Symbol) {
		s.logger.Info("[OpenOrders] order existed", zap.String("symbol", oscillator.Symbol), zap.Any("orders", openOrders))
		return nil
	}

	if err := s.checkOrderAndPositionQuantity(openOrders, openPositions); err != nil {
		s.logger.Info("[OpenOrders] check quantity error", zap.String("symbol", oscillator.Symbol), zap.Error(err))
		return nil
	}

	s.logger.Info("orders and positions", zap.Any("positions", openPositions), zap.Any("orders", openOrders))

	orders, err := s.create(ctx, oscillator.Symbol, oscillator.Stoch[s.settings.TradingInterval])
	if err != nil {
		s.logger.Info("[OpenOrders] failed to make orders", zap.Any("stoch", oscillator.Stoch[s.settings.TradingInterval]), zap.Error(err))
		return err
	}

	s.logger.Info("[OpenOrders] make orders success", zap.String("symbol", oscillator.Symbol), zap.Any("stoch", oscillator.Stoch[s.settings.TradingInterval]), zap.Any("orders", orders))

	resp, err := s.binance.OpenOrders(ctx, orders)
	if err != nil {
		s.logger.Error("[OpenOrders] failed to open orders", zap.Any("orders", orders), zap.Error(err))
		return err
	}

	s.cache.Set(oscillator.Symbol, orders)

	notifyMsg := fmt.Sprintf("Open orders success: %s #%s", helpers.ResolvePositionSide(oscillator.GetRSI(s.settings.TradingInterval)), oscillator.Symbol)
	err = s.notify.PushNotify(ctx, viper.GetInt64("notify.channels.futures_announcement"), notifyMsg)
	if err != nil {
		s.logger.Error("[OpenOrders] failed to push notification", zap.Error(err))
		return err
	}

	s.logger.Info("[OpenOrders] open order success", zap.Any("resp", resp))
	return nil
}

func (s *Orderer) checkOrderAndPositionQuantity(orders []*binancew.Order, positions []*binancew.Position) error {
	counter := 0

	for _, pos := range positions {
		if isPosititionOpened(pos) {
			counter++
		}
	}

	for _, ord := range orders {
		// ignore order type take_profit and stop_loss
		if strings.Contains(string(ord.Type), string(futures.OrderTypeStop)) ||
			strings.Contains(string(ord.Type), string(futures.OrderTypeTakeProfit)) {
			continue
		}

		if ord.Status == futures.OrderStatusTypeNew {
			counter++
		}
	}

	if counter >= int(s.settings.MaxPositionsPerTime) {
		return errors.New("trading: reached max opened")
	}

	return nil
}

func positionExisted(positions []*binancew.Position, symbol string) bool {
	for _, pos := range positions {
		if pos.Symbol == symbol && isPosititionOpened(pos) {
			return true
		}
	}
	return false
}

func orderExisted(orders []*binancew.Order, symbol string) bool {
	for _, order := range orders {
		if order.Symbol == symbol {
			return true
		}
	}
	return false
}

func isPosititionOpened(position *binancew.Position) bool {
	if position.EntryPrice != "" &&
		position.EntryPrice != "0.0" {
		return true
	}
	return false
}
