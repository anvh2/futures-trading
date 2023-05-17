package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/anvh2/futures-trading/internal/services/binance/helpers"
)

func (f *Binance) GetLeverageBracket(ctx context.Context, symbol string) ([]*LeverageBracket, error) {
	f.limiter.Wait(ctx)

	fullURL := fmt.Sprintf("%s/fapi/v1/leverageBracket", f.getURL())

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

	res, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return []*LeverageBracket{}, fmt.Errorf("error: %v", res.Status)
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	leverage := make([]*LeverageBracket, 0)
	err = json.Unmarshal(data, &leverage)
	if err != nil {
		return []*LeverageBracket{}, err
	}

	return leverage, nil
}
