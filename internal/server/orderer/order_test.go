package orderer

import (
	"os"
	"testing"

	"github.com/anvh2/futures-trading/internal/logger"
	"github.com/anvh2/futures-trading/internal/services/binance"
	"github.com/joho/godotenv"
)

var (
	_loggerTest         *logger.Logger
	_binanceTestnetInst *binance.Binance
)

func TestMain(m *testing.M) {
	godotenv.Load("../../../.env")

	_loggerTest = logger.NewDev()
	_binanceTestnetInst = binance.New(_loggerTest, true)
	os.Exit(m.Run())
}
