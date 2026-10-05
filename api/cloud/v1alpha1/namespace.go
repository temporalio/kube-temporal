package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	tmprlcorev1 "github.com/temporalio/kube-temporal/api/core/v1"
)

// Namespace represents an instance of a Temporal Cloud Namespace.
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Namespace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NamespaceSpec   `json:"spec,omitempty"`
	Status NamespaceStatus `json:"status,omitempty"`
}

// NamespaceSpec describes the desired state of a Temporal Cloud Namespace.
//
// Ref: https://saas-api.tmprl.cloud/docs/httpapi.html#tag/namespaces/POST/cloud/namespaces
// +k8s:openapi-gen=true
type NamespaceSpec struct {
	// ProjectID is the The id of the project in which the Temporal Cloud
	// Namespace belongs. If not set, defaults to the account's default
	// project.
	ProjectID string `json:"projectID,omitempty"`
	// Tags is the collection of string tags to attach to the Namespace.
	Tags []string `json:"tags,omitempty"`
}

// NamespaceStatus is the status for a Namespace resource
// +k8s:openapi-gen=true
type NamespaceStatus struct {
	tmprlcorev1.StatusBase `json:",inline"`
}

// NamespaceList is a list of Namespace resources
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type NamespaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Namespace `json:"items"`
}
