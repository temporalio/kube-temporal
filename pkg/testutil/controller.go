package testutil

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/temporalio/kube-temporal/pkg/controller"
	"github.com/temporalio/kube-temporal/pkg/controller/config"
	"github.com/temporalio/kube-temporal/pkg/types"
)

const (
	ControllerName = "test"
)

// NewController returns a new [controller.Controller] that uses the supplied
// fake client/reader and mocked reconciler and enables debug logging using the
// compact log format.
func NewController(
	cli client.Client,
	reader client.Reader,
	r types.Reconciler,
) *controller.Controller {
	cfg := config.New()
	cfg.SetDefaults()
	cfg.Logging.Level = "debug"
	cfg.Logging.Format = "compact"
	return controller.New(
		ControllerName, cfg, r,
		controller.WithClient(cli),
		controller.WithReader(reader),
	)
}
