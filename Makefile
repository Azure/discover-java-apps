
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: generate
generate: mockgen yaml2go
	$(MOCKGEN) -source=./springboot/contract.go -destination=./springboot/springboot_mock_test.go -package=springboot

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: fmt vet ginkgo generate ## Run unit tests.
	$(GINKGO) -r -v --race -trace -covermode atomic -coverprofile=output/coverage.out --junit-report=output/test_report.xml ./...

.PHONY: build
build: fmt vet yaml2go ## Build binary for release
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -a -o bin/discovery_darwin_arm64 cli/*.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -o bin/discovery_darwin_amd64 cli/*.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bin/discovery_l cli/*.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a -o bin/discovery.exe cli/*.go

.PHONY: yaml2go
yaml2go: yaml2go-cli  ## Generate yaml config struct
	$(YAML2GO) -i springboot/config.yml -o springboot/yaml_cfg.go -p springboot -s YamlConfig

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
MOCKGEN ?= $(LOCALBIN)/mockgen
GINKGO ?= $(LOCALBIN)/ginkgo
YAML2GO ?= $(LOCALBIN)/yaml2go-cli

## Tool Versions
MOCKGEN_VERSION ?= v1.6.0
GINKGO_VERSION ?= v2.9.2

.PHONY: ginkgo
ginkgo: $(GINKGO) ## Download ginkgo locally if necessary.
$(GINKGO): $(LOCALBIN)
	test -s $(LOCALBIN)/ginkgo || GOBIN=$(LOCALBIN) go install github.com/onsi/ginkgo/v2/ginkgo@$(GINKGO_VERSION)

.PHONY: mockgen
mockgen: $(MOCKGEN) ## Download mockgen locally if necessary.
$(MOCKGEN): $(LOCALBIN)
	test -s $(LOCALBIN)/mockgen || GOBIN=$(LOCALBIN) go install github.com/golang/mock/mockgen@$(MOCKGEN_VERSION)

.PHONY: yaml2go-cli
yaml2go-cli: $(YAML2GO) ## Download yaml2go-cli locally if necessary.
$(YAML2GO): $(LOCALBIN)
	test -s $(LOCALBIN)/yaml2go-cli || GOBIN=$(LOCALBIN) go install github.com/Icemap/yaml2go-cli@latest