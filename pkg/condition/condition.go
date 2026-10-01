package condition

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FirstOfType returns the first Condition in the resource's Conditions
// collection of the supplied type. If no such condition is found, returns nil.
//
// Note that the matching on Condition type is case-insensitive.
func FirstOfType(
	subject Manager,
	condType fmt.Stringer,
) *metav1.Condition {
	for _, condition := range subject.Conditions() {
		if strings.EqualFold(condition.Type, condType.String()) {
			return condition
		}
	}
	return nil
}

// AllOfType returns a slice of Conditions in the resource's Conditions
// collection of the supplied type.
//
// Note that the matching on Condition type is case-insensitive.
func AllOfType(
	subject Manager,
	condType fmt.Stringer,
) []*metav1.Condition {
	res := []*metav1.Condition{}
	for _, condition := range subject.Conditions() {
		if strings.EqualFold(condition.Type, condType.String()) {
			res = append(res, condition)
		}
	}
	return res
}

// Clear resets the resource's collection of Conditions to an empty list.
func Clear(
	subject Manager,
) {
	subject.ReplaceConditions([]*metav1.Condition{})
}
