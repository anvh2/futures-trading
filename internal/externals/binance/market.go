package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/bitly/go-simplejson"
)

func (f *Binance) GetExchangeInfo(ctx context.Context) (*futures.ExchangeInfo, error) {
	f.limiter.Wait(ctx)

	fullURL := fmt.Sprintf("%s/fapi/v1/exchangeInfo", f.getURL())

	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rawData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	res := &futures.ExchangeInfo{}
	err = json.Unmarshal(rawData, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (f *Binance) GetCurrentPrice(ctx context.Context, symbol string) (*futures.SymbolPrice, error) {
	f.limiter.Wait(ctx)

	url := fmt.Sprintf("%s/fapi/v1/ticker/price?symbol=%s", f.getURL(), symbol)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	res, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	price := &futures.SymbolPrice{}

	if err := json.Unmarshal(data, price); err != nil {
		return price, err
	}

	return price, nil
}

func (f *Binance) GetCandlesticks(ctx context.Context, symbol, interval string, limit int, startTime, endTime int64) ([]*binance.Kline, error) {
	f.limiter.Wait(ctx)

	url := fmt.Sprintf("https://www.binance.com/fapi/v1/continuousKlines?limit=%d&pair=%s&contractType=PERPETUAL&interval=%s", limit, symbol, interval)
	if startTime != 0 {
		url += fmt.Sprintf("&startTime=%d", startTime)
	}
	if endTime != 0 {
		url += fmt.Sprintf("&endTime=%d", endTime)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	res, err := f.client.Do(req)
	if err != nil {
		return []*binance.Kline{}, err
	}

	if res.StatusCode != 200 {
		return []*binance.Kline{}, fmt.Errorf("binance: failed, make request with code %d", res.StatusCode)
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []*binance.Kline{}, err
	}

	defer res.Body.Close()

	json, err := simplejson.NewJson(data)
	if err != nil {
		return []*binance.Kline{}, err
	}

	num := len(json.MustArray())
	resp := make([]*binance.Kline, num)
	for i := 0; i < num; i++ {
		item := json.GetIndex(i)
		if len(item.MustArray()) < 11 {
			return []*binance.Kline{}, fmt.Errorf("invalid kline response")
		}
		resp[i] = &binance.Kline{
			OpenTime:                 item.GetIndex(0).MustInt64(),
			Open:                     item.GetIndex(1).MustString(),
			High:                     item.GetIndex(2).MustString(),
			Low:                      item.GetIndex(3).MustString(),
			Close:                    item.GetIndex(4).MustString(),
			Volume:                   item.GetIndex(5).MustString(),
			CloseTime:                item.GetIndex(6).MustInt64(),
			QuoteAssetVolume:         item.GetIndex(7).MustString(),
			TradeNum:                 item.GetIndex(8).MustInt64(),
			TakerBuyBaseAssetVolume:  item.GetIndex(9).MustString(),
			TakerBuyQuoteAssetVolume: item.GetIndex(10).MustString(),
		}
	}

	return resp, nil
}
