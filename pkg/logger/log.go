package logger

import (
	"io"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	loggerWithoutCaller = zerolog.New(os.Stderr)
)

func Init(debug bool, pretty bool, additionalWriters ...io.Writer) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if pretty {
		additionalWriters = append(additionalWriters, zerolog.ConsoleWriter{Out: os.Stderr})
	} else {
		additionalWriters = append(additionalWriters, os.Stderr)
	}

	loggerWithoutCaller = log.Output(zerolog.MultiLevelWriter(additionalWriters...))
	log.Logger = loggerWithoutCaller.With().Caller().Logger()
}

func InitWithConfig(cfg Config) {
	Init(cfg.Debug, cfg.Pretty)
}

func InitFromEnv() {
	var cfg Config
	envconfig.MustProcess("LOG", &cfg)
	InitWithConfig(cfg)
}

// WithoutCaller return a clone logger without caller field
func WithoutCaller() zerolog.Logger {
	return loggerWithoutCaller
}
