package controller

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	ctrlrt "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	ctrlrtrecon "sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/temporalio/kube-temporal/pkg/controller/config"
	"github.com/temporalio/kube-temporal/pkg/types"
)

var once sync.Once

// Controller implements [ctrlrt.Reconciler] and reduces the amount of
// copy/paste necessary in constructing new operator repositories.
type Controller struct {
	// name is the name of the Controller.
	name string
	// finalizerName is the name of the finalizer this Controller attaches to
	// resources it manages.
	finalizerName string
	// client is the controller-runtime client.Client associated with the
	// Controller.
	client client.Client
	// reader if the direct, non-caching API reader associated with Controller.
	reader client.Reader
	// cfg contains the configuration settings used when creating the
	// Controller.
	cfg config.Config
	// logger is the root logger for the Reconciler
	logger *slog.Logger
	// r is the Controller's bound [types.ResourceReconciler] which
	// understands how to reconcile the kinds of [types.Resource] that this
	// Controller manages.
	r types.Reconciler
}

// Name returns the name of the Controller.
func (c Controller) Name() string {
	return c.name
}

// FinalizerName returns the name of the finalizer to attach to the associated
// [types.Resource]
func (c *Controller) FinalizerName(res types.Resource) string {
	if c.finalizerName != "" {
		return c.finalizerName
	}
	once.Do(func() {
		gk := res.GroupVersionKind().GroupKind()
		fname := fmt.Sprintf("%s/finalizer", strings.ToLower(gk.String()))
		c.finalizerName = fname
	})
	return c.finalizerName
}

// NewOption allows configuring options on the Controller returned by New.
type NewOption func(*Controller)

// WithLogger sets the Controller's root logger.
func WithLogger(logger *slog.Logger) NewOption {
	return func(c *Controller) {
		c.logger = logger
	}
}

// WithClient allows overriding the Controller's controller-runtime Client.
func WithClient(cli client.Client) NewOption {
	return func(c *Controller) {
		c.client = cli
	}
}

// WithReader allows overriding the Controller's controller-runtime Reader.
func WithReader(reader client.Reader) NewOption {
	return func(c *Controller) {
		c.reader = reader
	}
}

// New returns a new Controller and binds it to the supplied
// [ctrlrt.Manager] and [types.ResourceReconciler].
func New(
	name string,
	cfg config.Config,
	r types.Reconciler,
	opts ...NewOption,
) *Controller {
	r.SetConfig(cfg)
	c := Controller{
		name: name,
		cfg:  cfg,
		r:    r,
	}
	for _, opt := range opts {
		opt(&c)
	}
	if c.logger == nil {
		c.logger = slog.New(slog.DiscardHandler)
	}
	return &c
}

// BindManager binds the Controller and its associated [types.Reconciler] to
// the supplied [ctrlrt.Manager].
func (c *Controller) BindManager(mgr ctrlrt.Manager) error {
	c.client = mgr.GetClient()
	c.reader = mgr.GetAPIReader()
	c.r.SetKubeClient(c.client)
	c.r.SetKubeReader(c.reader)
	builder := ctrlrt.NewControllerManagedBy(mgr).
		Named(c.name).
		WithEventFilter(
			predicate.GenerationChangedPredicate{},
		)
	builder = c.r.AdaptBuilder(builder)
	return builder.Complete(c)
}

// compile-time verification that Controller implements the controller-runtime
// Reconciler contract.
var _ ctrlrtrecon.Reconciler = (*Controller)(nil)
