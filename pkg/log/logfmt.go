package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
)

type logfmtHandler struct {
	sync.Mutex
	out  io.Writer
	opts *slog.HandlerOptions
}

func (h *logfmtHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

func (h *logfmtHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h2 := logfmtHandler{
		out:  h.out,
		opts: h.opts,
	}
	return &h2 // simplified: real impl should store attrs
}

func (h *logfmtHandler) WithGroup(name string) slog.Handler {
	return &logfmtHandler{out: h.out, opts: h.opts}
}

func (h *logfmtHandler) Handle(_ context.Context, r slog.Record) error {
	var buf strings.Builder

	fmt.Fprintf(&buf, "time=%s level=%s msg=%s",
		r.Time.Format("2006-01-02T15:04:05.000Z07:00"),
		r.Level,
		quote(r.Message),
	)

	r.Attrs(func(a slog.Attr) bool {
		appendAttr(&buf, a)
		return true
	})

	buf.WriteByte('\n')

	h.Lock()
	defer h.Unlock()
	_, err := h.out.Write([]byte(buf.String()))
	return err
}

func appendAttr(buf *strings.Builder, a slog.Attr) {
	switch a.Value.Kind() {
	case slog.KindGroup:
		// Flatten group: prefix keys with group name
		for _, ga := range a.Value.Group() {
			key := a.Key + "." + ga.Key
			fmt.Fprintf(buf, " %s=%s", key, formatValue(ga.Value))
		}
	default:
		fmt.Fprintf(buf, " %s=%s", a.Key, formatValue(a.Value))
	}
}

func formatValue(v slog.Value) string {
	switch v.Kind() {
	case slog.KindString:
		return quote(v.String())
	case slog.KindDuration:
		return fmt.Sprintf("%d", v.Duration().Nanoseconds())
	case slog.KindTime:
		return quote(v.Time().Format("2006-01-02T15:04:05.000Z07:00"))
	default:
		return quote(fmt.Sprintf("%v", v.Any()))
	}
}

// quote wraps the value in double quotes if it contains spaces, =, or newlines.
func quote(s string) string {
	if s == "" || strings.ContainsAny(s, " \t\n\r\"=") {
		return fmt.Sprintf("%q", s)
	}
	return s
}
