GO ?= go

.PHONY: all
all: hooks tools ## Wire up git hooks and install local tooling

.PHONY: hooks
hooks: ## Point git at .githooks so commits are linted and tested
	git config core.hooksPath .githooks

.PHONY: tools
tools: ## Install commitlint locally (used by the commit-msg hook)
	npm install --no-save @commitlint/cli @commitlint/config-conventional

.PHONY: test
test: ## Run the test suite
	$(GO) test -race -shuffle=on ./...

.PHONY: lint
lint: ## Run every linter CI runs
	golangci-lint run
	yamllint .

.PHONY: help
help: ## Show this help
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'
