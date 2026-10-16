package connection

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	tclient "go.temporal.io/sdk/client"
	tlog "go.temporal.io/sdk/log"
	kcorev1 "k8s.io/api/core/v1"
	ktypes "k8s.io/apimachinery/pkg/types"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"

	tcorev1 "github.com/temporalio/kube-temporal/api/core/v1"
	"github.com/temporalio/kube-temporal/pkg/tlsutil"
)

// NOTE(jaypipes): This is needed in order to unit test the Temporal SDK
// client.NewAPIKeyDynamicCredentials package-level function. Package-level
// functions cannot be mocked in Go.
type NewAPIKeyDynamicCredentialsFunc func(func(context.Context) (string, error)) tclient.Credentials

// Loader can load a Temporal sdkclient.Client from a variety of sources,
// including a ConnectionSpec.
type Loader struct {
	kc  kclient.Client
	kns string

	tcOpts     *tclient.Options
	tcDialFunc TemporalClientDialFunc
	tcLogger   tlog.Logger

	// really only useful for dependency injection for unit testing
	tcNewAPIKeyDynamicCredentialsFunc NewAPIKeyDynamicCredentialsFunc

	connSpec *tcorev1.ConnectionSpec
}

// KubeClient returns the Kubernetes controller-runtime Client associated with
// the Loader.
func (l Loader) KubeClient() kclient.Client {
	return l.kc
}

// KubeNamespace returns the Kubernetes Namespace associated with the
// Loader.
func (l Loader) KubeNamespace() string {
	return l.kns
}

// TemporalClientDialFunc returns the function callback that dials the Temporal
// API server and returns a Temporal SDK client.Client.
func (l *Loader) TemporalClientDialFunc() TemporalClientDialFunc {
	if l.tcDialFunc == nil {
		l.tcDialFunc = tclient.Dial
	}
	return l.tcDialFunc
}

// TemporalClientLogger returns the logger that is set in the client.Options.
func (l Loader) TemporalClientLogger() tlog.Logger {
	return l.tcLogger
}

// NewAPIKeyDynamicCredentialsFunc returns the function callback that returns a
// new Temporal sdkclient.Credentials constructed from a callback.
func (l *Loader) NewAPIKeyDynamicCredentialsFunc() NewAPIKeyDynamicCredentialsFunc {
	if l.tcNewAPIKeyDynamicCredentialsFunc == nil {
		l.tcNewAPIKeyDynamicCredentialsFunc = tclient.NewAPIKeyDynamicCredentials
	}
	return l.tcNewAPIKeyDynamicCredentialsFunc
}

// TemporalClientOptions returns the Temporal SDK client Options to use when
// dialing the Temporal client.
func (l *Loader) TemporalClientOptions() (*tclient.Options, error) {
	if l.tcOpts != nil {
		return l.tcOpts, nil
	}
	if l.connSpec == nil {
		return nil, fmt.Errorf("ConnectionSpec required")
	}
	return sync.OnceValues(func() (*tclient.Options, error) {
		tcOpts, err := l.clientOptionsFromConnectionSpec()
		if err != nil {
			return nil, err
		}
		l.tcOpts = tcOpts
		return tcOpts, nil
	})()
}

// NewLoaderOption allows modifying the Loader returned from NewLoader.
type NewLoaderOption func(*Loader)

// WithKubeClient allows the caller to supply a Kubernetes client to use if
// looking up credentials in Kubernetes Secrets or ConfigMaps.
func WithKubeClient(kc kclient.Client) NewLoaderOption {
	return func(l *Loader) {
		l.kc = kc
	}
}

// WithKubeNamespace allows the caller to supply a Kubernetes Namespace to use
// if looking up credentials in Kubernetes Secrets or ConfigMaps.
func WithKubeNamespace(ns string) NewLoaderOption {
	return func(l *Loader) {
		l.kns = ns
	}
}

// WithTemporalClientDialFunc allows the caller to supply a function to use
// when dialing a Temporal API server.
//
// Useful for overriding the dialer behaviour to return a fake client in
// testing. Defaults to Temporal sdkclient.Dial function.
func WithTemporalClientDialFunc(fn TemporalClientDialFunc) NewLoaderOption {
	return func(l *Loader) {
		l.tcDialFunc = fn
	}
}

