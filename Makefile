
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

OS=$(shell uname | tr '[:upper:]' '[:lower:]')

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
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

##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

GOLANGCI_LINT = $(shell pwd)/bin/golangci-lint
GOLANGCI_LINT_VERSION ?= v1.54.2
golangci-lint:
	@[ -f $(GOLANGCI_LINT) ] || { \
	set -e ;\
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell dirname $(GOLANGCI_LINT)) $(GOLANGCI_LINT_VERSION) ;\
	}

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter & yamllint
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

##@ Build

.PHONY: check-go-target
check-go-target: ## Check presence of GOOS and GOARCH vars.
	@if [ -z "$(GOOS)" ]; then \
		echo "GOOS is not defined. Define GOOS y try again."; \
		exit 1; \
	fi
	@if [ -z "$(GOARCH)" ]; then \
		echo "GOARCH is not defined. Define GOARCH y try again."; \
		exit 1; \
	fi

check-cgo-switch: ## Check presence of CGO_ENABLED var.
	@if [ -z "$(CGO_ENABLED)" ]; then \
		echo "CGO_ENABLED is not defined. Define CGO_ENABLED y try again."; \
		exit 1; \
	fi

.PHONY: build
build: fmt vet check-go-target check-cgo-switch ## Build CLI binary.
	go build -o bin/nsmurder-$(GOOS)-$(GOARCH) cmd/main.go

.PHONY: run
run: fmt vet ## Run a CLI from your host.
	go run ./cmd/main.go

PACKAGE_NAME ?= package.tar.gz
.PHONY: package
package: check-go-target ## Package binary.
	@printf "\nCreating package at dist/$(PACKAGE_NAME) \n"
	@mkdir -p dist

	@if [ "$(OS)" = "linux" ]; then \
		tar --transform="s/nsmurder-$(GOOS)-$(GOARCH)/nsmurder/" -cvzf dist/$(PACKAGE_NAME) -C bin nsmurder-$(GOOS)-$(GOARCH) -C ../ LICENSE README.md; \
	elif [ "$(OS)" = "darwin" ]; then \
		tar -cvzf dist/$(PACKAGE_NAME) -s '/nsmurder-$(GOOS)-$(GOARCH)/nsmurder/' -C bin nsmurder-$(GOOS)-$(GOARCH) -C ../ LICENSE README.md; \
	else \
		echo "Unsupported OS: $(GOOS)"; \
		exit 1; \
	fi

.PHONY: package-signature
package-signature: ## Create a signature for the package.
	@printf "\nCreating package signature at dist/$(PACKAGE_NAME).md5 \n"
	md5sum dist/$(PACKAGE_NAME) | awk '{ print $$1 }' > dist/$(PACKAGE_NAME).md5
