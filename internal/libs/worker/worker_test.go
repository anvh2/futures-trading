package worker

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"testing"
	"time"

	logdev "github.com/anvh2/futures-trading/internal/libs/logger"
	"github.com/anvh2/futures-trading/internal/models"
)

var (
	log *logdev.Logger
)

func TestWorker(t *testing.T) {
	log = logdev.NewDev()

	w, _ := New(log, &PoolConfig{NumProcess: 64})
	w.WithProcess(test_Process)
	w.Start()

	go func() {
		ticker := time.NewTicker(time.Second)

		for range ticker.C {
			for i := 0; i < 1000; i++ {
				w.SendJob(context.Background(), func() string {
					summary := &models.CandleSummary{
						Symbol: "BTCUSDT",
					}
					return summary.String()
				}())
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second)

		for range ticker.C {
			fmt.Println(runtime.NumGoroutine())
		}
	}()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Server now listening")

	go func() {
		<-sigs
		w.Stop()

		close(done)
	}()

	fmt.Println("Ctrl-C to interrupt...")
	<-done
	fmt.Println("Exiting...")
}

func test_Process(ctx context.Context, data interface{}) error {
	return nil
}
