package annotation

import (
	"slices"
	"strings"

	"github.com/samber/lo"
)

type Annotations map[string]string

// HasKey returns true if the subject has an annotation matching the supplied
// key.
//
// Note that the matching on Condition type is case-insensitive.
func HasKey(
	subject Manager,
	search string,
) bool {
	keys := lo.Keys(subject.Annotations())
	return lo.ContainsBy(
		keys,
		func(key string) bool {
			return strings.EqualFold(key, search)
		},
	)
}

// HasKeyPrefix returns true if the subject has any annotation matching the
// supplied key prefix.
//
// Note that the matching on Condition type is case-insensitive.
func HasKeyPrefix(
	subject Manager,
	search string,
) bool {
	search = strings.ToLower(search)
	keys := lo.Keys(subject.Annotations())
	return lo.ContainsBy(
		keys,
		func(key string) bool {
			key = strings.ToLower(key)
			return strings.HasPrefix(key, search)
		},
	)
}

// Value returns the value of the subject's annotation matching the supplied
// key and a bool indicating whether the key was found.
//
// Note that the matching on Condition type is case-insensitive.
func Value(
	subject Manager,
	search string,
) (string, bool) {
	annos := subject.Annotations()
	keys := lo.Filter(
		lo.Keys(annos),
		func(key string, _ int) bool {
			return strings.EqualFold(key, search)
		},
	)
	if len(keys) == 0 {
		return "", false
	}
	slices.Sort(keys)
	return annos[keys[0]], true
}

// HavingKeyPrefix returns Annotations whose key matches the supplied key
// prefix.
//
// Note that the matching on Condition type is case-insensitive.
func HavingKeyPrefix(
	subject Manager,
	search string,
) Annotations {
	search = strings.ToLower(search)
	annos := subject.Annotations()
	return lo.PickBy(
		annos,
		func(key string, _ string) bool {
			key = strings.ToLower(key)
			return strings.HasPrefix(key, search)
		},
	)
}

// Clear resets the resource's collection of Conditions to an empty list.
func Clear(subject Manager) {
	subject.ReplaceAnnotations(Annotations{})
}
