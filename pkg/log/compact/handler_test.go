package compact_test

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/temporalio/kube-temporal/pkg/log/compact"
)

func TestLoggerFormat(t *testing.T) {
	require := require.New(t)
	var b strings.Builder

	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	logger := slog.New(compact.NewHandler(&b, opts))

	require.NotNil(logger)

	logger.Debug("foo")

	got := b.String()
	require.Contains(got, " DEBUG \"foo\"")
	b.Reset()

	// test that attrs in no group produce properly-formatted k=v pairs in the
	// log output
	logger.Info("foo", "bar", "baz")

	got = b.String()
	require.Contains(got, " INFO \"foo\"")
	require.Contains(got, " bar=\"baz\"")
	b.Reset()

	// test that groups and attrs produce properly-formatted k=v pairs in the
	// log output with the group key prefix.
	groupLogger := logger.WithGroup("group").With("bar", "baz")

	groupLogger.Debug("foo", "bar", "baz")

	got = b.String()
	require.Contains(got, " DEBUG \"foo\"")
	require.Contains(got, " group.bar=\"baz\"")
	b.Reset()

	// tests two groups deep logger correctly prepends key prefix.
	twoGroupLogger := logger.WithGroup("group1").WithGroup("group2").With("bar", "baz")

	twoGroupLogger.Debug("foo")

	got = b.String()
	require.Contains(got, " DEBUG \"foo\"")
	require.Contains(got, " group1.group2.bar=\"baz\"")
	b.Reset()

	// test preformatting grouped attrs preserves group key prefix.
	twoGroupLoggerPreserve := logger.WithGroup("group1").With("bar", "baz")

	twoGroupLoggerPreserve.Debug("foo", "nonpreformat", "string")

	got = b.String()
	require.Contains(got, " DEBUG \"foo\"")
	require.Contains(got, " group1.bar=\"baz\"")
	require.Contains(got, " group1.nonpreformat=\"string\"")
	b.Reset()

	twoGroupLoggerPreserve = twoGroupLoggerPreserve.WithGroup("group2").With("bob", "alice")

	twoGroupLoggerPreserve.Debug("foo")

	got = b.String()
	require.Contains(got, " DEBUG \"foo\"")
	require.Contains(got, " group1.bar=\"baz\"")
	// because these attrs were not set with With(), they should not be saved
	// in preformatted attr string.
	require.NotContains(got, " group1.nonpreformat=\"string\"")
	require.Contains(got, " group1.group2.bob=\"alice\"")
	b.Reset()
}
