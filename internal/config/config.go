package config

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Trading  TradingConfig  `mapstructure:"trading"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Market   MarketConfig   `mapstructure:"market"`
	Chart    ChartConfig    `mapstructure:"chart"`
	Binance  BinanceConfig  `mapstructure:"binance"`
	Notify   NotifyConfig   `mapstructure:"notify"`
	Telegram TelegramConfig `mapstructure:"telegram"`
	Topics   TopicsConfig   `mapstructure:"topics"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type TradingConfig struct {
	LogPath string `mapstructure:"log_path"`
}

type RedisConfig struct {
	Addr string `mapstructure:"addr"`
}

type MarketConfig struct {
	Intervals []string `mapstructure:"intervals"`
}

type ChartConfig struct {
	Candles CandlesConfig `mapstructure:"candles"`
}

type CandlesConfig struct {
	Limit int `mapstructure:"limit"`
}

type BinanceConfig struct {
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	Secrets   SecretsConfig   `mapstructure:"secrets"`
}

type RateLimitConfig struct {
	Requests int    `mapstructure:"requests"`
	Duration string `mapstructure:"duration"`
}

type SecretsConfig struct {
	APIKey string `mapstructure:"api_key"`
}

type NotifyConfig struct {
	Channels map[string]int64 `mapstructure:"channels"`
	Config   NotifyDetail     `mapstructure:"config"`
}

type NotifyDetail struct {
	Expiration string `mapstructure:"expiration"`
}

type TelegramConfig struct {
	Token string `mapstructure:"token"`
}

type TopicsConfig struct {
	SymbolsRetryTopic       string `mapstructure:"symbols_retry_topic"`
	SymbolsSignalTopic      string `mapstructure:"symbols_signal_topic"`
	SymbolsTradeIntentTopic string `mapstructure:"symbols_trade_intent_topic"`
}
