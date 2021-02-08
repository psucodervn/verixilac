package bot

import (
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/tucnak/telebot.v2"

	"github.com/psucodervn/verixilac/internal/config"
	"github.com/psucodervn/verixilac/internal/game"
	"github.com/psucodervn/verixilac/internal/telegram"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bot",
		Short: "command description",
		Run:   run,
	}
	return cmd
}

func run(cmd *cobra.Command, args []string) {
	cfg := config.MustReadBotConfig()

	bot, err := telebot.NewBot(telebot.Settings{
		Token:  cfg.Telegram.BotToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to telegram bot")
	}

	manager := game.NewManager()

	bh := telegram.NewHandler(manager, bot)
	if err := bh.Start(); err != nil {
		log.Fatal().Err(err).Msg("failed to start bot handler")
	}
}
