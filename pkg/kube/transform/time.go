package transform

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TimeFromProto converts a protobuf Timestamp to a Kubernetes metav1.Time.
// A nil timestamp converts to nil.
func TimeFromProto(ts *timestamppb.Timestamp) *metav1.Time {
	if ts == nil {
		return nil
	}
	t := metav1.NewTime(ts.AsTime())
	return &t
}
