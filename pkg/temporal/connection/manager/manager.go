package manager

import (
	"fmt"
	"sync"

	"github.com/samber/lo"
	tcorev1 "github.com/temporalio/kube-temporal/api/core/v1"
	tclient "go.temporal.io/sdk/client"

	"github.com/temporalio/kube-temporal/pkg/temporal/connection"
	"github.com/temporalio/kube-temporal/pkg/temporal/connection/loader"
)

// NewOption adapts the Manager returned from New.
type NewOption func(*Manager)

// WithLoaderOptions sets the Manager's set of loader.NewOptions used when the
// Manager lazily creates Temporal SDK clients.
func WithLoaderOptions(opts ...loader.NewOption) NewOption {
	return func(m *Manager) {
		m.loaderOpts = opts
	}
}

// New returns a new Temporal connection manager.
func New(opts ...NewOption) *Manager {
	m := Manager{
		hostPortNamespaceConnSpecs: map[connection.HostPortNamespace][]*tcorev1.ConnectionSpec{},
		hostPortNamespaceClients:   map[connection.HostPortNamespace]tclient.Client{},
		connSpecs:                  map[string]*tcorev1.ConnectionSpec{},
		clients:                    map[string]tclient.Client{},
	}
	for _, opt := range opts {
		opt(&m)
	}
	return &m
}

// Manager manages a cache of Temporal connection-related objects.
//
// Temporal SDK client.Client objects must specify both a HostPort and a
// Temporal Namespace. This is unlike client connections to Temporal Cloud Ops
// API which must only specify a HostPort and a single client connection may
// perform operations across multiple Temporal Namespaces. Because of this
// requirement, the caches for Temporal SDK clients managed by this Manager are
// maps keyed by HostPortNamespace structs.
type Manager struct {
	sync.RWMutex
	// loaderOpts contains any loader.NewOptions the Manager uses when
	// lazily creating Temporal SDK clients.
	loaderOpts []loader.NewOption
	// hostPortNamespaceConnSpecs contains a map, keyed by HostPortNamespace,
	// of slices of [tcorev1.ConnectionSpec] objects that contain instructions
	// for creating clients to that HostPort and Namespace combination.
	hostPortNamespaceConnSpecs map[connection.HostPortNamespace][]*tcorev1.ConnectionSpec
	// hostPortNamespaceClients contains a map, keyed by HostPortNamespace, of
	// [tclient.Client] objects associated with that HostPort and Namespace
	// combination.
	hostPortNamespaceClients map[connection.HostPortNamespace]tclient.Client
	// connSpecs contains the set of [tcorev1.ConnectionSpec] objects
	// associated with the hash of a [tcorev1.ConnectionSpec].
	connSpecs map[string]*tcorev1.ConnectionSpec
	// clients contains the set of [tclient.Client] objects associated with
	// the hash of a [tcorev1.ConnectionSpec].
	clients map[string]tclient.Client
}

// SetLoaderOptions sets the Manager's set of loader.NewOptions the Manager
// uses when lazily creating Temporal SDK clients.
func (m *Manager) SetLoaderOptions(opts ...loader.NewOption) {
	m.loaderOpts = opts
}

type registerOptions struct {
	tc        tclient.Client
	connSpecs []*tcorev1.ConnectionSpec
}

// RegisterOption allows callers to the Manager.Register function to control
// what is being registered with the manager.
type RegisterOption func(*registerOptions)

// RegisterWithClient allows overriding the Temporal client associated
// with the Register subject.
func RegisterWithClient(tc tclient.Client) RegisterOption {
	return func(o *registerOptions) {
		o.tc = tc
	}
}

// RegisterWithConnectionSpecs allows callers to the Manager.Register function
// to associate one or more [tcorev1.ConnectionSpec] objects with the
// HostPort being registered.
func RegisterWithConnectionSpecs(
	specs ...*tcorev1.ConnectionSpec,
) RegisterOption {
	return func(o *registerOptions) {
		o.connSpecs = append(o.connSpecs, specs...)
	}
}

