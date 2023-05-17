package binance

import (
	"os"
	"testing"

	"github.com/anvh2/futures-trading/internal/logger"
	"github.com/joho/godotenv"
)

var (
	test_binanceInst *Binance
)

func TestMain(m *testing.M) {
	godotenv.Load("../../../.env")

	test_binanceInst = New(logger.NewDev(), true)

	os.Exit(m.Run())
}
