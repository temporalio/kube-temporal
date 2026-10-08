package base

import (
	"log/slog"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/temporalio/kube-temporal/pkg/controller/config"
)

// Controller is a base struct with common methods and a standardized reconcile
// loop flow. It is meant to be embedded in controller structs to reduce the
// amount of copy/paste necessary in constructing new operator repositories.
type Controller struct {
	Client client.Client
	Reader client.Reader

	// cfg is the configuration settings used when creating the Reconciler.
	cfg *config.Config
	// logger is the root logger for the Reconciler
	logger *slog.Logger
}

// Logger returns the root logger for the Controller.
func (c Controller) Logger() *slog.Logger {
	return c.logger
}

// New returns a new base Controller.
func New(
	cfg *config.Config,
	logger *slog.Logger,
) Controller {
	return Controller{
		cfg:    cfg,
		logger: logger,
	}
}
