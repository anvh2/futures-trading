package binance

import "github.com/adshao/go-binance/v2/futures"

type Error struct {
	Code int    `json:"code,omitempty"`
	Msg  string `json:"msg,omitempty"`
}

type CreateOrderResp struct {
	*Error
	OrderId       int    `json:"orderId,omitempty"`
	Symbol        string `json:"symbol,omitempty"`
	Status        string `json:"status,omitempty"`
	ClientOrderId string `json:"clientOrderId,omitempty"`
	Price         string `json:"price,omitempty"`
	AvgPrice      string `json:"avgPrice,omitempty"`
	OrigQty       string `json:"origQty,omitempty"`
	ExecutedQty   string `json:"executedQty,omitempty"`
	CumQty        string `json:"cumQty,omitempty"`
	CumQuote      string `json:"cumQuote,omitempty"`
	TimeInForce   string `json:"timeInForce,omitempty"`
	Type          string `json:"type,omitempty"`
	ReduceOnly    bool   `json:"reduceOnly,omitempty"`
	ClosePosition bool   `json:"closePosition,omitempty"`
	Side          string `json:"side,omitempty"`
	PositionSide  string `json:"positionSide,omitempty"`
	StopPrice     string `json:"stopPrice,omitempty"`
	WorkingType   string `json:"workingType,omitempty"`
	PriceProtect  bool   `json:"priceProtect,omitempty"`
	OrigType      string `json:"origType,omitempty"`
	UpdateTime    int64  `json:"updateTime,omitempty"`
}

type Position struct {
	*Error
	EntryPrice       string `json:"entryPrice"`
	MarginType       string `json:"marginType"`
	IsAutoAddMargin  string `json:"isAutoAddMargin"`
	IsolatedMargin   string `json:"isolatedMargin"`
	Leverage         string `json:"leverage"`
	LiquidationPrice string `json:"liquidationPrice"`
	MarkPrice        string `json:"markPrice"`
	MaxNotionalValue string `json:"maxNotionalValue"`
	PositionAmt      string `json:"positionAmt"`
	Symbol           string `json:"symbol"`
	UnRealizedProfit string `json:"unRealizedProfit"`
	PositionSide     string `json:"positionSide"`
	Notional         string `json:"notional"`
	IsolatedWallet   string `json:"isolatedWallet"`
}

type Order struct {
	Error            *Error
	Symbol           string                   `json:"symbol"`
	OrderID          int64                    `json:"orderId"`
	ClientOrderID    string                   `json:"clientOrderId"`
	Price            string                   `json:"price"`
	ReduceOnly       bool                     `json:"reduceOnly"`
	OrigQuantity     string                   `json:"origQty"`
	ExecutedQuantity string                   `json:"executedQty"`
	CumQuantity      string                   `json:"cumQty"`
	CumQuote         string                   `json:"cumQuote"`
	Status           futures.OrderStatusType  `json:"status"`
	TimeInForce      futures.TimeInForceType  `json:"timeInForce"`
	Type             futures.OrderType        `json:"type"`
	Side             futures.SideType         `json:"side"`
	StopPrice        string                   `json:"stopPrice"`
	Time             int64                    `json:"time"`
	UpdateTime       int64                    `json:"updateTime"`
	WorkingType      futures.WorkingType      `json:"workingType"`
	ActivatePrice    string                   `json:"activatePrice"`
	PriceRate        string                   `json:"priceRate"`
	AvgPrice         string                   `json:"avgPrice"`
	OrigType         string                   `json:"origType"`
	PositionSide     futures.PositionSideType `json:"positionSide"`
	PriceProtect     bool                     `json:"priceProtect"`
	ClosePosition    bool                     `json:"closePosition"`
}

// LeverageBracket define the leverage bracket
type LeverageBracket struct {
	Symbol   string    `json:"symbol"`
	Brackets []Bracket `json:"brackets"`
}

// Bracket define the bracket
type Bracket struct {
	Bracket          int     `json:"bracket"`
	InitialLeverage  int     `json:"initialLeverage"`
	NotionalCap      float64 `json:"notionalCap"`
	NotionalFloor    float64 `json:"notionalFloor"`
	MaintMarginRatio float64 `json:"maintMarginRatio"`
	Cum              float64 `json:"cum"`
}
