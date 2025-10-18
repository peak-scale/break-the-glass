# Version
GIT_HEAD_COMMIT ?= $(shell git rev-parse --short HEAD)
VERSION         ?= $(or $(shell git describe --abbrev=0 --tags --match "v*" 2>/dev/null),$(GIT_HEAD_COMMIT))
GOOS            ?= $(shell go env GOOS)
GOARCH          ?= $(shell go env GOARCH)

# Defaults
REGISTRY        ?= ghcr.io
REPOSITORY      ?= peak-scale/break-the-glass
GIT_TAG_COMMIT  ?= $(shell git rev-parse --short $(VERSION))
GIT_MODIFIED_1  ?= $(shell git diff $(GIT_HEAD_COMMIT) $(GIT_TAG_COMMIT) --quiet && echo "" || echo ".dev")
GIT_MODIFIED_2  ?= $(shell git diff --quiet && echo "" || echo ".dirty")
GIT_MODIFIED    ?= $(shell echo "$(GIT_MODIFIED_1)$(GIT_MODIFIED_2)")
GIT_REPO        ?= $(shell git config --get remote.origin.url)
BUILD_DATE      ?= $(shell git log -1 --format="%at" | xargs -I{} sh -c 'if [ "$(shell uname)" = "Darwin" ]; then date -r {} +%Y-%m-%dT%H:%M:%S; else date -d @{} +%Y-%m-%dT%H:%M:%S; fi')
IMG_BASE        ?= $(REPOSITORY)
IMG             ?= $(IMG_BASE):$(VERSION)
FULL_IMG          ?= $(REGISTRY)/$(IMG_BASE)

## Kubernetes Version Support
KUBERNETES_SUPPORTED_VERSION ?= v1.33.0

## Tool Binaries
KUBECTL ?= kubectl
HELM ?= helm

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

plugin-build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o kubectl-accessdev plugin/main.go

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

# Generate License Header
license-headers: nwa
	$(NWA) config

.PHONY: golint
golint: golangci-lint
	$(GOLANGCI_LINT) run -c .golangci.yml

.PHONY: golint-fix
golint-fix: golangci-lint
	$(GOLANGCI_LINT) run -c .golangci.yml --fix

manifests: controller-gen 
	$(CONTROLLER_GEN) crd:generateEmbeddedObjectMeta=true webhook paths="./..." output:crd:artifacts:config=charts/break-the-glass/crds
	make apidocs

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) crd:generateEmbeddedObjectMeta=true object:headerFile="hack/boilerplate.go.txt" paths="./..."


apidocs: TARGET_DIR      := $(shell mktemp -d)
apidocs: apidocs-gen generate
	$(APIDOCS_GEN) crdoc --resources charts/break-the-glass/crds --output docs/reference.md --template ./hack/templates/crds.tmpl

.PHONY: test
test: generate manifests mocks setup-envtest
	@GO111MODULE=on go test -v $(shell go list ./... | grep -v "e2e") -coverprofile coverage.out

.PHONY: test-clean
test-clean: ## Clean tests cache
	@go clean -testcache

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter & yamllint
	$(GOLANGCI_LINT) run -c .golangci.yml

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run -c .golangci.yml --fix

# generate mocks
.PHONY: mocks
mocks: mockgen
	$(MOCKGEN) -destination internal/mocks/client/mock.go sigs.k8s.io/controller-runtime/pkg/client Client,SubResourceWriter

##@ Build

.PHONY: build
build: manifests generate fmt vet ## Build manager binary.
	go build -o bin/manager cmd/main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./cmd/main.go

####################
# -- Docker
####################

KO_PLATFORM     ?= linux/$(GOARCH)
KOCACHE         ?= /tmp/ko-cache
KO_REGISTRY     := ko.local
KO_TAGS         ?= "latest"
ifdef VERSION
KO_TAGS         := $(KO_TAGS),$(VERSION)
endif

LD_FLAGS        := "-X github.com/peak-scale/break-the-glass/internal/version.Version=$(VERSION) \
					-X github.com/peak-scale/break-the-glass/internal/version.GitCommit=$(GIT_HEAD_COMMIT) \
					-X ithub.com/peak-scale/break-the-glass/internal/version.BuildDate=$(BUILD_DATE)"
