package finalizer

import (
	"slices"
	"strings"

	"github.com/samber/lo"
)

// HasFinalizer returns true if the subject has a finalier matching the
// supplied search string.
//
// Note that the matching on finalizer name is case-insensitive.
func HasFinalizer(
	subject HasFinalizers,
	search string,
) bool {
	return lo.ContainsBy(
		subject.Finalizers(),
		func(name string) bool {
			return strings.EqualFold(name, search)
		},
	)
}

// AddFinalizer adds the supplied finalizer to the subject.
func AddFinalizer(subject Manager, name string) {
	if HasFinalizer(subject, name) {
		return
	}
	finalizers := subject.Finalizers()
	finalizers = append(finalizers, name)
	subject.ReplaceFinalizers(finalizers)
}

// RemoveFinalizer removes the supplied finalizer from the subject.
func RemoveFinalizer(subject Manager, name string) {
	if !HasFinalizer(subject, name) {
		return
	}
	finalizers := subject.Finalizers()
	finalizers = slices.DeleteFunc(finalizers, func(s string) bool {
		return strings.EqualFold(name, s)
	})
	subject.ReplaceFinalizers(finalizers)
}
