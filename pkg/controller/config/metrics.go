package config

import (
	"github.com/spf13/pflag"
	metricsfilters "sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

const (
	DefaultMetricsDisabled  = false
	flagMetricsDisabled     = "metrics-disabled"
	flagMetricsDisabledDesc = "Use to disable the metrics endpoint entirely."

	DefaultMetricsBindAddress  = ":8080"
	flagMetricsBindAddress     = "metrics-bind-address"
	flagMetricsBindAddressDesc = "Address the metrics endpoint will bind to."

	DefaultMetricsTLSDisabled  = false
	flagMetricsTLSDisabled     = "metrics-tls-disabled"
	flagMetricsTLSDisabledDesc = "Use to force the metrics endpoint to be served insecurely over HTTP."

	DefaultMetricsTLSDir  = ""
	flagMetricsTLSDir     = "metrics-tls-dir"
	flagMetricsTLSDirDesc = "Directory to find the TLS server certificate file to use for the metrics endpoint."

	DefaultMetricsTLSCertName  = "tls.crt"
	flagMetricsTLSCertName     = "metrics-tls-cert-name"
	flagMetricsTLSCertNameDesc = "Name of the TLS server certificate file to use for the metrics endpoint."

	DefaultMetricsTLSKeyName  = "tls.key"
	flagMetricsTLSKeyName     = "metrics-tls-key-name"
	flagMetricsTLSKeyNameDesc = "Name of the TLS server key file to use for the metrics endpoint."
)

// MetricsConfig contains leader election configuration options for the
// Temporal Cluster Operator
type MetricsConfig struct {
	// Disabled indicates the metrics endpoint will not be set up.
	Disabled bool `json:"disabled"`
	// BindAddress the metrics endpoint will bind to.
	BindAddress string `json:"bindAddress"`
	// TLS contains server TLS configuration for the metrics endpoint.
	TLS TLSServerConfig `json:"tls"`
}

// BindFlags bings the supplied flagset to the MetricsConfig's fields.
func (c *MetricsConfig) BindFlags(fs *pflag.FlagSet) {
	pflag.BoolVar(
		&c.Disabled,
		flagMetricsDisabled,
		DefaultMetricsDisabled,
		flagMetricsDisabledDesc,
	)
	pflag.StringVar(
		&c.BindAddress,
		flagMetricsBindAddress,
		DefaultMetricsBindAddress,
		flagMetricsBindAddressDesc,
	)
	pflag.BoolVar(
		&c.TLS.Disabled,
		flagMetricsTLSDisabled,
		DefaultMetricsTLSDisabled,
		flagMetricsTLSDisabledDesc,
	)
	pflag.StringVar(
		&c.TLS.Dir,
		flagMetricsTLSDir,
		DefaultMetricsTLSDir,
		flagMetricsTLSDirDesc,
	)
	pflag.StringVar(
		&c.TLS.CertName,
		flagMetricsTLSCertName,
		DefaultMetricsTLSCertName,
		flagMetricsTLSCertNameDesc,
	)
	pflag.StringVar(
		&c.TLS.KeyName,
		flagMetricsTLSKeyName,
		DefaultMetricsTLSKeyName,
		flagMetricsTLSKeyNameDesc,
	)
}

// ToMetricsServerOptions returns the
// [controller-runtime/pkg/metrics/server.Options] representation of the
// metrics configuration.
func (c *MetricsConfig) ToMetricsServerOptions() metricsserver.Options {
	opts := metricsserver.Options{}
	if c.Disabled {
		return opts
	}
	opts.BindAddress = c.BindAddress
	opts.SecureServing = !c.TLS.Disabled
	if !c.TLS.Disabled {
		opts.CertDir = c.TLS.Dir
		opts.CertName = c.TLS.CertName
		opts.KeyName = c.TLS.KeyName
		opts.TLSOpts = c.TLS.ToOptions()
		// FilterProvider is used to protect the metrics endpoint with
		// authn/authz.  These configurations ensure that only authorized users
		// and service accounts can access the metrics endpoint. The RBAC are
		// configured in 'config/rbac/kustomization.yaml'.
		//
		// More info: https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.23.3/pkg/metrics/filters#WithAuthenticationAndAuthorization
		opts.FilterProvider = metricsfilters.WithAuthenticationAndAuthorization
	}
	return opts
}
