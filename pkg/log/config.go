package log

import (
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/pflag"
)

const (
	DefaultLogLevel  = "info"
	flagLogLevel     = "logging-level"
	flagLogLevelDesc = "The log level. The default is info. The options are, in descending order of chattiness: debug, info, warn, error"
	EnvVarLogLevel   = "LOG_LEVEL"

	DefaultLogFormat  = "json"
	flagLogFormat     = "logging-format"
	flagLogFormatDesc = "The log format. The default is json. The options are: json, text, logfmt"
	EnvVarLogFormat   = "LOG_FORMAT"
)

// Config contains logging configuration options.
type Config struct {
	// Level is the log level to use.
	//
	// The default is "info". The options are, in descending order of
	// chattiness: "debug", "warn", "info", "error".
	Level string `json:"level"`
	// Format describes the format for the logger to use.
	//
	// The default is "json". The options are "json", "text" and "logfmt".
	Format string `json:"format"`
}

func (c *Config) SetDefaults() {
	if c.Level == "" {
		lvl := os.Getenv(EnvVarLogLevel)
		if lvl == "" {
			lvl = DefaultLogLevel
		}
		c.Level = lvl
	}
	if c.Format == "" {
		f := os.Getenv(EnvVarLogFormat)
		if f == "" {
			f = DefaultLogFormat
		}
		c.Format = f
	}
}

// BindFlags binds the supplied flagset to the Config's fields.
func (c *Config) BindFlags(fs *pflag.FlagSet) {
	pflag.StringVar(
		&c.Level,
		flagLogLevel,
		"", // empty string to allow defaulting to env var value
		flagLogLevelDesc,
	)
	pflag.StringVar(
		&c.Format,
		flagLogFormat,
		"", // empty string to allow defaulting to env var value
		flagLogFormatDesc,
	)
}

// Logger returns a slog.Logger initialized with the Config's options.
func (c Config) Logger() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: logLevelToSlogLevel(c.Level),
	}
	var handler slog.Handler
	switch c.Format {
	case "json":
		handler = slog.NewJSONHandler(os.Stderr, opts)
	case "text":
		handler = slog.NewTextHandler(os.Stderr, opts)
	default:
		handler = &logfmtHandler{out: os.Stderr, opts: opts}
	}
	return slog.New(handler)
}

// logLevelToSlogLevel translates the string log level to the slog log level
// integer type.
func logLevelToSlogLevel(lvl string) slog.Level {
	switch strings.ToLower(lvl) {
	case "error":
		return slog.LevelError
	case "warn":
		return slog.LevelWarn
	case "debug":
		return slog.LevelDebug
	default:
		return slog.LevelInfo
	}
}
