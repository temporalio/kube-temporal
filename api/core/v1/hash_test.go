package v1_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	tcorev1 "github.com/temporalio/kube-temporal/api/core/v1"
	kcorev1 "k8s.io/api/core/v1"
)

func TestHashDiffs(t *testing.T) {
	cases := []struct {
		name string
		a    tcorev1.ConnectionSpec
		b    tcorev1.ConnectionSpec
		exp  bool
	}{
		{
			name: "empty specs same hash",
			a:    tcorev1.ConnectionSpec{},
			b:    tcorev1.ConnectionSpec{},
			exp:  true,
		},
		{
			name: "same host port, same hash",
			a: tcorev1.ConnectionSpec{
				HostPort: "host:port",
			},
			b: tcorev1.ConnectionSpec{
				HostPort: "host:port",
			},
			exp: true,
		},
		{
			name: "diff host port, diff hash",
			a: tcorev1.ConnectionSpec{
				HostPort: "a:port",
			},
			b: tcorev1.ConnectionSpec{
				HostPort: "b:port",
			},
			exp: false,
		},
		{
			name: "same host port, diff identity, diff hash",
			a: tcorev1.ConnectionSpec{
				HostPort: "host:port",
				Identity: "a",
			},
			b: tcorev1.ConnectionSpec{
				HostPort: "host:port",
				Identity: "b",
			},
			exp: false,
		},
		{
			name: "same host port, same creds, same hash",
			a: tcorev1.ConnectionSpec{
				HostPort: "host:port",
				Credentials: tcorev1.CredentialsConfig{
					APIKey: &tcorev1.APIKeyCredentialsConfig{
						Path: "/my/api-key",
					},
				},
			},
			b: tcorev1.ConnectionSpec{
				HostPort: "host:port",
				Credentials: tcorev1.CredentialsConfig{
					APIKey: &tcorev1.APIKeyCredentialsConfig{
						Path: "/my/api-key",
					},
				},
			},
			exp: true,
		},
		{
			name: "same host port, diff creds, diff hash",
			a: tcorev1.ConnectionSpec{
				HostPort: "host:port",
				Credentials: tcorev1.CredentialsConfig{
					APIKey: &tcorev1.APIKeyCredentialsConfig{
						Path: "/my/api-key",
					},
				},
			},
			b: tcorev1.ConnectionSpec{
				HostPort: "host:port",
				Credentials: tcorev1.CredentialsConfig{
					APIKey: &tcorev1.APIKeyCredentialsConfig{
						SecretRef: &kcorev1.SecretKeySelector{
							LocalObjectReference: kcorev1.LocalObjectReference{
								Name: "secret",
							},
						},
					},
				},
			},
			exp: false,
		},
		{
			name: "same host port, same server TLS, same hash",
			a: tcorev1.ConnectionSpec{
				HostPort: "host:port",
				TLS: tcorev1.ServerTLSConfig{
					ServerName: "server",
				},
			},
			b: tcorev1.ConnectionSpec{
				HostPort: "host:port",
				TLS: tcorev1.ServerTLSConfig{
					ServerName: "server",
				},
			},
			exp: true,
		},
		{
			name: "same host port, diff server TLS, diff hash",
			a: tcorev1.ConnectionSpec{
				HostPort: "host:port",
				TLS: tcorev1.ServerTLSConfig{
					ServerName: "server-a",
				},
			},
			b: tcorev1.ConnectionSpec{
				HostPort: "host:port",
				TLS: tcorev1.ServerTLSConfig{
					ServerName: "server-b",
				},
			},
			exp: false,
		},
		{
			name: "delimiter ensures diff hash on similar fields",
			a: tcorev1.ConnectionSpec{
				HostPort: "host:port",
				Identity: "a",
			},
			b: tcorev1.ConnectionSpec{
				HostPort: "host:por",
				Identity: "ta",
			},
			exp: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(tt *testing.T) {
			require := require.New(tt)
			if tc.exp {
				require.Equal(tc.a.Hash(), tc.b.Hash())
			} else {
				require.NotEqual(tc.a.Hash(), tc.b.Hash())
			}
		})
	}
}