# Docker Image Build
# ------------------
.PHONY: ko-build-controller
ko-build-controller: ko 
	@echo Building Controller $(FULL_IMG) - $(KO_TAGS) >&2
	@LD_FLAGS=$(LD_FLAGS) KOCACHE=$(KOCACHE) KO_DOCKER_REPO=$(FULL_IMG) \
		$(KO) build ./cmd/ --bare --tags=$(KO_TAGS) --push=false --local --platform=$(KO_PLATFORM)

.PHONY: ko-build-all
ko-build-all:  ko-build-controller

# Docker Image Publish
# ------------------

REGISTRY_PASSWORD   ?= dummy
REGISTRY_USERNAME   ?= dummy

.PHONY: ko-login
ko-login: ko
	@$(KO) login $(REGISTRY) --username $(REGISTRY_USERNAME) --password $(REGISTRY_PASSWORD)

.PHONY: ko-publish-controller
ko-publish-controller: ko-login
	@echo Publishing Controller $(FULL_IMG) - $(KO_TAGS) >&2
	@LD_FLAGS=$(LD_FLAGS) KOCACHE=$(KOCACHE) KO_DOCKER_REPO=$(FULL_IMG) \
		$(KO) build ./cmd/ --bare --tags=$(KO_TAGS) --push=true

.PHONY: ko-publish-all
ko-publish-all: ko-publish-controller

# CLI Publish

cli-test-release: goreleaser
	$(GORELEASER) --skip=publish --snapshot --clean --parallelism 2

####################
# -- Helm
####################

# Helm
SRC_ROOT = $(shell git rev-parse --show-toplevel)

helm-docs: helm-doc
	$(HELM_DOCS) --chart-search-root ./charts

helm-lint: ct
	@$(CT) lint --config .github/configs/ct.yaml --lint-conf .github/configs/lintconf.yaml --all --debug

helm-schema: helm-plugin-schema
	cd charts/sops-operator && $(HELM) schema --use-helm-docs

helm-test: kind ct
	@$(KIND) create cluster --wait=60s --name helm-break-the-glass --image=kindest/node:$(KUBERNETES_SUPPORTED_VERSION)
	@$(MAKE) helm-test-exec
	@$(KIND) delete cluster --name helm-break-the-glass

helm-test-exec: ct ko-build-all
	@$(KIND) load docker-image --name helm-break-the-glass $(FULL_IMG):latest
	@$(CT) install --config $(SRC_ROOT)/.github/configs/ct.yaml --all --debug


####################
# -- Install E2E Tools
####################
CLUSTER_NAME ?= "break-the-glass"

e2e: e2e-build e2e-exec e2e-destroy

e2e-build: kind
	$(KIND) create cluster --wait=60s --config e2e/kind.yaml --name $(CLUSTER_NAME) --image=kindest/node:$(KUBERNETES_SUPPORTED_VERSION)
	$(MAKE) e2e-install

e2e-exec: ginkgo
	$(GINKGO) -r -vv ./e2e

e2e-destroy: kind
	$(KIND) delete cluster --name $(CLUSTER_NAME)

e2e-install: e2e-install-distro e2e-install-addon-helm

e2e-install-addon-helm: VERSION :=v0.0.0
e2e-install-addon-helm: KO_TAGS :=v0.0.0
e2e-install-addon-helm: e2e-load-image ko-build-all
	helm upgrade \
	    --dependency-update \
		--debug \
		--install \
		--namespace break-the-glass \
		--create-namespace \
		--set 'image.pullPolicy=Never' \
		--set "image.tag=$(VERSION)" \
		--set args.logLevel=10 \
		--set args.pprof=true \
		break-the-glass \
		./charts/break-the-glass

e2e-install-distro:
	@$(KUBECTL) kustomize e2e/manifests/flux/ | kubectl apply -f -
	@$(KUBECTL) kustomize e2e/manifests/distro/ | kubectl apply -f -
	@$(MAKE) wait-for-helmreleases

.PHONY: e2e-load-image
e2e-load-image: kind ko-build-all
	$(KIND) load docker-image --name $(CLUSTER_NAME) $(FULL_IMG):$(VERSION)

