package config

import (
	"log/slog"

	"github.com/go-logr/logr"
	ctrlrt "sigs.k8s.io/controller-runtime"
)

// SetupLogging initializes the controller's logging based on the Config's
// Logging configuration.
func (c *Config) SetupLogging() {
	logger := c.Logging.Logger()
	slog.SetDefault(logger)
	ctrlrt.SetLogger(logr.FromSlogHandler(logger.Handler()))
}