// Register safely adds the supplied subject to the Manager's various caches.
//
// If subject is a [tcorev1.ConnectionSpec], the Manager maps the
// spec's host port and namespace to the connection spec.
//
// If subject is a HostPort or HostPortNamespace at least one RegisterOption
// should be provided that associates the HostPort or HostPortNamespace to a
// Temporal client or a set of related connection specs.
func (m *Manager) Register(subject any, opts ...RegisterOption) error {
	ro := registerOptions{}
	for _, opt := range opts {
		opt(&ro)
	}

	// Only allow a single caller to Register at a time.
	m.Lock()
	defer m.Unlock()

	var key connection.HostPortNamespace

	switch subject := subject.(type) {
	case *tcorev1.ConnectionSpec:
		specHash := string(subject.Hash())
		_, ok := m.connSpecs[specHash]

		if ok {
			// We already know about this subject
			return nil
		}

		key = connection.HostPortNamespace{
			HostPort:  connection.HostPort(subject.HostPort),
			Namespace: subject.Namespace,
		}
		key.SetDefaults()

		m.registerHostPortNamespaceConnSpecs(key, ro, subject)
		return nil
	case connection.HostPort:
		key = connection.HostPortNamespace{
			HostPort: subject,
		}
	case connection.HostPortNamespace:
		key = subject
	default:
		return fmt.Errorf("unsupported type %T", subject)
	}

	key.SetDefaults()

	err := m.validateRegisterHostPortNamespace(key, ro)
	if err != nil {
		return err
	}

	_, ok := m.hostPortNamespaceConnSpecs[key]
	if ok {
		// We already know about this subject
		return nil
	}

	// The HostPort may have been previously registered with no connection
	// specs and just a pre-created Temporal Client.
	_, ok = m.hostPortNamespaceClients[key]
	if ok {
		// We already know about this subject
		return nil
	}
	if len(ro.connSpecs) == 0 {
		m.hostPortNamespaceClients[key] = ro.tc
		return nil
	}
	m.registerHostPortNamespaceConnSpecs(key, ro, ro.connSpecs...)
	return nil
}

// validateRegisterHostPortNamespace returns an error if the supplied
// registerOptions are not valid for the supplied HostPortNamespace being
// registered.
func (m *Manager) validateRegisterHostPortNamespace(
	subject connection.HostPortNamespace,
	ro registerOptions,
) error {

	if ro.tc == nil && len(ro.connSpecs) == 0 {
		return fmt.Errorf(
			"connection specs or temporal client " +
				"required to register HostPort or HostPortNamespace",
		)
	}

	// Ensure that all the supplied connection specs actually use the
	// subject HostPort amd Namespace specified in register options.
	for _, spec := range ro.connSpecs {
		specHPNS := connection.HostPortNamespace{
			HostPort:  connection.HostPort(spec.HostPort),
			Namespace: spec.Namespace,
		}
		specHPNS.SetDefaults()
		if specHPNS.HostPort != subject.HostPort {
			return fmt.Errorf(
				"connection spec HostPort %q does not match subject HostPort %q",
				specHPNS.HostPort, subject.HostPort,
			)
		}
		if specHPNS.Namespace != subject.Namespace {
			return fmt.Errorf(
				"connection spec Namespace %q does not match subject "+
					"Namespace %q",
				specHPNS.Namespace, subject.Namespace,
			)
		}
	}
	return nil
}

// registerHostPortNamespaceConnSpecs registers the supplied
// tcorev1.ConnectionSpecs for the supplied HostPortNamespace.
//
// Always called with write lock acquired.
func (m *Manager) registerHostPortNamespaceConnSpecs(
	subject connection.HostPortNamespace,
	ro registerOptions,
	specs ...*tcorev1.ConnectionSpec,
) {
	for _, spec := range specs {
		specHash := string(spec.Hash())

		m.connSpecs[specHash] = spec
		if ro.tc != nil {
			m.clients[specHash] = ro.tc
		}

		key := connection.HostPortNamespace{
			HostPort:  connection.HostPort(spec.HostPort),
			Namespace: spec.Namespace,
		}
		if key.HostPort == "" {
			key.HostPort = connection.HostPort(
				tcorev1.DefaultTemporalConnectionHostPort,
			)
		}
		if key.Namespace == "" {
			key.Namespace = tcorev1.DefaultTemporalNamespace
		}

		keySpecs := m.hostPortNamespaceConnSpecs[key]
		if len(keySpecs) == 0 {
			keySpecs = []*tcorev1.ConnectionSpec{}
		}
		keySpecs = append(keySpecs, spec)
		m.hostPortNamespaceConnSpecs[key] = keySpecs

		if ro.tc != nil {
			_, ok := m.hostPortNamespaceClients[key]
			if !ok {
				m.hostPortNamespaceClients[key] = ro.tc
			}
		}
	}
}

