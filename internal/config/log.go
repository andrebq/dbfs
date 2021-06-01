package config

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func (o *Output) MustConfigureGlobalLogger() {
	var level zerolog.Level
	switch o.Level {
	case "info", "I", "INFO":
		level = zerolog.InfoLevel
	case "error", "err", "e", "E", "ERROR", "ERR":
		level = zerolog.ErrorLevel
	case "warn", "w", "W", "WARN":
		level = zerolog.WarnLevel
	case "debug":
		level = zerolog.DebugLevel
	case "trace":
		level = zerolog.TraceLevel
	default:
		log.Panic().Str("log-level", o.Level).Msg("Log level is not valid. Please check")
	}
	zerolog.SetGlobalLevel(level)
	switch o.LogFormat {
	case "human":
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		})
	case "json", "js":
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	default:
		log.Fatal().Str("log-format", o.Format).Msg("Log format is invalid. Please check")
	}
}
