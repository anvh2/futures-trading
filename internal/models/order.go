package models

import (
	"encoding/json"

	"github.com/adshao/go-binance/v2/futures"
)

type Order struct {
	OrderId          string                   `json:"order_id,omitempty"`
	Symbol           string                   `json:"symbol,omitempty"`
	Side             futures.SideType         `json:"side,omitempty"`
	PositionSide     futures.PositionSideType `json:"position_side,omitempty"`
	OrderType        futures.OrderType        `json:"order_type,omitempty"`
	TimeInForce      futures.TimeInForceType  `json:"time_in_force,omitempty"`
	Quantity         string                   `json:"quantity,omitempty"`
	ReduceOnly       bool                     `json:"reduce_only,omitempty"`
	Price            string                   `json:"price,omitempty"`
	NewClientOrderId string                   `json:"new_client_order_id,omitempty"`
	StopPrice        string                   `json:"stop_price,omitempty"`
	WorkingType      futures.WorkingType      `json:"working_type,omitempty"`
	ActivationPrice  string                   `json:"activation_price,omitempty"`
	CallbackRate     string                   `json:"callback_rate,omitempty"`
	PriceProtect     bool                     `json:"price_protect,omitempty"`
	NewOrderRespType futures.NewOrderRespType `json:"new_order_resp_type,omitempty"`
	ClosePosition    bool                     `json:"close_position,omitempty"`
}

func (o *Order) String() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func (o *Order) Parse(val string) error {
	return json.Unmarshal([]byte(val), o)
}
