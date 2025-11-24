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
		cmd/nomad-pipeline/cmd.go
	@echo "==> Done"

bin/%/nomad-pipeline-runner: GO_OUT ?= $@
bin/%/nomad-pipeline-runner: ## Build Nomad Pipeline Runner for GOOS & GOARCH; eg. bin/linux_amd64/nomad-pipeline-runner
	@echo "==> Building $@..."
	@GOOS=$(firstword $(subst _, ,$*)) \
		GOARCH=$(lastword $(subst _, ,$*)) \
		go build \
		-o $(GO_OUT) \
		-trimpath \
		-ldflags "$(GO_LDFLAGS)" \
		cmd/nomad-pipeline-runner/cmd.go
	@echo "==> Done"

.PHONY: build
build: ## Build a development version of Nomad Pipeline
	@echo "==> Building Nomad Pipeline..."
	@go build \
		-o bin/nomad-pipeline \
		-trimpath \
		-ldflags "$(GO_LDFLAGS)" \
		cmd/nomad-pipeline/cmd.go
	@echo "==> Done"
	@echo "==> Building Nomad Pipeline Runner..."
	@go build \
		-o bin/nomad-pipeline-runner \
		-trimpath \
		-ldflags "$(GO_LDFLAGS)" \
		cmd/nomad-pipeline-runner/cmd.go
	@echo "==> Done"

.PHONY: clean
clean: ## Clean built binaries
	@echo "==> Cleaning up..."
	@rm -rf bin/
	@echo "==> Done"

.PHONY: build-docker-all
build-docker-all: ## Build all Docker images for all supported platforms
	@$(MAKE) clean
	@$(MAKE) bin/linux_amd64/nomad-pipeline
	@$(MAKE) bin/linux_arm64/nomad-pipeline
	@$(MAKE) bin/linux_amd64/nomad-pipeline-runner
	@$(MAKE) bin/linux_arm64/nomad-pipeline-runner
	@echo "==> Building all Docker images..."
	@docker buildx build --platform linux/amd64,linux/arm64 -f build/controller/Dockerfile .
	@docker buildx build --platform linux/amd64,linux/arm64 -f build/runner/Dockerfile .
	@echo "==> Done"

HELP_FORMAT="    \033[36m%-27s\033[0m %s\n"
.PHONY: help
help: ## Display this usage information
	@echo "Valid targets:"
	@grep -E '^[^ ]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		sort | \
		awk 'BEGIN {FS = ":.*?## "}; \
			{printf $(HELP_FORMAT), $$1, $$2}'
	@echo ""
