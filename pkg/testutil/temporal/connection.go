package temporal

import (
	"context"

	tcorev1 "github.com/temporalio/kube-temporal/api/core/v1"
	"github.com/temporalio/kube-temporal/pkg/temporal/connection/loader"
	tclient "go.temporal.io/sdk/client"
)

const (
	FakeUserAgent      = "kube-temporal-testing"
	FakeHostPort       = "fake.testing.tmprl.cloud:443"
	FakeAPIKeyContents = "indy!bad-dates!"
)

// NewConnectionSpec returns a tcorev1.ConnectionSpec for
// connecting to the fake Temporal API endpoint.
func NewConnectionSpec(
	apiKeyPath string,
) *tcorev1.ConnectionSpec {
	return &tcorev1.ConnectionSpec{
		HostPort: FakeHostPort,
		Credentials: tcorev1.CredentialsConfig{
			APIKey: &tcorev1.APIKeyCredentialsConfig{
				Path: apiKeyPath,
			},
		},
	}
}

// NewClientOptions returns new Temporal client options configured
// with a noop logger.
func NewClientOptions() *tclient.Options {
	return &tclient.Options{
		Logger: NewNoopLogger(),
	}
}

// TrackingClient is a simple mock used for tracking the client.Options
// provided to a mocked Dial func.
type TrackingClient struct {
	tclient.Client
	Options tclient.Options
}

// NewTrackingClientDialFunc allows unit tests to "mock" the Temporal SDK
// tclient.Options.DialFn to return a TrackingClient containing the supplied
// expected tclient.Options.
func NewTrackingClientDialFunc(opts tclient.Options) (tclient.Client, error) {
	return &TrackingClient{Options: opts}, nil
}

// TrackingAPIKeyDynamicCredentials is a simple mock used for tracking
// client.Credentials returned from the Temporal SDK
// client.NewAPIKeyDynamicCredentials package-level function.
type TrackingAPIKeyDynamicCredentials struct {
	tclient.Credentials
	Value string // the value returned from the mocked dynamic credentials func
}

// NewTrackingAPIKeyDynamicCredentialsFunc allows unit tests to "mock" the
// Temporal SDK client.NewAPIKeyDynamicCredentials package-level function to
// return a TrackingAPIKeyDynamicCredentials struct containing the supplied API
// key.
func NewTrackingAPIKeyDynamicCredentialsFunc(
	value string, // value is either an API key path or API key data
) loader.NewAPIKeyDynamicCredentialsFunc {
	return func(_ func(_ context.Context) (string, error)) tclient.Credentials {
		return &TrackingAPIKeyDynamicCredentials{Value: value}
	}
}