// WithTemporalClientLogger allows the caller to supply a custom Temporal SDK
// Logger to set in the client.Options.
func WithTemporalClientLogger(logger tlog.Logger) NewLoaderOption {
	return func(l *Loader) {
		l.tcLogger = logger
	}
}

// WithTemporalClientOptions allows the caller to override the Options supplied
// to the Temporal Client dialer when constructing the Temporal Client.
func WithTemporalClientOptions(tcOpts tclient.Options) NewLoaderOption {
	return func(l *Loader) {
		l.tcOpts = &tcOpts
	}
}

// WithNewAPIKeyDynamicCredentialsFunc allows the caller to supply a function
// to use in place of the Temporal sdkclient.NewAPIKeyDynamicCredentials
// package-level function.
func WithNewAPIKeyDynamicCredentialsFunc(fn NewAPIKeyDynamicCredentialsFunc) NewLoaderOption {
	return func(l *Loader) {
		l.tcNewAPIKeyDynamicCredentialsFunc = fn
	}
}

// WithConnectionSpec sets the Loader's ConnectionSpec which informs the Loader
// where to find various credentials and connectivity information for the
// Temporal sdkclient.Client.
func WithConnectionSpec(spec tcorev1.ConnectionSpec) NewLoaderOption {
	return func(l *Loader) {
		l.connSpec = &spec
	}
}

// NewLoader returns a Loader that knows how to load a Connection from a
// variety of sources, including a ConnectionSpec.
func NewLoader(opts ...NewLoaderOption) Loader {
	l := Loader{}
	for _, opt := range opts {
		opt(&l)
	}

	return l
}

// clientOptionsFromConnectionSpec returns Temporal sdkclient.Options
// constructed from the Loader's ConnectionSpec.
func (l *Loader) clientOptionsFromConnectionSpec() (*tclient.Options, error) {
	spec := l.connSpec
	if spec == nil {
		return nil, fmt.Errorf("ConnectionSpec required")
	}
	tcOpts := tclient.Options{
		Logger:    l.tcLogger,
		HostPort:  spec.HostPort,
		Namespace: spec.Namespace,
		Identity:  spec.Identity,
	}

	if spec.Credentials.APIKey != nil {
		tcCreds, err := l.credsFromAPIKeyCredentialsConfig(
			*spec.Credentials.APIKey,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"failed getting Temporal client.Credentials from "+
					"APIKeyCredentialsConfig: %w",
				err,
			)
		}
		tcOpts.Credentials = tcCreds
	}

	var err error
	var tlsCfg *tls.Config

	if spec.Credentials.MutualTLS != nil {
		// For mTLS authentication, the ConnectionOptions.TLS configuration
		// *is* the credentials configuration. We don't set any additional CA
		// cert or system pool when using mTLS authentication.
		//
		// NOTE(jaypipes): We don't use the sdkclient.NewMTLSCredentials
		// function because it does not allow us to override the root CAs or
		// set the TLS ServerName. There is no actual MTLSCredentials struct in
		// the Temporal sdkclient at all. Instead, the NewMTLSCredentials
		// merely adapts the sdkclient.Options.TLS struct.
		tlsCfg, err = l.tlsConfigFromMutualTLSCredentialsConfig(
			*spec.Credentials.MutualTLS,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"failed getting TLS Config from "+
					"MutualTLSCredentialsConfig: %w",
				err,
			)
		}
	} else {
		// For API Key or no authentication, the ConnectionOptions.TLS is
		// configured by examining the ConnectionSpec.TLS field, which
		// represents the TLS configuration for connecting to the Temporal API
		// Server.
		tlsCfg, err = l.tlsConfigFromServerTLSConfig(spec.TLS)
		if err != nil {
			return nil, fmt.Errorf(
				"failed getting TLS Config from ServerTLSConfig: %w", err,
			)
		}
	}

	// If the TLS ServerName has been set, make sure we set it in the
	// tls.Config struct in ConnectionOptions.TLS.
	if tlsCfg == nil {
		if spec.TLS.ServerName != "" {
			tlsCfg = &tls.Config{}
			tlsCfg.ServerName = spec.TLS.ServerName
		}
	} else {
		if spec.TLS.ServerName != "" {
			tlsCfg.ServerName = spec.TLS.ServerName
		}
	}
	tcOpts.ConnectionOptions.TLS = tlsCfg
	return &tcOpts, nil
}

