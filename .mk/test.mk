# ==== Test targets ====

##@ Test

.PHONY: test
test: test-unit ## Run all tests.

.PHONY: test-unit
test-unit: ## Run all unit tests.
	@go test -v ./... -coverprofile=cover.out
