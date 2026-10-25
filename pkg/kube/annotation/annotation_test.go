package annotation_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	itest "github.com/temporalio/kube-temporal/internal/testing"
	"github.com/temporalio/kube-temporal/pkg/kube/annotation"
)

func TestHasKey(t *testing.T) {
	annoKey := "temporal.io/hostport"
	annoValue := "localhost:0000"

	cases := []struct {
		name   string
		res    annotation.Manager
		search string
		exp    bool
	}{
		{
			name:   "empty annotations yields false",
			res:    &itest.Resource{},
			search: "temporal.io/hostport",
			exp:    false,
		},
		{
			name: "not found annotation yields false",
			res: &itest.Resource{
				ObjectMeta: itest.ResourceMetaWithAnnotations(
					annotation.Annotations{
						"otherkey": "othervalue",
					},
				),
			},
			search: annoKey,
			exp:    false,
		},
		{
			name: "found annotation yields true",
			res: &itest.Resource{
				ObjectMeta: itest.ResourceMetaWithAnnotations(
					annotation.Annotations{
						annoKey: annoValue,
					},
				),
			},
			search: annoKey,
			exp:    true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(tt *testing.T) {
			require := require.New(tt)
			got := annotation.HasKey(tc.res, tc.search)
			require.Equal(tc.exp, got)
		})
	}
}

func TestHasKeyPrefix(t *testing.T) {
	annoKeyPrefix := "temporal.io/"
	annoKey := "temporal.io/hostport"
	annoValue := "localhost:0000"

	cases := []struct {
		name   string
		res    annotation.Manager
		search string
		exp    bool
	}{
		{
			name:   "empty annotations yields false",
			res:    &itest.Resource{},
			search: "temporal.io/hostport",
			exp:    false,
		},
		{
			name: "not found annotation yields false",
			res: &itest.Resource{
				ObjectMeta: itest.ResourceMetaWithAnnotations(
					annotation.Annotations{
						"otherkey": "othervalue",
					},
				),
			},
			search: annoKeyPrefix,
			exp:    false,
		},
		{
			name: "found annotation yields true",
			res: &itest.Resource{
				ObjectMeta: itest.ResourceMetaWithAnnotations(
					annotation.Annotations{
						"otherkey": "othervalue",
						annoKey:    annoValue,
					},
				),
			},
			search: annoKeyPrefix,
			exp:    true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(tt *testing.T) {
			require := require.New(tt)
			got := annotation.HasKeyPrefix(tc.res, tc.search)
			require.Equal(tc.exp, got)
		})
	}
}

func TestValue(t *testing.T) {
	annoKey := "temporal.io/hostport"
	annoValue := "localhost:0000"

	cases := []struct {
		name     string
		res      annotation.Manager
		search   string
		expVal   string
		expFound bool
	}{
		{
			name:     "empty annotations yields false",
			res:      &itest.Resource{},
			search:   "temporal.io/hostport",
			expFound: false,
		},
		{
			name: "not found annotation yields false",
			res: &itest.Resource{
				ObjectMeta: itest.ResourceMetaWithAnnotations(
					annotation.Annotations{
						"otherkey": "othervalue",
					},
				),
			},
			search:   annoKey,
			expFound: false,
		},
		{
			name: "found annotation yields true",
			res: &itest.Resource{
				ObjectMeta: itest.ResourceMetaWithAnnotations(
					annotation.Annotations{
						annoKey: annoValue,
					},
				),
			},
			search:   annoKey,
			expFound: true,
			expVal:   annoValue,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(tt *testing.T) {
			require := require.New(tt)
			got, found := annotation.Value(tc.res, tc.search)
			require.Equal(tc.expFound, found)
			if tc.expFound {
				require.Equal(tc.expVal, got)
			}
		})
	}
}

func TestHavingKeyPrefix(t *testing.T) {
	annoKeyPrefix := "temporal.io/"
	annoKey := "temporal.io/hostport"
	annoValue := "localhost:0000"

	cases := []struct {
		name   string
		res    annotation.Manager
		search string
		exp    annotation.Annotations
	}{
		{
			name:   "empty annotations yields empty annotations",
			res:    &itest.Resource{},
			search: "temporal.io/hostport",
			exp:    annotation.Annotations{},
		},
		{
			name: "not found annotation yields empty annotations",
			res: &itest.Resource{
				ObjectMeta: itest.ResourceMetaWithAnnotations(
					annotation.Annotations{
						"otherkey": "othervalue",
					},
				),
			},
			search: annoKeyPrefix,
			exp:    annotation.Annotations{},
		},
		{
			name: "found annotation yields filtered annotations",
			res: &itest.Resource{
				ObjectMeta: itest.ResourceMetaWithAnnotations(
					annotation.Annotations{
						"otherkey": "othervalue",
						annoKey:    annoValue,
					},
				),
			},
			search: annoKeyPrefix,
			exp: annotation.Annotations{
				annoKey: annoValue,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(tt *testing.T) {
			require := require.New(tt)
			got := annotation.HavingKeyPrefix(tc.res, tc.search)
			require.Equal(tc.exp, got)
		})
	}
}
