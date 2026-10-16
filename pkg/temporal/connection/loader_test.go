package connection_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	tcorev1 "github.com/temporalio/kube-temporal/api/core/v1"
	"github.com/temporalio/kube-temporal/pkg/temporal/connection"
	"github.com/temporalio/kube-temporal/pkg/testutil"
	"github.com/temporalio/kube-temporal/pkg/tlsutil"
	tclient "go.temporal.io/sdk/client"
	kcorev1 "k8s.io/api/core/v1"
	kcrfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	tlsCAName     = "test CA"
	tlsServerName = "localhost"
	tlsClientName = "test client"
	kubeNamespace = "kube-temporal"
)

var (
	once sync.Once
)

var validCAKeyPair = sync.OnceValues(
	func() (*testutil.TLSKeyPair, error) {
		return testutil.TLSNewCAKeyPair(tlsCAName)
	},
)

var validServerKeyPair = sync.OnceValues(
	func() (*testutil.TLSKeyPair, error) {
		caKP, err := validCAKeyPair()
		if err != nil {
			return nil, fmt.Errorf("failed generating CA key pair: %w", err)
		}
		return testutil.TLSNewServerKeyPair(tlsServerName, caKP)
	},
)

var validClientKeyPair = sync.OnceValues(
	func() (*testutil.TLSKeyPair, error) {
		caKP, err := validCAKeyPair()
		if err != nil {
			return nil, fmt.Errorf("failed generating CA key pair: %w", err)
		}
		return testutil.TLSNewClientKeyPair(tlsClientName, caKP)
	},
)

func newTCOptions() *tclient.Options {
	return &tclient.Options{
		Logger: noopLogger,
	}
}

func TestLoader_ClientOptions(t *testing.T) {
	tcOpts := newTCOptions()

	cases := []struct {
		name   string
		args   []connection.NewLoaderOption
		exp    *tclient.Options
		expErr string
	}{
		{
			name:   "fail if no connection spec or tclient opts",
			args:   []connection.NewLoaderOption{},
			expErr: "ConnectionSpec required",
		},
		{
			name: "tclient opts override provided",
			args: []connection.NewLoaderOption{
				connection.WithTemporalClientLogger(noopLogger),
				connection.WithTemporalClientOptions(*tcOpts),
			},
			exp: tcOpts,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(tt *testing.T) {
			require := require.New(tt)
			loader := connection.NewLoader(tc.args...)
			got, err := loader.TemporalClientOptions()
			if tc.expErr != "" {
				require.ErrorContains(err, tc.expErr)
				return
			}
			require.Nil(err)
			require.NotNil(got)
			require.EqualExportedValues(tc.exp, got)
		})
	}
}

