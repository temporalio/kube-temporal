package condition

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// Manager describes a thing that can set and retrieve Condition
// objects.
type Manager interface {
	// Conditions returns the set of Conditions.
	Conditions() []*metav1.Condition
	// ReplaceConditions replaces the resource's set of Condition structs with
	// the supplied slice of Conditions.
	ReplaceConditions([]*metav1.Condition)
}
