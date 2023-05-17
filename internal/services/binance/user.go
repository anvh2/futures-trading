package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/anvh2/futures-trading/internal/services/binance/helpers"
)

func (f *Binance) GetListenKey(ctx context.Context) (string, error) {
	f.limiter.Wait(ctx)

	fullURL := fmt.Sprintf("%s/fapi/v1/listenKey", f.getURL())

	signed, err := helpers.Signed(http.MethodGet, fullURL, nil, f.testnet)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, signed.FullURL, signed.Body)
	if err != nil {
		return "", err
	}

	req = req.WithContext(ctx)
	req.Header = signed.Header

	resp, err := f.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("error: %v", resp.Status)
	}

	rawData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	type RespData struct {
		ListenKey string `json:"listenKey,omitempty"`
	}

	respData := &RespData{}

	if err := json.Unmarshal(rawData, respData); err != nil {
		return "", err
	}

	return respData.ListenKey, nil
}
