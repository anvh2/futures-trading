package binance

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/anvh2/futures-trading/internal/models"
	"github.com/anvh2/futures-trading/internal/services/binance/helpers"
)

func (f *Binance) GetPositionRisk(ctx context.Context, symbol string) ([]*Position, error) {
	f.limiter.Wait(ctx)

	fullURL := fmt.Sprintf("%s/fapi/v2/positionRisk", f.getURL())

	params := &url.Values{
		"symbol": []string{symbol},
	}

	signed, err := helpers.Signed(http.MethodGet, fullURL, params, f.testnet)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, signed.FullURL, signed.Body)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Header = signed.Header

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error: %v", resp.Status)
	}

	rawData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	isObject := len(rawData) > 0 && rawData[0] == '{' && rawData[len(rawData)-1] == '}'
	if isObject {
		position := &Position{}
		json.Unmarshal(rawData, position)
		return []*Position{position}, nil
	}

	res := make([]*Position, 0)
	err = json.Unmarshal(rawData, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (f *Binance) GetOpenOrders(ctx context.Context, symbol string) ([]*Order, error) {
	f.limiter.Wait(ctx)

	fullURL := fmt.Sprintf("%s/fapi/v1/openOrders", f.getURL())

	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	query := req.URL.Query()
	query.Add("symbol", symbol)
	query.Add("timestamp", fmt.Sprint(time.Now().UnixMilli()))

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rawData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	isObject := len(rawData) > 0 && rawData[0] == '{' && rawData[len(rawData)-1] == '}'
	if isObject {
		order := &Order{}
		json.Unmarshal(rawData, order)
		return []*Order{order}, nil
	}

	res := make([]*Order, 0)
	err = json.Unmarshal(rawData, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (f *Binance) OpenOrders(ctx context.Context, orders []*models.Order) ([]*CreateOrderResp, error) {
	f.limiter.Wait(ctx)

	fullURL := fmt.Sprintf("%s/fapi/v1/batchOrders", f.getURL())

	ordersMap := make([]map[string]interface{}, len(orders))

	for idx, order := range orders {
		ordersMap[idx] = map[string]interface{}{
			"symbol":           order.Symbol,
			"side":             order.Side,
			"type":             order.OrderType,
			"quantity":         order.Quantity,
			"newOrderRespType": order.NewOrderRespType,
		}

		if order.PositionSide != "" {
			ordersMap[idx]["positionSide"] = order.PositionSide
		}
		if order.TimeInForce != "" {
			ordersMap[idx]["timeInForce"] = order.TimeInForce
		}
		if order.ReduceOnly {
			ordersMap[idx]["reduceOnly"] = fmt.Sprint(order.ReduceOnly)
		}
		if order.Price != "" {
			ordersMap[idx]["price"] = order.Price
		}
		if order.NewClientOrderId != "" {
			ordersMap[idx]["newClientOrderId"] = order.NewClientOrderId
		}
		if order.StopPrice != "" {
			ordersMap[idx]["stopPrice"] = order.StopPrice
		}
		if order.WorkingType != "" {
			ordersMap[idx]["workingType"] = order.WorkingType
		}
		if order.PriceProtect {
			ordersMap[idx]["priceProtect"] = order.PriceProtect
		}
		if order.ActivationPrice != "" {
			ordersMap[idx]["activationPrice"] = order.ActivationPrice
		}
		if order.CallbackRate != "" {
			ordersMap[idx]["callbackRate"] = order.CallbackRate
		}
		if order.ClosePosition {
			ordersMap[idx]["closePosition"] = order.ClosePosition
		}
	}

	b, err := json.Marshal(ordersMap)
	if err != nil {
		return nil, err
	}

	params := map[string]interface{}{
		"batchOrders": string(b),
		"timestamp":   time.Now().UnixMilli(),
	}

	form := &url.Values{
		"batchOrders": []string{string(b)},
		"timestamp":   []string{fmt.Sprint(time.Now().UnixMilli())},
	}

	for key, val := range params {
		form.Set(key, fmt.Sprint(val))
	}

	bodyStr := form.Encode()
	body := bytes.NewBufferString(bodyStr)

	header := http.Header{}
	header.Set("X-MBX-APIKEY", os.Getenv("TEST_API_KEY"))
	header.Set("Content-Type", "application/x-www-form-urlencoded")

	mac := hmac.New(sha256.New, []byte(os.Getenv("TEST_SECRET_KEY")))
	_, err = mac.Write([]byte(bodyStr))
	if err != nil {
		return nil, err
	}

	v := url.Values{}
	v.Set("signature", fmt.Sprintf("%x", (mac.Sum(nil))))

	queryString := v.Encode()
	if queryString != "" {
		fullURL = fmt.Sprintf("%s?%s", fullURL, queryString)
	}

	req, err := http.NewRequest(http.MethodPost, fullURL, body)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Header = header

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rawData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	isObject := len(rawData) > 0 && rawData[0] == '{' && rawData[len(rawData)-1] == '}'
	if isObject {
		order := &CreateOrderResp{}
		json.Unmarshal(rawData, order)
		return []*CreateOrderResp{order}, nil
	}

	jsonData := make([]*json.RawMessage, 0)
	err = json.Unmarshal(rawData, &jsonData)
	if err != nil {
		return nil, err
	}

	createResp := make([]*CreateOrderResp, len(jsonData))

	for idx, data := range jsonData {
		order := &CreateOrderResp{}
		json.Unmarshal(*data, order)
		createResp[idx] = order
	}

	return createResp, nil
}
