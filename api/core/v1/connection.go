package v1

import (
	"crypto/md5"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConnectionSpec describes the desired configuration of a Connection to a
// Temporal API server. It describes where credentials can be found for
// instantiating a Temporal client object to that Temporal API server.
type ConnectionSpec struct {
	// HostPort is the host:port of the Temporal API server endpoint to connect
	// to.
	//
	// Defaults to "localhost:7233"
	//
	// This is a gRPC address and therefore can also support a
	// special-formatted address of "<resolver>:///<value>" that will use a
	// registered resolver. By default all hosts returned from the resolver
	// will be used in a round-robin fashion.
	//
	// The "dns" resolver is registered by and used by default.
	// +optional
	HostPort string `json:"hostPort,omitempty"`
	// Namespace is the **Temporal Namespace** associated with this Connection.
	// Defaults to "default".
	// +optional
	Namespace string `json:"namespace,omitempty"`
	// Identity is the identifier to associate with the Temporal
	// client.Options.Identity field.
	// Defaults to a string that includes hostname, groupName and process ID.
	// +optional
	Identity string `json:"identity,omitempty"`
	// TLS configures TLS behaviour for the Temporal server connection.
	// +optional
	TLS ServerTLSConfig `json:"tls,omitempty"`
	// Credentials contains desired configuration for the Temporal client
	// credentials.
	Credentials CredentialsConfig `json:"credentials"`
}

// Hash returns a deterministic MD5 hash of the ConnectionSpec's field
// components.
func (s ConnectionSpec) Hash() []byte {
	out := md5.New()
	if s.HostPort != "" {
		out.Write([]byte{'|'})
		out.Write([]byte(s.HostPort))
	}
	if s.Namespace != "" {
		out.Write([]byte{'|'})
		out.Write([]byte(s.Namespace))
	}
	if s.Identity != "" {
		out.Write([]byte{'|'})
		out.Write([]byte(s.Identity))
	}
	out.Write([]byte{'|'})
	out.Write(s.TLS.Hash())
	out.Write([]byte{'|'})
	out.Write(s.Credentials.Hash())
	return out.Sum(nil)
}

// ConnectionStatus defines the observed state of a Connection
type ConnectionStatus struct {
	StatusBase `json:",inline"`
}

//+kubebuilder:object:root=true
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Endpoint",type="string",JSONPath=".spec.hostPort",description="Temporal API server host:port"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Age"

// Connection describes a client Connection to a Temporal API server.
type Connection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConnectionSpec   `json:"spec,omitempty"`
	Status ConnectionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ConnectionList contains a list of Connection
type ConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Connection `json:"items"`
}
