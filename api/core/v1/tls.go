package v1

import corev1 "k8s.io/api/core/v1"

// ServerTLSConfig defines TLS settings for a Connection.
type ServerTLSConfig struct {
	// ServerName overrides the server name used to verify the Temporal server
	// certificate when TLS is enabled.
	// +optional
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9.-]+$`
	ServerName string `json:"serverName,omitempty"`

	// CACert defines where to find the certificate authority (CA) certificate
	// to use for a Connection. This certificate is appended to/ the system
	// trust store for the Connection.
	//
	// Note that this is only applicable when the Connection uses API Key or no
	// authentication credentials. When using Mutual TLS authentication, the CA
	// certificate is configured using the MutualTLSConfig struct.
	// +optional
	CACert *CACertConfig `json:"caCert,omitempty"`
}

// CACertConfig defines where to find the certificate authority (CA)
// certificate to use for a Connection.
type CACertConfig struct {
	// SecretRef references a Kubernetes Secret that is either `type: opaque`
	// or `type: kubernetes.io/tls`.
	//
	// When the Secret type is `opaque`, the `SecretKeySelector.type` field
	// should contain the key in the Secret data where the CA certificate file
	// contents can be found.
	//
	// When the Secret type is `kubernetes.io/tls`, the
	// `SecretKeySelector.type` field is ignored. The CA certificate file
	// contents should be at the  standard `ca.crt` key location in the Secret
	// data.
	//
	// More information about creating a TLS secret:
	// https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets
	// +optional
	SecretRef *corev1.SecretKeySelector `json:"secretRef,omitempty"`
	// Path contains the path in the local Pod filesystem where the CA
	// certificate file can be found.
	//
	// Use Path when the CA certificate file contents are managed by an
	// external system or sidecar.
	// +optional
	Path string `json:"path,omitempty"`
}

// MutualTLSPaths configures mTLS authentication to a Temporal API server using
// cert/key material loaded from file paths inside a Pod. These file paths are
// typically populated by an init container, sidecar container (e.g.
// spiffe-helper), or CSI driver.
type MutualTLSPaths struct {
	// Cert is the path to the PEM-encoded client certificate.
	// +kubebuilder:validation:MinLength=1
	Cert string `json:"cert"`

	// Key is the path to the PEM-encoded client private key.
	// +kubebuilder:validation:MinLength=1
	Key string `json:"key"`

	// CA is an optional path to a PEM-encoded CA bundle used to verify the
	// Temporal API server's certificate. If empty, the system CA pool is used.
	// +optional
	CA string `json:"ca,omitempty"`
}