func TestLoader_ClientOptions_ControllerSpec_NoCredentials(t *testing.T) {
	hostPort := "host:12345"
	tns := "temporal-namespace"

	caKP, err := validCAKeyPair()
	require.Nil(t, err)

	poolWithValidCACert, err := x509.SystemCertPool()
	require.Nil(t, err)
	added := poolWithValidCACert.AppendCertsFromPEM(caKP.Cert)
	require.True(t, added)

	tlsCACertSecretName := "tls-ca-cert-secret-name"
	tlsCACertOpaqueSecretDataKey := "ca.pem"

	tlsCACertOpaqueSecret := &kcorev1.Secret{}
	tlsCACertOpaqueSecret.Name = tlsCACertSecretName
	tlsCACertOpaqueSecret.Data = map[string][]byte{
		tlsCACertOpaqueSecretDataKey: caKP.Cert,
	}

	tlsCACertInvalidSecret := &kcorev1.Secret{}
	tlsCACertInvalidSecret.Name = tlsCACertSecretName
	tlsCACertInvalidSecret.Data = map[string][]byte{
		tlsCACertOpaqueSecretDataKey: []byte("invalid cert"),
	}

	tlsCACertTLSSecret := &kcorev1.Secret{}
	tlsCACertTLSSecret.Type = kcorev1.SecretTypeTLS
	tlsCACertTLSSecret.Name = tlsCACertSecretName
	tlsCACertTLSSecret.Data = map[string][]byte{
		"ca.crt": caKP.Cert,
	}

	cases := []struct {
		name   string
		args   []connection.NewLoaderOption
		exp    *tclient.Options
		expErr string
	}{
		{
			name: "empty spec yields default tclient opts",
			args: []connection.NewLoaderOption{
				connection.WithTemporalClientLogger(noopLogger),
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{},
				),
			},
			exp: newTCOptions(),
		},
		{
			name: "specify only host port",
			args: []connection.NewLoaderOption{
				connection.WithTemporalClientLogger(noopLogger),
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{
						HostPort: hostPort,
					},
				),
			},
			exp: func() *tclient.Options {
				tco := newTCOptions()
				tco.HostPort = hostPort
				return tco
			}(),
		},
		{
			name: "specify only namespace",
			args: []connection.NewLoaderOption{
				connection.WithTemporalClientLogger(noopLogger),
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{
						Namespace: tns,
					},
				),
			},
			exp: func() *tclient.Options {
				tco := newTCOptions()
				tco.Namespace = tns
				return tco
			}(),
		},
		{
			name: "specify tls server name",
			args: []connection.NewLoaderOption{
				connection.WithTemporalClientLogger(noopLogger),
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{
						TLS: tcorev1.ServerTLSConfig{
							ServerName: tlsServerName,
						},
					},
				),
			},
			exp: func() *tclient.Options {
				tco := newTCOptions()
				tco.ConnectionOptions.TLS = &tls.Config{
					ServerName: tlsServerName,
				}
				return tco
			}(),
		},
		{
			name: "fail if api key secret no kube client",
			args: []connection.NewLoaderOption{
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{
						TLS: tcorev1.ServerTLSConfig{
							CACert: &tcorev1.CACertConfig{
								SecretRef: &kcorev1.SecretKeySelector{
									LocalObjectReference: kcorev1.LocalObjectReference{
										Name: tlsCACertSecretName,
									},
									Key: tlsCACertOpaqueSecretDataKey,
								},
							},
						},
					},
				),
			},
			expErr: "Loader requires KubeClient to load CA cert from Secret",
		},
		{
			name: "fail if tls CA cert secret no such secret",
			args: []connection.NewLoaderOption{
				connection.WithKubeClient(
					kcrfake.NewFakeClient(),
				),
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{
						TLS: tcorev1.ServerTLSConfig{
							CACert: &tcorev1.CACertConfig{
								SecretRef: &kcorev1.SecretKeySelector{
									LocalObjectReference: kcorev1.LocalObjectReference{
										Name: tlsCACertSecretName,
									},
									Key: tlsCACertOpaqueSecretDataKey,
								},
							},
						},
					},
				),
			},
			expErr: "secrets \"tls-ca-cert-secret-name\" not found",
		},
		{
			name: "tls CA cert secret invalid cert",
			args: []connection.NewLoaderOption{
				connection.WithKubeClient(
					kcrfake.NewFakeClient(tlsCACertInvalidSecret),
				),
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{
						TLS: tcorev1.ServerTLSConfig{
							CACert: &tcorev1.CACertConfig{
								SecretRef: &kcorev1.SecretKeySelector{
									LocalObjectReference: kcorev1.LocalObjectReference{
										Name: tlsCACertSecretName,
									},
									Key: tlsCACertOpaqueSecretDataKey,
								},
							},
						},
					},
				),
			},
			expErr: "failed appending CA cert to root CAs",
		},
		{
			name: "tls CA cert opaque secret type",
			args: []connection.NewLoaderOption{
				connection.WithTemporalClientLogger(noopLogger),
				connection.WithKubeClient(
					kcrfake.NewFakeClient(tlsCACertOpaqueSecret),
				),
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{
						TLS: tcorev1.ServerTLSConfig{
							CACert: &tcorev1.CACertConfig{
								SecretRef: &kcorev1.SecretKeySelector{
									LocalObjectReference: kcorev1.LocalObjectReference{
										Name: tlsCACertSecretName,
									},
									Key: tlsCACertOpaqueSecretDataKey,
								},
							},
						},
					},
				),
			},
			exp: func() *tclient.Options {
				tco := newTCOptions()
				tco.ConnectionOptions.TLS = &tls.Config{
					RootCAs: poolWithValidCACert,
				}
				return tco
			}(),
		},
		{
			name: "tls CA cert tls secret tpe",
			args: []connection.NewLoaderOption{
				connection.WithTemporalClientLogger(noopLogger),
				connection.WithKubeClient(
					kcrfake.NewFakeClient(tlsCACertTLSSecret),
				),
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{
						TLS: tcorev1.ServerTLSConfig{
							CACert: &tcorev1.CACertConfig{
								SecretRef: &kcorev1.SecretKeySelector{
									LocalObjectReference: kcorev1.LocalObjectReference{
										Name: tlsCACertSecretName,
									},
									// NOTE(jaypipes): Deliberately leave Key
									// field empty since this Secret is of type
									// kubernetes.io/tls and we want to verify
									// the autoselecting of the fixed `ca.crt`
									// data key.
								},
							},
						},
					},
				),
			},
			exp: func() *tclient.Options {
				tco := newTCOptions()
				tco.ConnectionOptions.TLS = &tls.Config{
					RootCAs: poolWithValidCACert,
				}
				return tco
			}(),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(tt *testing.T) {
			require := require.New(tt)
			loader := connection.NewLoader(tc.args...)
			got, err := loader.TemporalClientOptions()
			if tc.expErr != "" {
				require.ErrorContains(err, tc.expErr)
				return
			}
			require.Nil(err)
			require.NotNil(got)
			require.EqualExportedValues(tc.exp, got)

			if tc.exp.ConnectionOptions.TLS != nil {
				require.NotNil(got.ConnectionOptions.TLS)
				require.True(
					tc.exp.ConnectionOptions.TLS.RootCAs.Equal(
						got.ConnectionOptions.TLS.RootCAs,
					),
				)
			}
		})
	}
}

