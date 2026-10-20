package config

import (
	"time"

	"github.com/spf13/pflag"
)

const (
	DefaultReconcileRequeueOnSuccessDuration  = time.Second * 60
	flagReconcileRequeueOnSuccessDuration     = "reconcile-requeue-on-success-duration"
	flagReconcileRequeueOnSuccessDurationDesc = "Control how long between successful reconciliations for a resource."
)

// ReconcileConfig contains configuration settings that control the
// reconciler's behaviour.
type ReconcileConfig struct {
	// RequeueOnSuccessDuration is the amount of time the controller should
	// requeue the resource for reconciliation after a successful
	// reconciliation.
	//
	// Defaults to 60 seconds.
	RequeueOnSuccessDuration time.Duration `json:"requeueOnSuccessDuration,omitempty"`
}

// BindFlags binds the supplied flagset to the ReconcileConfig's fields.
func (c *ReconcileConfig) BindFlags(fs *pflag.FlagSet) {
	pflag.DurationVar(
		&c.RequeueOnSuccessDuration,
		flagReconcileRequeueOnSuccessDuration,
		DefaultReconcileRequeueOnSuccessDuration,
		flagReconcileRequeueOnSuccessDurationDesc,
	)
}
