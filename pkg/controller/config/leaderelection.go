package config

import (
	"github.com/spf13/pflag"
)

const (
	flagLeaderElectionEnabled     = "leader-election-enabled"
	flagLeaderElectionEnabledDesc = "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager."
)

// LeaderElectionConfig contains leader election configuration options for the
// Temporal Cluster Operator
type LeaderElectionConfig struct {
	// Enabled turns on leader election for the controller manager.
	Enabled bool `json:"enabled"`
}

// BindFlags bings the supplied flagset to the LeaderElectionConfig's fields.
func (c *LeaderElectionConfig) BindFlags(fs *pflag.FlagSet) {
	pflag.BoolVar(
		&c.Enabled,
		flagLeaderElectionEnabled, false, flagLeaderElectionEnabledDesc,
	)
}
