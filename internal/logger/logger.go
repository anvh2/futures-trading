package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

func New(file string) (*Logger, error) {
	config := zap.NewProductionConfig()

	config.OutputPaths = []string{file}
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.TimeKey = "ts"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.MessageKey = "message"
	config.DisableStacktrace = true

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{
		Logger: logger,
	}, nil
}

func NewDev() *Logger {
	logger, _ := zap.NewDevelopment()
	return &Logger{
		Logger: logger,
	}
}