type mockAPIKeyCreds struct {
	tclient.Credentials
	key string
}

func newMockAPIKeyDynamicCredsFunc(key string) connection.NewAPIKeyDynamicCredentialsFunc {
	return func(_ func(_ context.Context) (string, error)) tclient.Credentials {
		return &mockAPIKeyCreds{key: key}
	}
}

func TestLoader_ClientOptions_ControllerSpec_APIKeyCredentials(t *testing.T) {
	apiKeySecretName := "api-key-secret-name"
	apiKeySecretDataKey := "api-key-secret-data-key"
	apiKeySecretDataValue := "api-key-secret-data-value"

	apiKeySecret := &kcorev1.Secret{}
	apiKeySecret.Name = apiKeySecretName
	apiKeySecret.Data = map[string][]byte{
		apiKeySecretDataKey: []byte(apiKeySecretDataValue),
	}

	apiKeySecretCreds := &mockAPIKeyCreds{key: apiKeySecretDataValue}

	apiKeySecretWithNamespace := &kcorev1.Secret{}
	apiKeySecretWithNamespace.Name = apiKeySecretName
	apiKeySecretWithNamespace.Namespace = kubeNamespace
	apiKeySecretWithNamespace.Data = map[string][]byte{
		apiKeySecretDataKey: []byte(apiKeySecretDataValue),
	}

	apiKeyPath := t.TempDir()

	apiKeyPathCreds := &mockAPIKeyCreds{key: apiKeyPath}

	cases := []struct {
		name   string
		args   []connection.NewLoaderOption
		exp    *tclient.Options
		expErr string
	}{
		{
			name: "fail if api key secret no kube client",
			args: []connection.NewLoaderOption{
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{
						Credentials: tcorev1.CredentialsConfig{
							APIKey: &tcorev1.APIKeyCredentialsConfig{
								SecretRef: &kcorev1.SecretKeySelector{
									LocalObjectReference: kcorev1.LocalObjectReference{
										Name: apiKeySecretName,
									},
									Key: apiKeySecretDataKey,
								},
							},
						},
					},
				),
			},
			expErr: "Loader requires KubeClient to load API Key from Secret",
		},
		{
			name: "fail if api key secret no such secret",
			args: []connection.NewLoaderOption{
				connection.WithKubeClient(
					kcrfake.NewFakeClient(),
				),
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{
						Credentials: tcorev1.CredentialsConfig{
							APIKey: &tcorev1.APIKeyCredentialsConfig{
								SecretRef: &kcorev1.SecretKeySelector{
									LocalObjectReference: kcorev1.LocalObjectReference{
										Name: apiKeySecretName,
									},
									Key: apiKeySecretDataKey,
								},
							},
						},
					},
				),
			},
			expErr: "secrets \"api-key-secret-name\" not found",
		},
		{
			name: "fail if api key secret no such secret in namespace",
			args: []connection.NewLoaderOption{
				connection.WithKubeNamespace(kubeNamespace),
				connection.WithKubeClient(
					// This is the non-namespaced secret, which should not
					// match the lookup with the kubeNamespace override from
					// the loader options.
					kcrfake.NewFakeClient(apiKeySecret),
				),
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{
						Credentials: tcorev1.CredentialsConfig{
							APIKey: &tcorev1.APIKeyCredentialsConfig{
								SecretRef: &kcorev1.SecretKeySelector{
									LocalObjectReference: kcorev1.LocalObjectReference{
										Name: apiKeySecretName,
									},
									Key: apiKeySecretDataKey,
								},
							},
						},
					},
				),
			},
			expErr: "secrets \"api-key-secret-name\" not found",
		},
		{
			name: "api key secret",
			args: []connection.NewLoaderOption{
				connection.WithTemporalClientLogger(noopLogger),
				connection.WithKubeClient(
					kcrfake.NewFakeClient(apiKeySecret),
				),
				connection.WithNewAPIKeyDynamicCredentialsFunc(
					newMockAPIKeyDynamicCredsFunc(apiKeySecretDataValue),
				),
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{
						Credentials: tcorev1.CredentialsConfig{
							APIKey: &tcorev1.APIKeyCredentialsConfig{
								SecretRef: &kcorev1.SecretKeySelector{
									LocalObjectReference: kcorev1.LocalObjectReference{
										Name: apiKeySecretName,
									},
									Key: apiKeySecretDataKey,
								},
							},
						},
					},
				),
			},
			exp: func() *tclient.Options {
				tco := newTCOptions()
				tco.Credentials = apiKeySecretCreds
				return tco
			}(),
		},
		{
			name: "api key secret in namespace",
			args: []connection.NewLoaderOption{
				connection.WithKubeNamespace(kubeNamespace),
				connection.WithTemporalClientLogger(noopLogger),
				connection.WithKubeClient(
					kcrfake.NewFakeClient(apiKeySecretWithNamespace),
				),
				connection.WithNewAPIKeyDynamicCredentialsFunc(
					newMockAPIKeyDynamicCredsFunc(apiKeySecretDataValue),
				),
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{
						Credentials: tcorev1.CredentialsConfig{
							APIKey: &tcorev1.APIKeyCredentialsConfig{
								SecretRef: &kcorev1.SecretKeySelector{
									LocalObjectReference: kcorev1.LocalObjectReference{
										Name: apiKeySecretName,
									},
									Key: apiKeySecretDataKey,
								},
							},
						},
					},
				),
			},
			exp: func() *tclient.Options {
				tco := newTCOptions()
				tco.Credentials = apiKeySecretCreds
				return tco
			}(),
		},
		{
			name: "api key secret with tls server name",
			args: []connection.NewLoaderOption{
				connection.WithTemporalClientLogger(noopLogger),
				connection.WithKubeClient(
					kcrfake.NewFakeClient(apiKeySecret),
				),
				connection.WithNewAPIKeyDynamicCredentialsFunc(
					newMockAPIKeyDynamicCredsFunc(apiKeySecretDataValue),
				),
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{
						TLS: tcorev1.ServerTLSConfig{
							ServerName: tlsServerName,
						},
						Credentials: tcorev1.CredentialsConfig{
							APIKey: &tcorev1.APIKeyCredentialsConfig{
								SecretRef: &kcorev1.SecretKeySelector{
									LocalObjectReference: kcorev1.LocalObjectReference{
										Name: apiKeySecretName,
									},
									Key: apiKeySecretDataKey,
								},
							},
						},
					},
				),
			},
			exp: func() *tclient.Options {
				tco := newTCOptions()
				tco.ConnectionOptions.TLS = &tls.Config{
					ServerName: tlsServerName,
				}
				tco.Credentials = apiKeySecretCreds
				return tco
			}(),
		},
		{
			name: "fail if api key path not found",
			args: []connection.NewLoaderOption{
				connection.WithNewAPIKeyDynamicCredentialsFunc(
					newMockAPIKeyDynamicCredsFunc(apiKeyPath),
				),
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{
						Credentials: tcorev1.CredentialsConfig{
							APIKey: &tcorev1.APIKeyCredentialsConfig{
								Path: "path/does/not/exist",
							},
						},
					},
				),
			},
			expErr: "path/does/not/exist: no such file or directory",
		},
		{
			name: "api key path",
			args: []connection.NewLoaderOption{
				connection.WithTemporalClientLogger(noopLogger),
				connection.WithNewAPIKeyDynamicCredentialsFunc(
					newMockAPIKeyDynamicCredsFunc(apiKeyPath),
				),
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{
						Credentials: tcorev1.CredentialsConfig{
							APIKey: &tcorev1.APIKeyCredentialsConfig{
								Path: apiKeyPath,
							},
						},
					},
				),
			},
			exp: func() *tclient.Options {
				tco := newTCOptions()
				tco.Credentials = apiKeyPathCreds
				return tco
			}(),
		},
		{
			name: "api key path with tls server name",
			args: []connection.NewLoaderOption{
				connection.WithTemporalClientLogger(noopLogger),
				connection.WithNewAPIKeyDynamicCredentialsFunc(
					newMockAPIKeyDynamicCredsFunc(apiKeyPath),
				),
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{
						TLS: tcorev1.ServerTLSConfig{
							ServerName: tlsServerName,
						},
						Credentials: tcorev1.CredentialsConfig{
							APIKey: &tcorev1.APIKeyCredentialsConfig{
								Path: apiKeyPath,
							},
						},
					},
				),
			},
			exp: func() *tclient.Options {
				tco := newTCOptions()
				tco.ConnectionOptions.TLS = &tls.Config{
					ServerName: tlsServerName,
				}
				tco.Credentials = apiKeyPathCreds
				return tco
			}(),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(tt *testing.T) {
			require := require.New(tt)
			loader := connection.NewLoader(tc.args...)
			got, err := loader.TemporalClientOptions()
			if tc.expErr != "" {
				require.ErrorContains(err, tc.expErr)
				return
			}
			require.Nil(err)
			require.NotNil(got)
			require.EqualExportedValues(tc.exp, got)
		})
	}
}

