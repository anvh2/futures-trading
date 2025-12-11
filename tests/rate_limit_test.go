package test

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestRateLimit(t *testing.T) {
	limiter := rate.NewLimiter(rate.Every(time.Minute), 1000)

	client := &http.Client{}
	wg := &sync.WaitGroup{}
	start := time.Now()

	for i := 0; i < 5000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			limiter.Wait(context.Background())

			url := "https://www.binance.com/fapi/v1/continuousKlines?limit=1000&pair=BTCUSDT&contractType=PERPETUAL&interval=1h"

			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				return
			}

			res, err := client.Do(req)
			if err != nil {
				return
			}

			fmt.Println(res.Status)
		}()
	}

	wg.Wait()
	fmt.Println(time.Since(start).Seconds())
}

func TestRateLimiter(t *testing.T) {
	limiter := rate.NewLimiter(rate.Every(time.Minute), 1200)

	wg := &sync.WaitGroup{}

	for i := 0; i < 5000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			limiter.Wait(context.Background())

		}()
	}

	wg.Wait()
}