// credsFromAPIKeyCredentialsConfig returns the Temporal client.Credentials
// constructed from the supplied APIKeyCredentialsConfig.
func (l Loader) credsFromAPIKeyCredentialsConfig(
	cfg tcorev1.APIKeyCredentialsConfig,
) (tclient.Credentials, error) {
	switch {
	case cfg.SecretRef != nil:
		return l.credsFromAPIKeyCredentialsSecretRef(*cfg.SecretRef)
	case cfg.Path != "":
		return l.credsFromAPIKeyCredentialsPath(cfg.Path)
	default:
		return nil, fmt.Errorf("either SecretRef or Path must be non-empty")
	}
}

// credsFromAPIKeyCredentialsSecretRef returns the Temporal client.Credentials
// constructed from the supplied kcorev1.SecretKeySelector.
func (l Loader) credsFromAPIKeyCredentialsSecretRef(
	secretRef kcorev1.SecretKeySelector,
) (tclient.Credentials, error) {
	kc := l.KubeClient()
	if kc == nil {
		return nil, fmt.Errorf(
			"Loader requires KubeClient to load API Key from Secret",
		)
	}
	secretName := secretRef.Name
	if secretName == "" {
		return nil, fmt.Errorf("APIKey.SecretRef.Name cannot be empty")
	}
	secretKey := secretRef.Key
	if secretKey == "" {
		return nil, fmt.Errorf("APIKey.SecretRef.Key cannot be empty")
	}

	ctx := context.TODO()
	apiKeyGetter := func(ctx context.Context) (string, error) {
		lookup := ktypes.NamespacedName{
			Name:      secretName,
			Namespace: l.KubeNamespace(),
		}
		var secret kcorev1.Secret
		err := kc.Get(ctx, lookup, &secret)
		if err != nil {
			return "", fmt.Errorf(
				"failed reading CA secret %q: %w", secretName, err,
			)
		}
		apiKeyContents, ok := secret.Data[secretKey]
		if !ok || len(apiKeyContents) == 0 {
			return "", fmt.Errorf(
				"secret %q referenced by apiKey.secretRef "+
					"has no entry at key %q",
				secretName, secretKey,
			)
		}
		return string(apiKeyContents), nil
	}
	// Initial check that the secret actually exists. We load the API key
	// with dynamic credentials to allow for adapting to rotated tokens...
	_, err := apiKeyGetter(ctx)
	if err != nil {
		return nil, err
	}
	credFn := l.NewAPIKeyDynamicCredentialsFunc()
	return credFn(apiKeyGetter), nil
}

// credsFromAPIKeyCredentialsPath returns the Temporal client.Credentials
// constructed from the supplied file path.
func (l Loader) credsFromAPIKeyCredentialsPath(
	path string,
) (tclient.Credentials, error) {
	// Initial check that the file at Path actually exists. We load the API
	// key with dynamic credentials to allow for adapting to rotated
	// tokens...
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	f.Close()
	credFn := l.NewAPIKeyDynamicCredentialsFunc()
	return credFn(
		func(ctx context.Context) (string, error) {
			contents, err := os.ReadFile(path)
			if err != nil {
				return "", err
			}
			return strings.TrimSpace(string(contents)), nil
		},
	), nil
}

