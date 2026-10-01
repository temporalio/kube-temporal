This repository contains Custom Resource Definitions (CRDs) and common Go
packages for Kubernetes controllers that manage Temporal resources.

# Contributing

## Developing

### Generating code

There are several Makefile targets to assist you in generating code.

```
make gen-all
```

Will generate all Go code and Kubernetes deployment manifests.

## Testing

This repository does *not* use Ginkgo or the upstream Kubernetes
controller-runtime `envtest` package for testing.

### Unit testing

For unit testing, we use simple, [table-driven][table-driven-testing] tests and
the following test libraries:

* `github.com/stretchr/testify/`
* `github.com/stretchr/testify/mock`

Unit tests should validate the smallest testable unit of work possible. If your
unit tests cross lots of function/method boundaries, you might not be writing a
unit test but instead building a larger functional or integration test.

To run unit tests:

```
make test-unit
```

[table-driven-testing]: https://dave.cheney.net/2019/05/07/prefer-table-driven-tests 
