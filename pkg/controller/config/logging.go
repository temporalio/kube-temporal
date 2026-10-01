package config

import (
	"github.com/spf13/pflag"
	"go.uber.org/zap/zapcore"
	"k8s.io/klog/v2"
	ctrlrt "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	DefaultLoggingLevel  = "info"
	flagLoggingLevel     = "logging-level"
	flagLoggingLevelDesc = "The log level. The default is info. The options are: debug, info, warn, error"

	DefaultLoggingDevelopmentEnabled  = false
	flagLoggingDevelopmentEnabled     = "logging-development-enabled"
	flagLoggingDevelopmentEnabledDesc = "Enable local development logging mode."
)

var (
	defaultLogLevel = zapcore.InfoLevel
)

// LoggingConfig contains logging configuration options for the Temporal
// Cluster Operator
type LoggingConfig struct {
	// Level is the log level to use.
	Level string `json:"level"`
	// DevelopmentEnabled turns on local development logging mode.
	DevelopmentEnabled bool `json:"developmentEnabled"`
}

// BindFlags bings the supplied flagset to the LoggingConfig's fields.
func (c *LoggingConfig) BindFlags(fs *pflag.FlagSet) {
	pflag.StringVar(
		&c.Level,
		flagLoggingLevel,
		DefaultLoggingLevel,
		flagLoggingLevelDesc,
	)
	pflag.BoolVar(
		&c.DevelopmentEnabled,
		flagLoggingDevelopmentEnabled,
		DefaultLoggingDevelopmentEnabled,
		flagLoggingDevelopmentEnabledDesc,
	)
}

// SetupLogging initializes the controller's logging based on the Config's
// Logging configuration.
// SetupLogger initializes the logger used in the service controller
func (c *Config) SetupLogging() {
	lvl := defaultLogLevel
	lvl.UnmarshalText([]byte(c.Logging.Level))

	zapOptions := zap.Options{
		Development: c.Logging.DevelopmentEnabled,
		Level:       lvl,
		TimeEncoder: zapcore.ISO8601TimeEncoder,
	}
	logger := zap.New(zap.UseFlagOptions(&zapOptions))
	ctrlrt.SetLogger(logger)
	klog.SetLogger(logger)
}
