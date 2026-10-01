package condition

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kstatus "sigs.k8s.io/cli-utils/pkg/kstatus/status"
)

// Reconciling returns the Condition in the resource's Conditions collection
// that is of type kstatus.ConditionReconciling. If no such condition is found,
// returns nil.
func Reconciling(subject Manager) *metav1.Condition {
	return FirstOfType(subject, kstatus.ConditionReconciling)
}

// IsReconciling returns true if the resource's Conditions collection contains
// a Condition of type kstatus.ConditionReconciling and that Condition is True.
func IsReconciling(subject Manager) bool {
	c := Reconciling(subject)
	if c == nil {
		return false
	}
	return c.Status == metav1.ConditionTrue
}

// SetReconciling sets the subject's Condition of type
// kstatus.ConditionReconciling to the supplied status, optional message and
// reason.
func SetReconciling(
	subject Manager,
	status metav1.ConditionStatus,
	message string,
	reason string,
) {
	allConds := subject.Conditions()
	var c *metav1.Condition
	c = Reconciling(subject)
	if c == nil {
		c = &metav1.Condition{
			Type: kstatus.ConditionReconciling.String(),
		}
		allConds = append(allConds, c)
	}
	now := metav1.Now()
	c.LastTransitionTime = now
	c.Status = status
	c.Message = message
	c.Reason = reason
	subject.ReplaceConditions(allConds)
}

// Stalled returns the Condition in the resource's Conditions collection that
// is of type kstatus.ConditionStalled. If no such condition is found, returns
// nil.
func Stalled(subject Manager) *metav1.Condition {
	return FirstOfType(subject, kstatus.ConditionStalled)
}

// IsStalled returns true if the resource's Conditions collection contains a
// Condition of type kstatus.ConditionStalled and that Condition is True.
func IsStalled(subject Manager) bool {
	c := Stalled(subject)
	if c == nil {
		return false
	}
	return c.Status == metav1.ConditionTrue
}

// SetStalled sets the resource's Condition of type kstatus.ConditionStalled to
// the supplied status, optional message and reason.
func SetStalled(
	subject Manager,
	status metav1.ConditionStatus,
	message string,
	reason string,
) {
	allConds := subject.Conditions()
	var c *metav1.Condition
	c = Stalled(subject)
	if c == nil {
		c = &metav1.Condition{
			Type: kstatus.ConditionStalled.String(),
		}
		allConds = append(allConds, c)
	}
	now := metav1.Now()
	c.LastTransitionTime = now
	c.Status = status
	c.Message = message
	c.Reason = reason
	subject.ReplaceConditions(allConds)
}
