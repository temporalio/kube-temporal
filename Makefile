# Top-level Makefile containing make targets that automate build, release,
# deployment and local testing tasks.

KUBECTL ?= kubectl
ROOT_DIR ?= $(shell pwd)
SCRIPTS_DIR ?= $(ROOT_DIR)/scripts

BIN_DIR ?= $(ROOT_DIR)/bin
$(BIN_DIR):
	@mkdir -p "$(BIN_DIR)"

GIT_VERSION ?= $(shell git describe --tags --always --dirty || echo "unknown")
GIT_COMMIT ?= $(shell git rev-parse HEAD)

include .mk/go.mk
include .mk/gen.mk
include .mk/test.mk
include .mk/help.mk