// Client returns a [tclient.Client]-implementing struct configured for the
// matching the supplied subject.
//
// If a cached client is not found for the supplied subject, we attempt to
// create the client and cache the created client in the Manager.
//
// If subject is a [tcorev1.ConnectionSpec], we look up the client by the
// connection spec's hash key.
//
// If subject is a HostPort or HostPortNamespace, we find the first known
// connection spec having that HostPort and Namespace combination and then look
// up the client by that connection spec's hash key.
func (m *Manager) Client(subject any) (tclient.Client, error) {
	var specHash string
	var spec *tcorev1.ConnectionSpec

	switch subject := subject.(type) {
	case *tcorev1.ConnectionSpec:
		// Caller passed a ConnectionSpec. If we have a pre-created Temporal
		// client for this connection spec, just return it. Otherwise, we
		// fall through and attempt to create client from the supplied
		// connection spec that contains instructions to create a client for
		// this HostPort.
		if lo.IsNil(subject) {
			return nil, fmt.Errorf("nil connection spec")
		}

		specHash = string(subject.Hash())

		m.RLock()
		_, ok := m.connSpecs[specHash]
		m.RUnlock()
		if !ok {
			return nil, fmt.Errorf(
				"no registered connection spec matching hash",
			)
		}
		spec = subject
	case connection.HostPort:
		key := connection.HostPortNamespace{
			HostPort:  subject,
			Namespace: tcorev1.DefaultTemporalNamespace,
		}

		m.RLock()
		cli, ok := m.hostPortNamespaceClients[key]
		m.RUnlock()
		if ok {
			return cli, nil
		}

		// We don't have a cached Temporal for this HostPort and Namespace, so
		// get the first ConnectionSpec associated with the supplied HostPort
		// and Namespace.
		m.RLock()
		specs, ok := m.hostPortNamespaceConnSpecs[key]
		m.RUnlock()
		if !ok || len(specs) == 0 {
			return nil, fmt.Errorf(
				"no client or connection specs found for "+
					"HostPort %q and Namespace %q",
				key.HostPort, key.Namespace,
			)
		}
		spec = specs[0]
		specHash = string(spec.Hash())
	case connection.HostPortNamespace:
		m.RLock()
		cli, ok := m.hostPortNamespaceClients[subject]
		m.RUnlock()
		if ok {
			return cli, nil
		}

		// We don't have a cached Temporal for this HostPort and Namespace, so
		// get the first ConnectionSpec associated with the supplied HostPort
		// and Namespace.
		m.RLock()
		specs, ok := m.hostPortNamespaceConnSpecs[subject]
		m.RUnlock()
		if !ok || len(specs) == 0 {
			return nil, fmt.Errorf(
				"no client or connection specs found for "+
					"HostPort %q and Namespace %q",
				subject.HostPort, subject.Namespace,
			)
		}
		spec = specs[0]
		specHash = string(spec.Hash())
	default:
		return nil, fmt.Errorf(
			"unsupported type %T", subject,
		)

	}

	m.RLock()
	cli, ok := m.clients[specHash]
	m.RUnlock()
	if ok {
		return cli, nil
	}

	// lazily create the Temporal client.

	m.Lock()
	defer m.Unlock()

	lopts := m.loaderOpts
	lopts = append(lopts, loader.WithConnectionSpec(*spec))

	conn, err := connection.New(
		connection.WithLoader(
			loader.New(lopts...),
		),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed creating connection: %w", err,
		)
	}
	tc, err := conn.Client()
	if err != nil {
		return nil, fmt.Errorf(
			"failed getting client: %w", err,
		)
	}
	m.clients[specHash] = tc
	return tc, nil
}
