package bot

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/telebot.v3"

	"github.com/psucodervn/verixilac/internal/config"
	"github.com/psucodervn/verixilac/internal/game"
	"github.com/psucodervn/verixilac/internal/storage"
	"github.com/psucodervn/verixilac/internal/telegram"
	"github.com/psucodervn/verixilac/pkg/logger"
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
	logger.Init(true, true)

	cfg := config.MustReadBotConfig()

	go func() {
		log.Err(http.ListenAndServe("localhost:6060", nil)).Send()
	}()

	bot, err := telebot.NewBot(telebot.Settings{
		Token:  cfg.Telegram.BotToken,
		Poller: &telebot.LongPoller{Timeout: 3 * time.Second},
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to telegram bot")
	}

	store := storage.NewBadgerHoldStorage("data")
	manager := game.NewManager(store, cfg.MaxBet, cfg.MinDeal, cfg.Timeout)

	// listen to interrupt signal i.e Ctrl+C
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		log.Info().Msg("shutting down bot")
		// if _, err := bot.Close(); err != nil {
		// 	log.Err(err).Msg("failed to stop bot")
		// }
		if err := store.Close(); err != nil {
			log.Err(err).Msg("failed to close storage")
		}
		os.Exit(0)
	}()

	bh := telegram.NewHandler(manager, bot, store)
	log.Info().Msg("bot started")
	if err := bh.Start(); err != nil {
		log.Fatal().Err(err).Msg("failed to start bot handler")
	}
}
