package config

import (
	"crypto/tls"
)

// TLSServerConfig contains TLS configuration options for a server. It is meant
// to be included in other Config structs like MetricsConfig or WebhookConfig.
type TLSServerConfig struct {
	// Disabled causes the endpoint to be served via HTTP instead of HTTPS.
	Disabled bool `json:"disabled,omitempty"`
	// HTTP2Enabled allows http/2 connections to the server endpoint.
	//
	// This should only be used with extreme caution due to vulnerabilities associated with http/2.
	//
	// Keeping http/2 disabled (the default) will prevent the server endpoint
	// from being vulnerable to the HTTP/2 Stream Cancellation and Rapid Reset
	// CVEs. For more information see:
	//
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	HTTP2Enabled bool `json:"http2Enabled,omitempty"`
	// Dir contains the directory to find the TLS certificate and key file.
	Dir string `json:"dir,omitempty"`
	// CertName contains the name of the server TLS certificate file.
	CertName string `json:"certName,omitempty"`
	// KeyName contains the name of the server TLS key file.
	KeyName string `json:"keyName,omitempty"`
}

// ToOptions returns a representation of the TLSServerConfig as a slice of
// func(*tls.Config).
func (c *TLSServerConfig) ToOptions() []func(*tls.Config) {
	var tlsOpts []func(*tls.Config)
	disableHTTP2 := func(tc *tls.Config) {
		tc.NextProtos = []string{"http/1.1"}
	}

	if !c.HTTP2Enabled {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}
	return tlsOpts
}
