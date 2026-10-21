package v1

import (
	"crypto/md5"

	corev1 "k8s.io/api/core/v1"
)

// CredentialsConfig contains instructions as to where Temporal client
// credentials can be found.
//
// When API Key authentication is used, `apiKey` should be non-nil.
//
// When Mutual TLS authentication is used, `mutualTLS` should be non-nil.
//
// When no authentication is used, both `apiKey` nor `mutualTLS` should by nil.
//
// +kubebuilder:validation:XValidation:rule="((!has(self.apiKey) && !has(self.mutualTLS)) || (has(self.apiKey) != has(self.mutualTLS)))",message="apiKey and mutualTLS are mutually exclusive."
type CredentialsConfig struct {
	// APIKey contains the optional configuration for connecting to a Temporal
	// API server using Temporal API Key authentication.
	// +optional
	APIKey *APIKeyCredentialsConfig `json:"apiKey,omitempty"`
	// MutualTLS contains the optional configuration for connecting to a
	// Temporal API server using Temporal Mutual TLS authentication.
	// +optional
	MutualTLS *MutualTLSCredentialsConfig `json:"mutualTLS,omitempty"`
}

// Hash returns a deterministic MD5 hash of the CredentialsConfig's field
// components.
func (c CredentialsConfig) Hash() []byte {
	out := md5.New()
	if c.APIKey != nil {
		out.Write([]byte{'|'})
		out.Write(c.APIKey.Hash())
	} else if c.MutualTLS != nil {
		out.Write([]byte{'|'})
		out.Write(c.MutualTLS.Hash())
	}
	return out.Sum(nil)
}

// APIKeyCredentialsConfig contains configuration for connecting to a Temporal server
// using Temporal API Key authentication.
//
// +kubebuilder:validation:XValidation:rule="has(self.secretRef) != has(self.path)",message="Only one of secretRef or path may be set"
type APIKeyCredentialsConfig struct {
	// SecretRef selects the Secret key that contains the API key used for
	// authentication.
	//
	// The Secret must be `type: kubernetes.io/opaque` and must exist in the
	// same Kubernetes namespace as the Connection resource.
	//
	// SecretKeySelector encodes both:
	//   - LocalObjectReference.Name: the name of the Secret resource
	//   - Key: the data key within Secret.Data whose value is the API key token
	// +optional
	SecretRef *corev1.SecretKeySelector `json:"secretRef,omitempty"`
	// Path is the local file path in the Pod where the API Key can be found.
	// +optional
	Path string `json:"path,omitempty"`
}

// Hash returns a deterministic MD5 hash of the APIKeyCredentialsConfig's field
// components.
func (c APIKeyCredentialsConfig) Hash() []byte {
	out := md5.New()
	if c.SecretRef != nil {
		out.Write([]byte{'|'})
		out.Write([]byte(c.SecretRef.Name))
		out.Write([]byte{'|'})
		out.Write([]byte(c.SecretRef.Key))
	} else if c.Path != "" {
		out.Write([]byte{'|'})
		out.Write([]byte(c.Path))
	}
	return out.Sum(nil)
}

// MutualTLSCredentialsConfig defines where to find TLS key pair and
// certificate material when a Connection uses Mutual TLS authentication.
//
// See: https://docs.temporal.io/cloud/certificates
//
// You can populate the mTLS private key and certificate material via a
// Kubernetes Secret using the `secretRef` field.
//
// Alternately, you can write the private key and certificate material to a
// file paths in the Pod and use the `paths` field to tell the controller about
// these locations.
//
// +kubebuilder:validation:XValidation:rule="has(self.secretRef) != has(self.paths)",message="Only one of secretRef or paths may be set"
type MutualTLSCredentialsConfig struct {
	// SecretRef references a Kubernetes Secret that is either `type: opaque`
	// or `type: kubernetes.io/tls`.
	//
	// When the Secret type is `opaque`, the `SecretKeySelector.type` field
	// should contain the key in the Secret data where the TLS key and
	// certificate material bundle contents can be found. Opaque secrets are
	// useful when bundling tls.crt, tls.key, and ca.crt into a single secret
	// (e.g.  multi-file cert-manager outputs).
	//
	// When the Secret type is `kubernetes.io/tls`, the `SecretKeySelector.key`
	// field is ignored. The private key, certificate and CA certificate file
	// contents should be at the  standard `tls.key`, `tls.crt` and `ca.crt`
	// key locations in the Secret data.
	//
	// More information about creating a TLS secret:
	// https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets
	// +optional
	SecretRef *corev1.SecretKeySelector `json:"secretRef,omitempty"`
	// Paths points to where the mTLS private key and certificate material can
	// be found.
	// +optional
	Paths *MutualTLSPaths `json:"paths,omitempty"`
}

// Hash returns a deterministic MD5 hash of the MutualTLSCredentialsConfig's field
// components.
func (c MutualTLSCredentialsConfig) Hash() []byte {
	out := md5.New()
	if c.SecretRef != nil {
		out.Write([]byte{'|'})
		out.Write([]byte(c.SecretRef.Name))
		out.Write([]byte{'|'})
		out.Write([]byte(c.SecretRef.Key))
	} else if c.Paths != nil {
		out.Write([]byte{'|'})
		out.Write(c.Paths.Hash())
	}
	return out.Sum(nil)
}
