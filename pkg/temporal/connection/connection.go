package connection

import (
	"fmt"

	tclient "go.temporal.io/sdk/client"
)

type TemporalClientDialFunc func(tclient.Options) (tclient.Client, error)

// Connection wraps a Temporal client struct with an ability to load the client
// from various sources, including a Kubernetes Secret object or local files.
type Connection struct {
	tc     tclient.Client
	loader *Loader
}

// Client returns the Temporal client associated with the Connection.
func (c *Connection) Client() (tclient.Client, error) {
	if c.tc != nil {
		return c.tc, nil
	}
	if c.loader == nil {
		return nil, fmt.Errorf("no Loader configured for Connection")
	}

	dial := c.loader.TemporalClientDialFunc()
	clientOpts, err := c.loader.TemporalClientOptions()
	if err != nil {
		return nil, err
	}
	client, err := dial(*clientOpts)
	if err != nil {
		return nil, err
	}
	c.tc = client
	return client, nil
}

// Loader returns the Loader associated with the Connection, if any.
func (c *Connection) Loader() *Loader {
	return c.loader
}

type newOptions struct {
	loader *Loader
}

// NewOption allows modifying the returned Connection from New.
type NewOption func(*Connection)

// WithLoader modifies the Connection to use the supplied Loader.
func WithLoader(loader Loader) NewOption {
	return func(c *Connection) {
		c.loader = &loader
	}
}

// WithTemporalClient overrides the Connection's Temporal Client with the
// supplied object.
func WithTemporalClient(tc tclient.Client) NewOption {
	return func(c *Connection) {
		c.tc = tc
	}
}

// New returns a new Connection.
func New(opts ...NewOption) (*Connection, error) {
	c := Connection{}
	for _, opt := range opts {
		opt(&c)
	}

	if (c.tc == nil && c.loader == nil) || (c.tc != nil && c.loader != nil) {
		return nil, fmt.Errorf("either use WithLoader or WithTemporalClient")
	}

	return &c, nil
}
