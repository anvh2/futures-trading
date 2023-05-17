package helpers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

type SignedData struct {
	Body    *bytes.Buffer
	Header  http.Header
	FullURL string
}

func Signed(method, fullURL string, params *url.Values, testnet bool) (*SignedData, error) {
	var (
		bodyStr  string = ""
		queryStr string = ""
	)

	if params == nil {
		params = &url.Values{}
	}

	params.Set("timestamp", fmt.Sprint(time.Now().UnixMilli()))

	var (
		apiKey    string
		secretKey string
	)

	if testnet {
		apiKey = os.Getenv("TEST_API_KEY")
		secretKey = os.Getenv("TEST_SECRET_KEY")
	} else {
		apiKey = os.Getenv("LIVE_API_KEY")
		secretKey = os.Getenv("LIVE_SECRET_KEY")
	}

	header := http.Header{}
	header.Set("X-MBX-APIKEY", apiKey)

	if params != nil {
		switch method {
		case http.MethodGet:
			queryStr = params.Encode()

		case http.MethodPost:
			bodyStr = params.Encode()
			header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}

	mac := hmac.New(sha256.New, []byte(secretKey))
	_, err := mac.Write([]byte(fmt.Sprintf("%s%s", queryStr, bodyStr)))
	if err != nil {
		return nil, err
	}

	v := url.Values{}
	v.Set("signature", fmt.Sprintf("%x", mac.Sum(nil)))

	if queryStr == "" {
		queryStr = v.Encode()
	} else {
		queryStr = fmt.Sprintf("%s&%s", queryStr, v.Encode())
	}

	if queryStr != "" {
		fullURL = fmt.Sprintf("%s?%s", fullURL, queryStr)
	}

	return &SignedData{
		Body:    bytes.NewBufferString(bodyStr),
		Header:  header,
		FullURL: fullURL,
	}, nil
}
