package telegram

import (
	"context"
	"fmt"
	"testing"

	"github.com/anvh2/futures-trading/internal/logger"
)

func TestSend(t *testing.T) {
	bot, err := NewTelegramBot(logger.NewDev(), "5392735903:AAHgMUpDqcKyiSbYLgVfZkKaOLjYPsLkgBs")
	if err != nil {
		fmt.Println(err)
		return
	}

	bot.PushNotify(context.Background(), -1001795149770, "hello world")
}
