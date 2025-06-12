BUILD_COMMIT := $(shell git rev-parse HEAD)
BUILD_DIRTY := $(if $(shell git status --porcelain),+CHANGES)
BUILD_COMMIT_FLAG := github.com/hashicorp-forge/nomad-pipeline/internal/version.BuildCommit=$(BUILD_COMMIT)$(BUILD_DIRTY)

BUILD_TIME ?= $(shell TZ=UTC0 git show -s --format=%cd --date=format-local:'%Y-%m-%dT%H:%M:%SZ' HEAD)
BUILD_TIME_FLAG := github.com/hashicorp-forge/nomad-pipeline/internal/version.BuildTime=$(BUILD_TIME)

# Populate the ldflags using the Git commit information and and build time
# which will be present in the binary version output.
GO_LDFLAGS = -X $(BUILD_COMMIT_FLAG) -X $(BUILD_TIME_FLAG)

bin/%/nomad-pipeline: GO_OUT ?= $@
bin/%/nomad-pipeline: ## Build Nomad Pipeline for GOOS & GOARCH; eg. bin/linux_amd64/nomad-pipeline
	@echo "==> Building $@..."
	@GOOS=$(firstword $(subst _, ,$*)) \
		GOARCH=$(lastword $(subst _, ,$*)) \
		go build \
		-o $(GO_OUT) \
		-trimpath \
		-ldflags "$(GO_LDFLAGS)" \
		internal/cli/cli.go
	@echo "==> Done"

.PHONY: build
build: ## Build a development version of Nomad Pipeline
	@echo "==> Building Nomad Pipeline..."
	@go build \
		-o ./bin/nomad-pipeline \
		-trimpath \
		-ldflags "$(GO_LDFLAGS)" \
		internal/cli/cli.go
	@echo "==> Done"

HELP_FORMAT="    \033[36m%-25s\033[0m %s\n"
.PHONY: help
help: ## Display this usage information
	@echo "Valid targets:"
	@grep -E '^[^ ]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		sort | \
		awk 'BEGIN {FS = ":.*?## "}; \
			{printf $(HELP_FORMAT), $$1, $$2}'
	@echo ""
