package compare

import (
	"fmt"
	"reflect"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/temporalio/kube-temporal/pkg/types"
)

// EqualIgnoringFields returns true if the supplied expected and got arguments
// are equal but ignoring any specified fields by field path.
func EqualIgnoringFields(
	expected any,
	got any,
	ignore ...string,
) bool {
	diff := DiffIgnoringFields(expected, got, ignore...)
	return diff == ""
}

// DiffIgnoringFields returns the formatted diffference between a supplied
// expected and got argument while ignoring any specified fields by field path.
func DiffIgnoringFields(
	expected any,
	got any,
	ignore ...string,
) string {
	typ := structTypeFromAny(expected)
	opts := []cmp.Option{
		// examine all unexported fields to prevent panics.
		cmp.Exporter(func(_ reflect.Type) bool { return true }),
		cmpopts.IgnoreFields(typ, ignore...),
	}
	return cmp.Diff(expected, got, opts...)
}

func structTypeFromAny(expected any) any {
	subject := expected
	res, isRes := expected.(types.Resource)
	if isRes {
		subject = res.CR()
	}

	// Drill down into the supplied type, popping pointer to value for each
	// found pointer.
	val := reflect.ValueOf(subject)
	typ := reflect.TypeOf(subject)
	for {
		if typ.Kind() == reflect.Ptr {
			val = val.Elem()
			subject = val.Interface()
			typ = val.Type()
			continue
		}
		if typ.Kind() != reflect.Struct {
			panic(
				fmt.Sprintf(
					"subject must be a struct, but found %T",
					subject,
				),
			)
		}
		break
	}
	return subject
}
