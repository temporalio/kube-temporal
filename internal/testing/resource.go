package testing

import (
	tmprlcorev1 "github.com/temporalio/kube-temporal/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ResourceSpec contains the desired state fields of a Resource
type ResourceSpec struct {
}

// ResourceStatus contains the observed state fields of a Resource
type ResourceStatus struct {
	tmprlcorev1.StatusBase
}

// ResourceStatusWithConditions returns a ResourceStatus populated with the
// supplied Conditions.
func ResourceStatusWithConditions(
	conds ...*metav1.Condition,
) ResourceStatus {
	return ResourceStatus{
		StatusBase: tmprlcorev1.StatusBase{
			Conditions: conds,
		},
	}
}

// Resource is a Custom Resource Definition designed to test the core library.
type Resource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResourceSpec   `json:"spec,omitempty"`
	Status ResourceStatus `json:"status,omitempty"`
}

// Conditions returns the Resource's collection of Conditions from its status
// field. Satisfies the `pkg/condition.Manager` interface.
func (r Resource) Conditions() []*metav1.Condition {
	return r.Status.Conditions
}

// ReplaceConditions replaces the Resource's collection of Conditions with the
// supplied collection. Satisfies the `pkg/condition.Manager` interface.
func (r *Resource) ReplaceConditions(conds []*metav1.Condition) {
	r.Status.Conditions = conds
}
