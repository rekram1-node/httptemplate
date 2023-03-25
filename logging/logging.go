package logging

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

func init() {
	zerolog.TimestampFieldName = "timestamp"
	zerolog.DurationFieldUnit = time.Millisecond
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
}

type (
	LoggerOpt func(logger *zerolog.Logger)
)

func New(opts ...LoggerOpt) *zerolog.Logger {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	for _, opt := range opts {
		opt(&logger)
	}
	return &logger
}

func WithLogLevel(level string) LoggerOpt {
	return func(logger *zerolog.Logger) {
		l, err := zerolog.ParseLevel(strings.ToLower(level))
		if err != nil {
			logger.Fatal().Err(err).Msg("unable to configure logger")
		}
		zerolog.SetGlobalLevel(l)
	}
}

func WithServiceName(name string) LoggerOpt {
	return func(logger *zerolog.Logger) {
		logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Str("service", name)
		})
	}
}

func WithVersion(version string) LoggerOpt {
	return func(logger *zerolog.Logger) {
		logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Str("version", version)
		})
	}
}