// tlsConfigFromMutualTLSCredentialsConfig returns the tls.Config constructed
// from the supplied MutualTLSCredentialsConfig.
func (l Loader) tlsConfigFromMutualTLSCredentialsConfig(
	cfg tcorev1.MutualTLSCredentialsConfig,
) (*tls.Config, error) {
	var cert *tls.Certificate
	var caContents []byte
	var err error
	switch {
	case cfg.SecretRef != nil:
		cert, caContents, err = l.tlsCertFromMutualTLSCredentialsSecretRef(
			*cfg.SecretRef,
		)
	case cfg.Paths != nil:
		cert, caContents, err = l.tlsCertFromMutualTLSCredentialsPaths(
			*cfg.Paths,
		)
	default:
		return nil, fmt.Errorf("either SecretRef or Paths must be non-empty")
	}
	if err != nil {
		return nil, err
	}

	exp, err := tlsutil.CertExpirationTime(cert, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf(
			"failed checking certificate expiration: %w", err,
		)
	}
	if time.Now().After(exp) {
		return nil, fmt.Errorf(
			"certificate is expired or is going to expire soon",
		)
	}
	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{*cert},
	}
	// Append the secret's CA to the system pool so privately-signed servers
	// are trusted alongside public CAs. When ca.crt is absent, RootCAs stays
	// unset and Go uses the system pool.
	if len(caContents) > 0 {
		rootCAs, err := tlsutil.SystemCAPoolWithCert(caContents)
		if err != nil {
			return nil, err
		}
		tlsCfg.RootCAs = rootCAs
	}
	return tlsCfg, nil
}

// tlsCertFromMutualTLSCredentialsSecretRef returns the tls.Certificate and
// optional CA cert contents constructed after reading the supplied
// kcorev1.SecretKeySelector.
func (l Loader) tlsCertFromMutualTLSCredentialsSecretRef(
	secretRef kcorev1.SecretKeySelector,
) (*tls.Certificate, []byte, error) {
	secretName := secretRef.Name
	if secretName == "" {
		return nil, nil, fmt.Errorf("MutualTLS.SecretRef.Name cannot be empty")
	}
	secretKey := secretRef.Key

	ctx := context.TODO()
	kc := l.KubeClient()
	if kc == nil {
		return nil, nil, fmt.Errorf(
			"Loader requires KubeClient to load Mutual " +
				"TLS material from Secret",
		)
	}
	lookup := ktypes.NamespacedName{
		Name:      secretName,
		Namespace: l.KubeNamespace(),
	}
	var secret kcorev1.Secret
	err := kc.Get(ctx, lookup, &secret)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed reading CA secret %q: %w", secretName, err,
		)
	}
	secretType := secret.Type
	switch secretType {
	case kcorev1.SecretTypeOpaque, "":
		if secretKey == "" {
			return nil, nil, fmt.Errorf(
				"MutualTLS.SecretRef.Key cannot be empty when " +
					"Secret type is opaque",
			)
		}
		mtlsContents, ok := secret.Data[secretKey]
		if !ok || len(mtlsContents) == 0 {
			return nil, nil, fmt.Errorf(
				"secret %q referenced by mtls.secretRef "+
					"has no entry at key %q",
				secretName, secretKey,
			)
		}
		// The value is expected to be a PEM-encoded bytestring containing
		// the client key, client cert, and optional CA cert.
		cert, serverCAContents, err := tlsutil.CertFromPEMBundle(mtlsContents)
		if err != nil {
			return nil, nil, fmt.Errorf(
				"failed parsing composite PEM file: %w", err,
			)
		}
		return cert, serverCAContents, nil
	case kcorev1.SecretTypeTLS:
		certContents := secret.Data["tls.crt"]
		keyContents := secret.Data["tls.key"]
		caContents := secret.Data["ca.crt"]
		cert, err := tls.X509KeyPair(certContents, keyContents)
		if err != nil {
			return nil, nil, fmt.Errorf(
				"failed constructing x509 key pair for "+
					"mTLS authentication: %w",
				err,
			)
		}
		return &cert, caContents, nil
	default:
		return nil, nil, fmt.Errorf(
			"unhandled secret type %q for MutualTLS.SecretRef", secretType,
		)
	}
}