func TestLoader_ClientOptions_ControllerSpec_MutualTLSCredentials(t *testing.T) {
	mtlsSecretName := "mtls-secret-name"
	mtlsOpaqueSecretDataKey := "mtls"

	caKP, err := validCAKeyPair()
	require.Nil(t, err)

	poolWithValidCACert, err := x509.SystemCertPool()
	require.Nil(t, err)
	added := poolWithValidCACert.AppendCertsFromPEM(caKP.Cert)
	require.True(t, added)

	serverKP, err := validServerKeyPair()
	require.Nil(t, err)

	poolWithValidServerCACert, err := x509.SystemCertPool()
	require.Nil(t, err)
	added = poolWithValidServerCACert.AppendCertsFromPEM(serverKP.Cert)
	require.True(t, added)

	clientKP, err := validClientKeyPair()
	require.Nil(t, err)

	clientTLSCert, err := clientKP.TLSCertificate()
	require.Nil(t, err)

	clientPEMBundleBytes, err := tlsutil.CertToPEMBundle(*clientTLSCert)
	require.Nil(t, err)
	mtlsOpaqueSecretData := clientPEMBundleBytes

	mtlsOpaqueSecret := &kcorev1.Secret{}
	mtlsOpaqueSecret.Name = mtlsSecretName
	mtlsOpaqueSecret.Data = map[string][]byte{
		mtlsOpaqueSecretDataKey: []byte(mtlsOpaqueSecretData),
	}

	mtlsInvalidSecret := &kcorev1.Secret{}
	mtlsInvalidSecret.Name = mtlsSecretName
	mtlsInvalidSecret.Data = map[string][]byte{
		mtlsOpaqueSecretDataKey: []byte("invalid cert"),
	}

	mtlsTLSSecret := &kcorev1.Secret{}
	mtlsTLSSecret.Type = kcorev1.SecretTypeTLS
	mtlsTLSSecret.Name = mtlsSecretName
	mtlsTLSSecret.Data = map[string][]byte{
		"tls.key": clientKP.Key,
		"tls.crt": clientKP.Cert,
		"ca.crt":  caKP.Cert,
	}

	mtlsDir := t.TempDir()

	mtlsPathsKey := filepath.Join(mtlsDir, "key.pem")
	mtlsPathsCert := filepath.Join(mtlsDir, "cert.pem")
	mtlsPathsCA := filepath.Join(mtlsDir, "ca.pem")

	os.WriteFile(mtlsPathsKey, clientKP.Key, 0644)
	os.WriteFile(mtlsPathsCert, clientKP.Cert, 0644)
	os.WriteFile(mtlsPathsCA, caKP.Cert, 0644)

	cases := []struct {
		name   string
		args   []connection.NewLoaderOption
		exp    *tclient.Options
		expErr string
	}{
		{
			name: "fail if mtls key secret no kube client",
			args: []connection.NewLoaderOption{
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{
						Credentials: tcorev1.CredentialsConfig{
							MutualTLS: &tcorev1.MutualTLSCredentialsConfig{
								SecretRef: &kcorev1.SecretKeySelector{
									LocalObjectReference: kcorev1.LocalObjectReference{
										Name: mtlsSecretName,
									},
									Key: mtlsOpaqueSecretDataKey,
								},
							},
						},
					},
				),
			},
			expErr: "Loader requires KubeClient to load Mutual TLS material from Secret",
		},
		{
			name: "fail if mtls key secret no such secret",
			args: []connection.NewLoaderOption{
				connection.WithKubeClient(
					kcrfake.NewFakeClient(),
				),
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{
						Credentials: tcorev1.CredentialsConfig{
							MutualTLS: &tcorev1.MutualTLSCredentialsConfig{
								SecretRef: &kcorev1.SecretKeySelector{
									LocalObjectReference: kcorev1.LocalObjectReference{
										Name: mtlsSecretName,
									},
									Key: mtlsOpaqueSecretDataKey,
								},
							},
						},
					},
				),
			},
			expErr: "secrets \"mtls-secret-name\" not found",
		},
		{
			name: "mtls opaque secret type",
			args: []connection.NewLoaderOption{
				connection.WithTemporalClientLogger(noopLogger),
				connection.WithKubeClient(
					kcrfake.NewFakeClient(mtlsOpaqueSecret),
				),
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{
						Credentials: tcorev1.CredentialsConfig{
							MutualTLS: &tcorev1.MutualTLSCredentialsConfig{
								SecretRef: &kcorev1.SecretKeySelector{
									LocalObjectReference: kcorev1.LocalObjectReference{
										Name: mtlsSecretName,
									},
									Key: mtlsOpaqueSecretDataKey,
								},
							},
						},
					},
				),
			},
			exp: func() *tclient.Options {
				tco := newTCOptions()
				tco.ConnectionOptions.TLS = &tls.Config{
					Certificates: []tls.Certificate{*clientTLSCert},
					RootCAs:      poolWithValidServerCACert,
				}
				return tco
			}(),
		},
		{
			name: "mtls opaque secret with tls server name",
			args: []connection.NewLoaderOption{
				connection.WithTemporalClientLogger(noopLogger),
				connection.WithKubeClient(
					kcrfake.NewFakeClient(mtlsOpaqueSecret),
				),
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{
						TLS: tcorev1.ServerTLSConfig{
							ServerName: tlsServerName,
						},
						Credentials: tcorev1.CredentialsConfig{
							MutualTLS: &tcorev1.MutualTLSCredentialsConfig{
								SecretRef: &kcorev1.SecretKeySelector{
									LocalObjectReference: kcorev1.LocalObjectReference{
										Name: mtlsSecretName,
									},
									Key: mtlsOpaqueSecretDataKey,
								},
							},
						},
					},
				),
			},
			exp: func() *tclient.Options {
				tco := newTCOptions()
				tco.ConnectionOptions.TLS = &tls.Config{
					ServerName:   tlsServerName,
					Certificates: []tls.Certificate{*clientTLSCert},
					RootCAs:      poolWithValidServerCACert,
				}
				return tco
			}(),
		},
		{
			name: "mtls tls secret type",
			args: []connection.NewLoaderOption{
				connection.WithTemporalClientLogger(noopLogger),
				connection.WithKubeClient(
					kcrfake.NewFakeClient(mtlsTLSSecret),
				),
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{
						Credentials: tcorev1.CredentialsConfig{
							MutualTLS: &tcorev1.MutualTLSCredentialsConfig{
								SecretRef: &kcorev1.SecretKeySelector{
									LocalObjectReference: kcorev1.LocalObjectReference{
										Name: mtlsSecretName,
									},
									// NOTE(jaypipes): Deliberately leave Key
									// field empty since this Secret is of type
									// kubernetes.io/tls and we want to verify
									// the autoselecting of the fixed `ca.crt`
									// data key.
								},
							},
						},
					},
				),
			},
			exp: func() *tclient.Options {
				tco := newTCOptions()
				tco.ConnectionOptions.TLS = &tls.Config{
					Certificates: []tls.Certificate{*clientTLSCert},
					RootCAs:      poolWithValidCACert,
				}
				return tco
			}(),
		},
		{
			name: "mtls tls secret type with tls server name",
			args: []connection.NewLoaderOption{
				connection.WithTemporalClientLogger(noopLogger),
				connection.WithKubeClient(
					kcrfake.NewFakeClient(mtlsTLSSecret),
				),
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{
						TLS: tcorev1.ServerTLSConfig{
							ServerName: tlsServerName,
						},
						Credentials: tcorev1.CredentialsConfig{
							MutualTLS: &tcorev1.MutualTLSCredentialsConfig{
								SecretRef: &kcorev1.SecretKeySelector{
									LocalObjectReference: kcorev1.LocalObjectReference{
										Name: mtlsSecretName,
									},
									// NOTE(jaypipes): Deliberately leave Key
									// field empty since this Secret is of type
									// kubernetes.io/tls and we want to verify
									// the autoselecting of the fixed `ca.crt`
									// data key.
								},
							},
						},
					},
				),
			},
			exp: func() *tclient.Options {
				tco := newTCOptions()
				tco.ConnectionOptions.TLS = &tls.Config{
					ServerName:   tlsServerName,
					Certificates: []tls.Certificate{*clientTLSCert},
					RootCAs:      poolWithValidCACert,
				}
				return tco
			}(),
		},
		{
			name: "fail if mtls paths not found",
			args: []connection.NewLoaderOption{
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{
						Credentials: tcorev1.CredentialsConfig{
							MutualTLS: &tcorev1.MutualTLSCredentialsConfig{
								Paths: &tcorev1.MutualTLSPaths{
									Key: "path/does/not/exist",
								},
							},
						},
					},
				),
			},
			expErr: "path/does/not/exist: no such file or directory",
		},
		{
			name: "mtls paths",
			args: []connection.NewLoaderOption{
				connection.WithTemporalClientLogger(noopLogger),
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{
						Credentials: tcorev1.CredentialsConfig{
							MutualTLS: &tcorev1.MutualTLSCredentialsConfig{
								Paths: &tcorev1.MutualTLSPaths{
									Key:  mtlsPathsKey,
									Cert: mtlsPathsCert,
									CA:   mtlsPathsCA,
								},
							},
						},
					},
				),
			},
			exp: func() *tclient.Options {
				tco := newTCOptions()
				tco.ConnectionOptions.TLS = &tls.Config{
					Certificates: []tls.Certificate{*clientTLSCert},
					RootCAs:      poolWithValidCACert,
				}
				return tco
			}(),
		},
		{
			name: "mtls paths with tls server name",
			args: []connection.NewLoaderOption{
				connection.WithTemporalClientLogger(noopLogger),
				connection.WithConnectionSpec(
					tcorev1.ConnectionSpec{
						TLS: tcorev1.ServerTLSConfig{
							ServerName: tlsServerName,
						},
						Credentials: tcorev1.CredentialsConfig{
							MutualTLS: &tcorev1.MutualTLSCredentialsConfig{
								Paths: &tcorev1.MutualTLSPaths{
									Key:  mtlsPathsKey,
									Cert: mtlsPathsCert,
									CA:   mtlsPathsCA,
								},
							},
						},
					},
				),
			},
			exp: func() *tclient.Options {
				tco := newTCOptions()
				tco.ConnectionOptions.TLS = &tls.Config{
					ServerName:   tlsServerName,
					Certificates: []tls.Certificate{*clientTLSCert},
					RootCAs:      poolWithValidCACert,
				}
				return tco
			}(),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(tt *testing.T) {
			require := require.New(tt)
			loader := connection.NewLoader(tc.args...)
			got, err := loader.TemporalClientOptions()
			if tc.expErr != "" {
				require.ErrorContains(err, tc.expErr)
				return
			}
			require.Nil(err)
			require.NotNil(got)
			require.EqualExportedValues(tc.exp, got)
		})
	}
}
