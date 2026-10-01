package config

import (
	"github.com/spf13/pflag"
)

const (
	DefaultHealthzBindAddress  = ":8081"
	flagHealthzBindAddress     = "healthz-bind-address"
	flagHealthzBindAddressDesc = "The address the healthz probe endpoint binds to."
)

// HealthzConfig contains healthz probe configuration options for the Temporal
// Cluster Operator
type HealthzConfig struct {
	// BindAddress is the address to bind the healthz probe to.
	BindAddress string `json:"bindAddress"`
}

// BindFlags bings the supplied flagset to the HealthzConfig's fields.
func (c *HealthzConfig) BindFlags(fs *pflag.FlagSet) {
	pflag.StringVar(
		&c.BindAddress,
		flagHealthzBindAddress,
		DefaultHealthzBindAddress,
		flagHealthzBindAddressDesc,
	)
}