// tlsCertFromMutualTLSCredentialsPaths returns the tls.Certificate and
// optional CA cert contents constructed after reading the supplied
// tcorev1.MutualTLSPaths.
func (l Loader) tlsCertFromMutualTLSCredentialsPaths(
	paths tcorev1.MutualTLSPaths,
) (*tls.Certificate, []byte, error) {
	keyPath := paths.Key
	keyContents, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to read MTLS private key from %q: %w", keyPath, err,
		)
	}
	certPath := paths.Cert
	certContents, err := os.ReadFile(certPath)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to read MTLS private cert from %q: %w", certPath, err,
		)
	}

	var caContents []byte
	caPath := paths.CA
	if caPath != "" {
		caContents, err = os.ReadFile(caPath)
		if err != nil {
			return nil, nil, fmt.Errorf(
				"failed to read MTLS CA bundle from %q: %w", caPath, err,
			)
		}
	}
	cert, err := tls.X509KeyPair(certContents, keyContents)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed constructing x509 key pair for "+
				"mTLS authentication: %w",
			err,
		)
	}
	return &cert, caContents, nil
}

// tlsConfigFromServerTLSConfig returns the tls.Config constructed after
// examining the supplied tcorev1.ServerTLSConfig.
func (l Loader) tlsConfigFromServerTLSConfig(
	serverTLSCfg tcorev1.ServerTLSConfig,
) (
	*tls.Config,
	error,
) {
	spec := l.connSpec
	if spec.TLS.CACert == nil {
		return nil, nil
	}
	caCert, err := l.caCertFromCACertConfig(*spec.TLS.CACert)
	if err != nil {
		return nil, err
	}
	rootCAs, err := tlsutil.SystemCAPoolWithCert(caCert)
	if err != nil {
		return nil, err
	}
	if rootCAs == nil {
		// nothing to override, just let the system defaults be used.
		return nil, nil
	}
	return &tls.Config{
		RootCAs: rootCAs,
	}, nil
}

// caCertFromConnectionSpec returns the CA cert after examining the supplied
// tcorev1.CACertConfig.
func (l Loader) caCertFromCACertConfig(
	caCertCfg tcorev1.CACertConfig,
) ([]byte, error) {
	switch {
	case caCertCfg.SecretRef != nil:
		secretName := caCertCfg.SecretRef.Name
		if secretName == "" {
			return nil, fmt.Errorf("SecretRef.Name cannot be empty")
		}
		kc := l.KubeClient()
		if kc == nil {
			return nil, fmt.Errorf(
				"Loader requires KubeClient to load CA cert from Secret",
			)
		}
		ctx := context.TODO()
		lookup := ktypes.NamespacedName{
			Name:      secretName,
			Namespace: l.KubeNamespace(),
		}
		var secret kcorev1.Secret
		err := kc.Get(ctx, lookup, &secret)
		if err != nil {
			return nil, fmt.Errorf(
				"failed reading CA secret %q: %w", secretName, err,
			)
		}
		var secretKey string
		secretType := secret.Type
		switch secretType {
		case kcorev1.SecretTypeTLS:
			secretKey = "ca.crt"
		case kcorev1.SecretTypeOpaque, "":
			secretKey = caCertCfg.SecretRef.Key
			if secretKey == "" {
				return nil, fmt.Errorf(
					"SecretRef.Key cannot be empty for "+
						"Secret %q with opaque secret type",
					secretName,
				)
			}
		default:
			return nil, fmt.Errorf(
				"unsupported secret type %q for CA certificate Secret %q",
				secretType, secretName,
			)
		}
		caCert, ok := secret.Data[secretKey]
		if !ok || len(caCert) == 0 {
			return nil, fmt.Errorf(
				"CA secret %q referenced by tls.caCert.secretRef "+
					"has no entry at key %q",
				secretName, secretKey,
			)
		}
		return caCert, nil
	case caCertCfg.Path != "":
		caCert, err := os.ReadFile(caCertCfg.Path)
		if err != nil {
			return nil, fmt.Errorf(
				"failed reading CA cert from path: %w", err,
			)
		}
		return caCert, nil
	default:
		return nil, fmt.Errorf(
			"either CACert.SecretRef or CACert.Path must be set",
		)
	}
}
