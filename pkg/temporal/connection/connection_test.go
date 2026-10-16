package connection_test

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/temporalio/kube-temporal/pkg/temporal/connection"
	tclient "go.temporal.io/sdk/client"
	tlog "go.temporal.io/sdk/log"
)

var noopLogger = tlog.NewStructuredLogger(slog.New(slog.DiscardHandler))

type mockTClient struct {
	tclient.Client
	Options tclient.Options
}

func mockTClientDialFunc(opts tclient.Options) (tclient.Client, error) {
	return &mockTClient{Options: opts}, nil
}

func TestNew(t *testing.T) {
	tc, err := tclient.NewLazyClient(
		tclient.Options{
			Logger: noopLogger,
		},
	)
	require.Nil(t, err)

	clientOpts := tclient.Options{}

	loader := connection.NewLoader(
		connection.WithTemporalClientLogger(noopLogger),
		connection.WithTemporalClientDialFunc(mockTClientDialFunc),
		connection.WithTemporalClientOptions(clientOpts),
	)

	cases := []struct {
		name      string
		args      []connection.NewOption
		expTC     tclient.Client
		expLoader *connection.Loader
		expErr    string
	}{
		{
			name:   "fail if no loader or tclient provided",
			args:   []connection.NewOption{},
			expErr: "either use WithLoader or WithTemporalClient",
		},
		{
			name: "fail if both loader and tclient provided",
			args: []connection.NewOption{
				connection.WithTemporalClient(tc),
				connection.WithLoader(loader),
			},
			expErr: "either use WithLoader or WithTemporalClient",
		},
		{
			name: "temporal client override",
			args: []connection.NewOption{
				connection.WithTemporalClient(tc),
			},
			expTC:     tc,
			expLoader: nil,
		},
		{
			name: "loader",
			args: []connection.NewOption{
				connection.WithLoader(loader),
			},
			expTC:     &mockTClient{Options: clientOpts},
			expLoader: &loader,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(tt *testing.T) {
			require := require.New(tt)
			got, err := connection.New(tc.args...)
			if tc.expErr != "" {
				require.ErrorContains(err, tc.expErr)
				return
			}
			require.NotNil(got)
			if tc.expTC == nil {
				gotTC, err := got.Client()
				require.Nil(err)
				require.Nil(gotTC)
			} else {
				gotTC, err := got.Client()
				require.Nil(err)
				require.NotNil(gotTC)
				require.Equal(tc.expTC, gotTC)
			}
			gotLoader := got.Loader()
			if tc.expLoader == nil {
				require.Nil(gotLoader)
			} else {
				require.NotNil(gotLoader)
				gotTCOpts, err := gotLoader.TemporalClientOptions()
				require.Nil(err)
				expTCOpts, err := tc.expLoader.TemporalClientOptions()
				require.Nil(err)
				require.EqualExportedValues(expTCOpts, gotTCOpts)
			}
		})
	}
}
