package telegram

import (
	"context"
	"time"

	"github.com/anvh2/futures-trading/internal/libs/logger"
	"go.uber.org/zap"
	tb "gopkg.in/telebot.v3"
)

//go:generate moq -pkg telemock -out ./mocks/telegram_mock.go . Notify
type Notify interface {
	PushNotify(ctx context.Context, chatId int64, message string) error
	Stop()
}

type TelegramBot struct {
	logger *logger.Logger
	bot    *tb.Bot
}

func NewTelegramBot(logger *logger.Logger, token string) (*TelegramBot, error) {
	setting := tb.Settings{
		Token: token,
		Poller: &tb.LongPoller{
			Timeout: 10 * time.Second,
		},
	}

	bot, err := tb.NewBot(setting)
	if err != nil {
		logger.Error("failed to new telegram bot", zap.Error(err))
		return nil, err
	}

	go bot.Start()

	return &TelegramBot{
		bot:    bot,
		logger: logger,
	}, nil
}

func (t *TelegramBot) Handle(command string, handler func(ctx context.Context, args []string) (interface{}, error)) {
	t.bot.Handle(command, func(ctx tb.Context) error {
		resp, err := handler(context.Background(), ctx.Args())
		if err != nil {
			return err
		}

		return ctx.Send(resp)
	})
}

func (t *TelegramBot) PushNotify(ctx context.Context, chatId int64, message string) error {
	resp, err := t.bot.Send(&tb.User{ID: chatId}, message)
	if err != nil {
		t.logger.Error("[TelegramBot] failed to send message", zap.Any("message", message), zap.Error(err))
		return err
	}

	t.logger.Info("[TelegramBot] push message success", zap.Any("message", message), zap.Any("messageId", resp.ID))
	return nil
}

func (t *TelegramBot) Stop() {
	t.bot.Stop()
}
