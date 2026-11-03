package temporal

import (
	"log/slog"

	tlog "go.temporal.io/sdk/log"
)

func NewNoopLogger() tlog.Logger {
	return tlog.NewStructuredLogger(slog.New(slog.DiscardHandler))
}
