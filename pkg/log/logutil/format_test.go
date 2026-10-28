package logutil_test

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/temporalio/kube-temporal/pkg/log/logutil"
)

func TestFormatSourceFileAndLine(t *testing.T) {
	cases := []struct {
		name  string
		frame runtime.Frame
		exp   string
	}{
		{
			"no file",
			runtime.Frame{},
			"",
		},
		{
			"filepath more than two dirs",
			runtime.Frame{
				File: "/home/src/github.com/temporalio/kube-temporal/pkg/log/logutil/format.go",
				Line: 42,
			},
			"log/logutil/format.go:42",
		},
		{
			"filepath less than two dirs",
			runtime.Frame{
				File: "/logutil/format.go",
				Line: 42,
			},
			"logutil/format.go:42",
		},
		{
			"filepath has two dirs",
			runtime.Frame{
				File: "/log/logutil/format.go",
				Line: 42,
			},
			"log/logutil/format.go:42",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require := require.New(t)
			got := logutil.FormatSourceFileAndLine(c.frame)
			require.Equal(c.exp, got)
		})
	}
}
