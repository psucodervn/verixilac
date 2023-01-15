package config

import (
	"time"
)

type BotConfig struct {
	Telegram TelegramConfig `split_words:"true"`
	MaxBet   uint64         `split_words:"true" default:"200"`
	MinDeal  uint64         `split_words:"true" default:"1000"`
	Timeout  time.Duration  `split_words:"true" default:"1m"`
}

type TelegramConfig struct {
	BotToken string `split_words:"true" required:"true"`
}