wait-for-helmreleases:
	@ echo "Waiting for all HelmReleases to have observedGeneration >= 0..."
	@while [ "$$(kubectl get helmrelease -A -o jsonpath='{range .items[?(@.status.observedGeneration<0)]}{.metadata.namespace}{" "}{.metadata.name}{"\n"}{end}' | wc -l)" -ne 0 ]; do \
	  sleep 5; \
	done

# Setup development env
# Usage:
# 	LAPTOP_HOST_IP=<YOUR_LAPTOP_IP> make dev-setup
# For example:
#	LAPTOP_HOST_IP=192.168.10.101 make dev-setup
define TLS_CNF
[ req ]
default_bits       = 4096
distinguished_name = req_distinguished_name
req_extensions     = req_ext
[ req_distinguished_name ]
countryName                = SG
stateOrProvinceName        = SG
localityName               = SG
organizationName           = CAPSULE
commonName                 = CAPSULE
[ req_ext ]
subjectAltName = @alt_names
[alt_names]
IP.1   = $(LAPTOP_HOST_IP)
endef
export TLS_CNF
dev-setup:
	mkdir -p /tmp/k8s-webhook-server/serving-certs
	echo "$${TLS_CNF}" > _tls.cnf
	openssl req -newkey rsa:4096 -days 3650 -nodes -x509 \
		-subj "/C=SG/ST=SG/L=SG/O=CAPSULE/CN=CAPSULE" \
		-extensions req_ext \
		-config _tls.cnf \
		-keyout /tmp/k8s-webhook-server/serving-certs/tls.key \
		-out /tmp/k8s-webhook-server/serving-certs/tls.crt
	$(KUBECTL) create secret tls capsule-tls -n capsule-system \
		--cert=/tmp/k8s-webhook-server/serving-certs/tls.crt\
		--key=/tmp/k8s-webhook-server/serving-certs/tls.key || true
	rm -f _tls.cnf
	export WEBHOOK_URL="https://$${LAPTOP_HOST_IP}:9443"; \
	export CA_BUNDLE=`openssl base64 -in /tmp/k8s-webhook-server/serving-certs/tls.crt | tr -d '\n'`; \
	$(HELM) upgrade \
	    --dependency-update \
		--debug \
		--install \
		--namespace break-the-glass \
		--create-namespace \
		--set 'image.pullPolicy=Never' \
		--set "image.tag=$(VERSION)" \
		--set args.logLevel=10 \
		--set args.pprof=true \
		break-the-glass \
		./charts/break-the-glass
	$(KUBECTL) -n break-the-glass scale deployment --all --replicas=0 || true


## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

####################
# -- Helm Plugins
####################
HELM_SCHEMA_VERSION   := ""
helm-plugin-schema:
	$(HELM) plugin install https://github.com/losisin/helm-values-schema-json.git --version $(HELM_SCHEMA_VERSION) || true

HELM_DOCS         := $(LOCALBIN)/helm-docs
HELM_DOCS_VERSION := v1.14.2
HELM_DOCS_LOOKUP  := norwoodj/helm-docs
helm-doc:
	@test -s $(HELM_DOCS) || \
	$(call go-install-tool,$(HELM_DOCS),github.com/$(HELM_DOCS_LOOKUP)/cmd/helm-docs@$(HELM_DOCS_VERSION))

####################
# -- Tools
####################
CONTROLLER_GEN         := $(LOCALBIN)/controller-gen
CONTROLLER_GEN_VERSION := v0.19.0
CONTROLLER_GEN_LOOKUP  := kubernetes-sigs/controller-tools
controller-gen:
	@test -s $(CONTROLLER_GEN) && $(CONTROLLER_GEN) --version | grep -q $(CONTROLLER_GEN_VERSION) || \
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VERSION))

GINKGO := $(LOCALBIN)/ginkgo
ginkgo:
	$(call go-install-tool,$(GINKGO),github.com/onsi/ginkgo/v2/ginkgo)

NWA           := $(LOCALBIN)/nwa
NWA_VERSION   := v0.7.5
NWA_LOOKUP    := B1NARY-GR0UP/nwa
nwa:
	@test -s $(NWA) && $(NWA) -h | grep -q $(NWA_VERSION) || \
	$(call go-install-tool,$(NWA),github.com/$(NWA_LOOKUP)@$(NWA_VERSION))

