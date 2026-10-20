package resource

import (
	"fmt"

	"github.com/temporalio/kube-temporal/pkg/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	managedFieldsKey = "managedFields"
)

// CleanMetadata returns the supplied map[string]any, Resource or
// unstructured.Unstructed's metadata as a map[string]any with the Server-Side
// Apply managed fields garbage stripped.
func CleanMetadata(subject any) (map[string]any, error) {
	var md map[string]any
	switch subject := subject.(type) {
	case types.Resource:
		us, err := subject.Unstructured()
		if err != nil {
			return nil, fmt.Errorf("failed getting unstructured: %w", err)
		}
		var ok bool
		md, ok, err = unstructured.NestedMap(us.Object, "metadata")
		if err != nil {
			return nil, fmt.Errorf("failed getting metadata map: %w", err)
		}
		if !ok {
			return nil, nil
		}
	case unstructured.Unstructured:
		var err error
		var ok bool
		md, ok, err = unstructured.NestedMap(subject.Object, "metadata")
		if err != nil {
			return nil, fmt.Errorf("failed getting metadata map: %w", err)
		}
		if !ok {
			return nil, nil
		}
	case map[string]any:
		md = subject
	default:
		return nil, fmt.Errorf("unsupported type %T", subject)
	}
	unstructured.RemoveNestedField(md, managedFieldsKey)
	return md, nil
}
