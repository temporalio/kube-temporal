package manager_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	tcorev1 "github.com/temporalio/kube-temporal/api/core/v1"
	"github.com/temporalio/kube-temporal/pkg/temporal/connection"
	"github.com/temporalio/kube-temporal/pkg/temporal/connection/loader"
	"github.com/temporalio/kube-temporal/pkg/temporal/connection/manager"
	"github.com/temporalio/kube-temporal/pkg/testutil/mock"
	tutmprl "github.com/temporalio/kube-temporal/pkg/testutil/temporal"
	tclient "go.temporal.io/sdk/client"
)

func TestManager_Register(t *testing.T) {
	cases := []struct {
		name    string
		newArgs []manager.NewOption
		subject any
		regOpts []manager.RegisterOption
		expErr  string
	}{
		{
			name:    "fail if subject is an unsupported type",
			newArgs: []manager.NewOption{},
			subject: int64(1),
			expErr:  "unsupported type int64",
		},
		{
			name:    "fail if HostPort subject has no reg opts",
			newArgs: []manager.NewOption{},
			subject: connection.HostPort("host:port"),
			expErr:  "connection specs or temporal client required to register HostPort or HostPortNamespace",
		},
		{
			name:    "fail if HostPortNamespace subject has no reg opts",
			newArgs: []manager.NewOption{},
			subject: connection.HostPortNamespace{
				HostPort:  connection.HostPort("host:port"),
				Namespace: "ns",
			},
			expErr: "connection specs or temporal client required to register HostPort or HostPortNamespace",
		},
		{
			name:    "fail if HostPort subject has unmatched conn spec",
			newArgs: []manager.NewOption{},
			subject: connection.HostPort("host:port"),
			regOpts: []manager.RegisterOption{
				manager.RegisterWithConnectionSpecs(
					&tcorev1.ConnectionSpec{
						HostPort: "differenthost:port",
					},
				),
			},
			expErr: "connection spec HostPort \"differenthost:port\" does not match subject HostPort \"host:port\"",
		},
		{
			name:    "fail if HostPortNamespace subject has unmatched conn spec",
			newArgs: []manager.NewOption{},
			subject: connection.HostPortNamespace{
				HostPort:  connection.HostPort("host:port"),
				Namespace: "ns",
			},
			regOpts: []manager.RegisterOption{
				manager.RegisterWithConnectionSpecs(
					&tcorev1.ConnectionSpec{
						HostPort:  "host:port",
						Namespace: "different-ns",
					},
				),
			},
			expErr: "connection spec Namespace \"different-ns\" does not match subject Namespace \"ns\"",
		},
		{
			name:    "HostPort subject with conn spec",
			newArgs: []manager.NewOption{},
			subject: connection.HostPort("host:port"),
			regOpts: []manager.RegisterOption{
				manager.RegisterWithConnectionSpecs(
					&tcorev1.ConnectionSpec{
						HostPort: "host:port",
					},
				),
			},
		},
		{
			name:    "HostPortNamespace subject with conn spec",
			newArgs: []manager.NewOption{},
			subject: connection.HostPortNamespace{
				HostPort:  connection.HostPort("host:port"),
				Namespace: "ns",
			},
			regOpts: []manager.RegisterOption{
				manager.RegisterWithConnectionSpecs(
					&tcorev1.ConnectionSpec{
						HostPort:  "host:port",
						Namespace: "ns",
					},
				),
			},
		},
		{
			name:    "HostPort subject with client override",
			newArgs: []manager.NewOption{},
			subject: connection.HostPort("host:port"),
			regOpts: []manager.RegisterOption{
				manager.RegisterWithClient(new(mock.TemporalClient)),
			},
		},
		{
			name:    "HostPortNamespace subject with client override",
			newArgs: []manager.NewOption{},
			subject: connection.HostPortNamespace{
				HostPort:  connection.HostPort("host:port"),
				Namespace: "ns",
			},
			regOpts: []manager.RegisterOption{
				manager.RegisterWithClient(new(mock.TemporalClient)),
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(tt *testing.T) {
			require := require.New(tt)
			m := manager.New(tc.newArgs...)

			err := m.Register(tc.subject, tc.regOpts...)
			if tc.expErr != "" {
				require.ErrorContains(err, tc.expErr)
				return
			}
			require.Nil(err)
		})
	}
}

type registerFunc func(*manager.Manager) error

func TestManager_Client(t *testing.T) {
	tc := new(mock.TemporalClient)

	apiKeyDir := t.TempDir()
	apiKeyPath := filepath.Join(apiKeyDir, "api-key")

	err := os.WriteFile(
		apiKeyPath,
		[]byte(tutmprl.FakeAPIKeyContents),
		0644,
	)
	require.Nil(t, err)

	apiKeyCreds := &tutmprl.TrackingAPIKeyDynamicCredentials{
		Value: apiKeyPath,
	}

	cases := []struct {
		name     string
		m        *manager.Manager
		subject  any
		register []registerFunc
		expErr   string
		exp      tclient.Client
	}{
		{
			name:    "fail if subject is an unsupported type",
			m:       manager.New(),
			subject: int64(1),
			expErr:  "unsupported type int64",
		},
		{
			name:    "fail if HostPort subject has no client or spec registered",
			m:       manager.New(),
			subject: connection.HostPort("host:port"),
			expErr:  "no client or connection specs found for HostPort \"host:port\" and Namespace \"default\"",
		},
		{
			name: "fail if HostPortNamespace subject has no client or spec registered",
			m:    manager.New(),
			subject: connection.HostPortNamespace{
				HostPort:  connection.HostPort("host:port"),
				Namespace: "ns",
			},
			expErr: "no client or connection specs found for HostPort \"host:port\" and Namespace \"ns\"",
		},
		{
			name:    "fails HostPort subject with invalid conn spec",
			m:       manager.New(),
			subject: connection.HostPort("host:port"),
			register: []registerFunc{
				func(m *manager.Manager) error {
					return m.Register(
						connection.HostPort("host:port"),
						manager.RegisterWithConnectionSpecs(
							&tcorev1.ConnectionSpec{
								HostPort: "host:port",
								Credentials: tcorev1.CredentialsConfig{
									APIKey: &tcorev1.APIKeyCredentialsConfig{
										Path: "/does/not/exist",
									},
								},
							},
						),
					)
				},
			},
			expErr: "/does/not/exist: no such file or directory",
		},
		{
			name: "fails HostPortNamespace subject with invalid conn spec",
			m:    manager.New(),
			subject: connection.HostPortNamespace{
				HostPort:  connection.HostPort("host:port"),
				Namespace: "ns",
			},
			register: []registerFunc{
				func(m *manager.Manager) error {
					return m.Register(
						connection.HostPortNamespace{
							HostPort:  connection.HostPort("host:port"),
							Namespace: "ns",
						},
						manager.RegisterWithConnectionSpecs(
							&tcorev1.ConnectionSpec{
								HostPort:  "host:port",
								Namespace: "ns",
								Credentials: tcorev1.CredentialsConfig{
									APIKey: &tcorev1.APIKeyCredentialsConfig{
										Path: "/does/not/exist",
									},
								},
							},
						),
					)
				},
			},
			expErr: "/does/not/exist: no such file or directory",
		},
		{
			name:    "HostPort subject yields registered client",
			m:       manager.New(),
			subject: connection.HostPort("host:port"),
			register: []registerFunc{
				func(m *manager.Manager) error {
					return m.Register(
						connection.HostPort("host:port"),
						manager.RegisterWithClient(tc),
					)
				},
			},
			// We're testing the "pass through statically registered client
			// associated with the subject hostport" code paths here.
			exp: tc,
		},
		{
			name: "HostPortNamespace subject yields registered client",
			m:    manager.New(),
			subject: connection.HostPortNamespace{
				HostPort:  connection.HostPort("host:port"),
				Namespace: "ns",
			},
			register: []registerFunc{
				func(m *manager.Manager) error {
					return m.Register(
						connection.HostPortNamespace{
							HostPort:  connection.HostPort("host:port"),
							Namespace: "ns",
						},
						manager.RegisterWithClient(tc),
					)
				},
			},
			// We're testing the "pass through statically registered client
			// associated with the subject hostport" code paths here.
			exp: tc,
		},
		{
			name: "fails spec subject with invalid conn spec",
			m:    manager.New(),
			subject: &tcorev1.ConnectionSpec{
				HostPort: "host:port",
				Credentials: tcorev1.CredentialsConfig{
					APIKey: &tcorev1.APIKeyCredentialsConfig{
						Path: "/does/not/exist",
					},
				},
			},
			register: []registerFunc{
				func(m *manager.Manager) error {
					return m.Register(
						connection.HostPort("host:port"),
						manager.RegisterWithConnectionSpecs(
							&tcorev1.ConnectionSpec{
								HostPort: "host:port",
								Credentials: tcorev1.CredentialsConfig{
									APIKey: &tcorev1.APIKeyCredentialsConfig{
										Path: "/does/not/exist",
									},
								},
							},
						),
					)
				},
			},
			expErr: "/does/not/exist: no such file or directory",
		},
		{
			name: "spec subject yields registered client",
			m:    manager.New(),
			subject: &tcorev1.ConnectionSpec{
				HostPort: "host:port",
			},
			register: []registerFunc{
				func(m *manager.Manager) error {
					return m.Register(
						&tcorev1.ConnectionSpec{
							HostPort: "host:port",
						},
						manager.RegisterWithClient(tc),
					)
				},
			},
			// We're testing the "pass through statically registered client
			// associated with the spec hash" code paths here.
			exp: tc,
		},
		{
			name: "HostPort subject constructs client from registered conn spec",
			m: manager.New(
				manager.WithLoaderOptions(
					loader.WithTemporalClientDialFunc(tutmprl.NewTrackingClientDialFunc),
					loader.WithNewAPIKeyDynamicCredentialsFunc(
						tutmprl.NewTrackingAPIKeyDynamicCredentialsFunc(
							apiKeyPath,
						),
					),
				),
			),
			subject: connection.HostPort("host:port"),
			register: []registerFunc{
				func(m *manager.Manager) error {
					return m.Register(
						connection.HostPort("host:port"),
						manager.RegisterWithConnectionSpecs(
							&tcorev1.ConnectionSpec{
								HostPort: "host:port",
								Credentials: tcorev1.CredentialsConfig{
									APIKey: &tcorev1.APIKeyCredentialsConfig{
										Path: apiKeyPath,
									},
								},
							},
						),
					)
				},
			},
			// We're testing the "lazily create client object from registered
			// spec matching the supplied hostport subject" code paths here.
			exp: &tutmprl.TrackingClient{
				Options: tclient.Options{
					HostPort:    "host:port",
					Credentials: apiKeyCreds,
				},
			},
		},
		{
			name: "HostPortNamespace subject constructs client from registered conn spec",
			m: manager.New(
				manager.WithLoaderOptions(
					loader.WithTemporalClientDialFunc(tutmprl.NewTrackingClientDialFunc),
					loader.WithNewAPIKeyDynamicCredentialsFunc(
						tutmprl.NewTrackingAPIKeyDynamicCredentialsFunc(
							apiKeyPath,
						),
					),
				),
			),
			subject: connection.HostPortNamespace{
				HostPort:  connection.HostPort("host:port"),
				Namespace: "ns",
			},
			register: []registerFunc{
				func(m *manager.Manager) error {
					return m.Register(
						connection.HostPortNamespace{
							HostPort:  connection.HostPort("host:port"),
							Namespace: "ns",
						},
						manager.RegisterWithConnectionSpecs(
							&tcorev1.ConnectionSpec{
								HostPort:  "host:port",
								Namespace: "ns",
								Credentials: tcorev1.CredentialsConfig{
									APIKey: &tcorev1.APIKeyCredentialsConfig{
										Path: apiKeyPath,
									},
								},
							},
						),
					)
				},
			},
			// We're testing the "lazily create client object from registered
			// spec matching the supplied hostport subject" code paths here.
			exp: &tutmprl.TrackingClient{
				Options: tclient.Options{
					HostPort:    "host:port",
					Namespace:   "ns",
					Credentials: apiKeyCreds,
				},
			},
		},
		{
			name: "spec subject constructs client from registered conn spec",
			m: manager.New(
				manager.WithLoaderOptions(
					loader.WithTemporalClientDialFunc(tutmprl.NewTrackingClientDialFunc),
					loader.WithNewAPIKeyDynamicCredentialsFunc(
						tutmprl.NewTrackingAPIKeyDynamicCredentialsFunc(
							apiKeyPath,
						),
					),
				),
			),
			subject: &tcorev1.ConnectionSpec{
				HostPort: "host:port",
				Credentials: tcorev1.CredentialsConfig{
					APIKey: &tcorev1.APIKeyCredentialsConfig{
						Path: apiKeyPath,
					},
				},
			},
			register: []registerFunc{
				func(m *manager.Manager) error {
					return m.Register(
						connection.HostPort("host:port"),
						manager.RegisterWithConnectionSpecs(
							&tcorev1.ConnectionSpec{
								HostPort: "host:port",
								Credentials: tcorev1.CredentialsConfig{
									APIKey: &tcorev1.APIKeyCredentialsConfig{
										Path: apiKeyPath,
									},
								},
							},
						),
					)
				},
			},
			// We're testing the "lazily create client object from registered
			// spec" code paths here...
			exp: &tutmprl.TrackingClient{
				Options: tclient.Options{
					HostPort:    "host:port",
					Credentials: apiKeyCreds,
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(tt *testing.T) {
			require := require.New(tt)

			for _, reg := range tc.register {
				err := reg(tc.m)
				require.Nil(err)
			}
			got, err := tc.m.Client(tc.subject)
			if tc.expErr != "" {
				require.ErrorContains(err, tc.expErr)
				return
			}
			require.Nil(err)
			if gotTrackingCli, ok := got.(*tutmprl.TrackingClient); ok {
				expTrackingCli, expOK := tc.exp.(*tutmprl.TrackingClient)
				require.True(expOK)
				require.Equal(expTrackingCli.Options, gotTrackingCli.Options)
			}
		})
	}
}