CT         := $(LOCALBIN)/ct
CT_VERSION := v3.13.0
CT_LOOKUP  := helm/chart-testing
ct:
	@test -s $(CT) && $(CT) version | grep -q $(CT_VERSION) || \
	$(call go-install-tool,$(CT),github.com/$(CT_LOOKUP)/v3/ct@$(CT_VERSION))

KIND         := $(LOCALBIN)/kind
KIND_VERSION := v0.30.0
KIND_LOOKUP  := kubernetes-sigs/kind
kind:
	@test -s $(KIND) && $(KIND) --version | grep -q $(KIND_VERSION) || \
	$(call go-install-tool,$(KIND),sigs.k8s.io/kind/cmd/kind@$(KIND_VERSION))

KO           := $(LOCALBIN)/ko
KO_VERSION   := v0.18.0
KO_LOOKUP    := google/ko
ko:
	@test -s $(KO) && $(KO) -h | grep -q $(KO_VERSION) || \
	$(call go-install-tool,$(KO),github.com/$(KO_LOOKUP)@$(KO_VERSION))

ENVTEST             ?= $(LOCALBIN)/setup-envtest
ENVTEST_VERSION     ?= $(shell go list -m -f "{{ .Version }}" sigs.k8s.io/controller-runtime | awk -F'[v.]' '{printf "release-%d.%d", $$2, $$3}')
ENVTEST_K8S_VERSION ?= $(shell go list -m -f "{{ .Version }}" k8s.io/api | awk -F'[v.]' '{printf "1.%d", $$3}')
envtest:
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest,$(ENVTEST_VERSION))

.PHONY: setup-envtest
setup-envtest: envtest ## Download the binaries required for ENVTEST in the local bin directory.
	@echo "Setting up envtest binaries for Kubernetes version $(ENVTEST_K8S_VERSION)..."
	@$(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path || { \
		echo "Error: Failed to set up envtest binaries for version $(ENVTEST_K8S_VERSION)."; \
		exit 1; \
	}

GOLANGCI_LINT          := $(LOCALBIN)/golangci-lint
GOLANGCI_LINT_VERSION  := 2.5.0
GOLANGCI_LINT_LOOKUP   := golangci/golangci-lint
golangci-lint: ## Download golangci-lint locally if necessary.
		test -s $(GOLANGCI_LINT) && $(GOLANGCI_LINT) --version | grep -q $(GOLANGCI_LINT_VERSION) ||  \
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/$(GOLANGCI_LINT_LOOKUP)/v2/cmd/golangci-lint@v$(GOLANGCI_LINT_VERSION))

APIDOCS_GEN         := $(LOCALBIN)/crdoc
APIDOCS_GEN_VERSION := v0.6.4
APIDOCS_GEN_LOOKUP  := fybrik/crdoc
apidocs-gen: ## Download crdoc locally if necessary.
	@test -s $(APIDOCS_GEN) && $(APIDOCS_GEN) --version | grep -q $(APIDOCS_GEN_VERSION) || \
	$(call go-install-tool,$(APIDOCS_GEN),fybrik.io/crdoc@$(APIDOCS_GEN_VERSION))

MOCKGEN         := $(LOCALBIN)/mockgen
MOCKGEN_VERSION := v0.6.0
MOCKGEN_LOOKUP  := go.uber.org/mock/mockgen
mockgen:
	@test -s $(MOCKGEN) && $(MOCKGEN) -version | grep -q $(MOCKGEN_VERSION) || \
	$(call go-install-tool,$(MOCKGEN),$(MOCKGEN_LOOKUP)@$(MOCKGEN_VERSION))

GORELEASER          := $(LOCALBIN)/goreleaser
GORELEASER_VERSION  := 2.12.5
GORELEASER_LOOKUP   := goreleaser/goreleaser
goreleaser: ## Download goreleaser locally if necessary.
		test -s $(GORELEASER) && $(GORELEASER) --version | grep -q $(GORELEASER_VERSION) ||  \
	$(call go-install-tool,$(GORELEASER),github.com/$(GORELEASER_LOOKUP)/v2@v$(GORELEASER_VERSION))

# go-install-tool will 'go install' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-install-tool
[ -f $(1) ] || { \
    set -e ;\
    GOBIN=$(LOCALBIN) go install $(2) ;\
}
endef
