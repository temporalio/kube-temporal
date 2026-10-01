package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StatusBase is a base struct that all CRDs' Status struct embed.
type StatusBase struct {
	// Conditions is the collection of Conditions associated with the resource.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []*metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" `
}
