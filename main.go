package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/psucodervn/verixilac/cmd/bot"
	"github.com/psucodervn/verixilac/pkg/logger"
)

var (
	rootCmd = &cobra.Command{
		Use:              "main",
		PersistentPreRun: preRun,
	}
	envFiles []string
)

func preRun(cmd *cobra.Command, args []string) {
	if len(envFiles) == 0 {
		if _, err := os.Stat(".env"); err == nil {
			envFiles = append(envFiles, ".env")
		}
	}
	if len(envFiles) > 0 {
		if err := godotenv.Overload(envFiles...); err != nil {
			log.Err(err).Msg("read env files failed")
		}
	}
	logger.InitFromEnv()
}

func init() {
	rootCmd.AddCommand(
		bot.Command(),
	)
	rootCmd.PersistentFlags().StringSliceVarP(&envFiles, "envfile", "e", nil, "env files")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("execute failed")
	}
}
