package telemetry

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// NewLogger cria um logger estruturado com nível definido em configuração.
func NewLogger(level string) zerolog.Logger {
	zl := zerolog.New(os.Stdout).With().Timestamp().Logger()
	lvl, err := zerolog.ParseLevel(strings.ToLower(level))
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	zerolog.TimeFieldFormat = time.RFC3339Nano
	return zl.Level(lvl)
}
