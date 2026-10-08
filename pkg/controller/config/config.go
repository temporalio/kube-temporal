package config

import (
	"github.com/spf13/pflag"

	"github.com/temporalio/kube-temporal/pkg/log"
)

// WithOption modifies a Config returned from New.
type WithOption func(*Config)

// New returns a new Config.
func New(opts ...WithOption) *Config {
	c := &Config{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithFlags binds the supplied flags to the Config's fields.
func WithFlags(fs *pflag.FlagSet) WithOption {
	return func(c *Config) {
		c.BindFlags(fs)
	}
}

// Config contains configuration options for a standardized controller.
type Config struct {
	// LeaderElection contains options for configuring leader election
	// behavior.
	LeaderElection LeaderElectionConfig `json:"leaderElection"`
	// Logging contains options for configuring logging.
	Logging log.Config `json:"logging"`
	// Healthz contains options for configuring a healthz endpoint.
	Healthz HealthzConfig `json:"healthz"`
	// Metrics contains options for configuring metrics collection and
	// publishing.
	Metrics MetricsConfig `json:"metrics"`
}

// SetDefaults sets any default configuration values from environment
// variables.
func (c *Config) SetDefaults() {
	c.Logging.SetDefaults()
}

// BindFlags bings the supplied flagset to the Config's fields.
func (c *Config) BindFlags(fs *pflag.FlagSet) {
	c.LeaderElection.BindFlags(fs)
	c.Logging.BindFlags(fs)
	c.Healthz.BindFlags(fs)
	c.Metrics.BindFlags(fs)
}
