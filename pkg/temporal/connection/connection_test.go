package connection_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/temporalio/kube-temporal/pkg/temporal/connection"
	"github.com/temporalio/kube-temporal/pkg/temporal/connection/loader"
	tutmprl "github.com/temporalio/kube-temporal/pkg/testutil/temporal"
	tclient "go.temporal.io/sdk/client"
)

func TestNew(t *testing.T) {
	tc, err := tclient.NewLazyClient(
		tclient.Options{
			Logger: tutmprl.NewNoopLogger(),
		},
	)
	require.Nil(t, err)

	clientOpts := tclient.Options{}

	ldr := loader.New(
		loader.WithTemporalClientLogger(tutmprl.NewNoopLogger()),
		loader.WithTemporalClientDialFunc(tutmprl.NewTrackingClientDialFunc),
		loader.WithTemporalClientOptions(clientOpts),
	)

	cases := []struct {
		name      string
		args      []connection.NewOption
		expTC     tclient.Client
		expLoader *loader.Loader
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
				connection.WithLoader(ldr),
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
				connection.WithLoader(ldr),
			},
			expTC:     &tutmprl.TrackingClient{Options: clientOpts},
			expLoader: &ldr,
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
