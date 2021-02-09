package config

type BotConfig struct {
	Telegram TelegramConfig `split_words:"true"`
	MaxBet   uint64         `split_words:"true" default:"200"`
}

type TelegramConfig struct {
	BotToken string `split_words:"true" required:"true"`
}
