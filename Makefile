# set defaults
ifndef DOCKER_HUB_REPO
    DOCKER_HUB_REPO := portworx
    $(warning DOCKER_HUB_REPO not defined, using '$(DOCKER_HUB_REPO)' instead)
endif
ifndef DOCKER_HUB_PX_OBJECT_CONTROLLER_IMG
    DOCKER_HUB_PX_OBJECT_CONTROLLER_IMG := px-object-controller
    $(warning DOCKER_HUB_PX_OBJECT_CONTROLLER_IMG not defined, using '$(DOCKER_HUB_PX_OBJECT_CONTROLLER_IMG)' instead)
endif
ifndef DOCKER_HUB_PX_OBJECT_CONTROLLER_TAG
    DOCKER_HUB_PX_OBJECT_CONTROLLER_TAG := latest
    $(warning DOCKER_HUB_PX_OBJECT_CONTROLLER_TAG not defined, using '$(DOCKER_HUB_PX_OBJECT_CONTROLLER_TAG)' instead)
endif
ifndef DOCKER_HUB_PX_OBJECT_CONTROLLER_TEST_IMG
    DOCKER_HUB_PX_OBJECT_CONTROLLER_TEST_IMG := px-object-controller-test
    $(warning DOCKER_HUB_PX_OBJECT_CONTROLLER_TEST_IMG not defined, using '$(DOCKER_HUB_PX_OBJECT_CONTROLLER_TEST_IMG)' instead)
endif
ifndef DOCKER_HUB_PX_OBJECT_CONTROLLER_TEST_TAG
    DOCKER_HUB_PX_OBJECT_CONTROLLER_TEST_TAG := latest
    $(warning DOCKER_HUB_PX_OBJECT_CONTROLLER_TEST_TAG not defined, using '$(DOCKER_HUB_PX_OBJECT_CONTROLLER_TEST_TAG)' instead)
endif

HAS_GOMODULES := $(shell go help mod why 2> /dev/null)

ifdef HAS_GOMODULES
export GO111MODULE=on
export GOFLAGS=-mod=vendor
else
$(error px-object controller can only be built with go 1.11+ which supports go modules)
endif

ifndef PKGS
PKGS := $(shell GOFLAGS=-mod=vendor go list ./... 2>&1 )
endif

GO_FILES := $(shell find . -name '*.go' | grep -v vendor | \
                                   grep -v '\.pb\.go' | \
                                   grep -v '\.pb\.gw\.go' | \
                                   grep -v 'externalversions' | \
                                   grep -v 'versioned' | \
                                   grep -v 'client')

ifeq ($(BUILD_TYPE),debug)
BUILDFLAGS += -gcflags "-N -l"
endif

RELEASE_VER := 1.0.0
BASE_DIR    := $(shell git rev-parse --show-toplevel)
GIT_SHA     := $(shell git rev-parse --short HEAD)
BIN         := $(BASE_DIR)/bin

VERSION = $(RELEASE_VER)-$(GIT_SHA)

LDFLAGS += "-s -w -X github.com/portworx/px-object-controller/pkg/version.Version=$(VERSION)"
BUILD_OPTIONS := -ldflags=$(LDFLAGS)

PX_OBJECT_CONTROLLER_IMG=$(DOCKER_HUB_REPO)/$(DOCKER_HUB_PX_OBJECT_CONTROLLER_IMG):$(DOCKER_HUB_PX_OBJECT_CONTROLLER_TAG)
PX_OBJECT_CONTROLLER_TEST_IMG=$(DOCKER_HUB_REPO)/$(DOCKER_HUB_PX_OBJECT_CONTROLLER_TEST_IMG):$(DOCKER_HUB_PX_OBJECT_CONTROLLER_TEST_TAG)
.DEFAULT_GOAL=all
.PHONY: px-object-controller deploy clean vendor vendor-update test

all: px-object-controller pretest

vendor-update:
	export GOSUMDB=off
	go mod download

vendor:
	go mod vendor

# Tools download  (if missing)
# - please make sure $GOPATH/bin is in your path, also do not use $GOBIN

$(GOPATH)/bin/golint:
	GO111MODULE=off go get -u golang.org/x/lint/golint

$(GOPATH)/bin/errcheck:
	GO111MODULE=off go get -u github.com/kisielk/errcheck

$(GOPATH)/bin/staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@v0.2.1

$(GOPATH)/bin/revive:
	GO111MODULE=off go get -u github.com/mgechev/revive

$(GOPATH)/bin/gomock:
	go get -u github.com/golang/mock/gomock

$(GOPATH)/bin/mockgen:
	go get -u github.com/golang/mock/mockgen

$(GOPATH)/bin/contextcheck:
	GO111MODULE=off go get -u github.com/sylvia7788/contextcheck

setup-travis:
	curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.11.1/kind-linux-amd64
	chmod +x ./kind
	sudo mv ./kind /usr/local/bin
	curl -LO "https://dl.k8s.io/release/v1.23.5/bin/linux/amd64/kubectl"
	chmod +x ./kubectl
	sudo mv ./kubectl /usr/local/bin

# Static checks

vendor-tidy:
	go mod tidy

lint: $(GOPATH)/bin/golint
	# golint check ...
	@for file in $(GO_FILES); do \
		golint $${file}; \
		if [ -n "$$(golint $${file})" ]; then \
			exit 1; \
		fi; \
	done

vet:
	# go vet check ...
	@go vet $(PKGS)

contextcheck: $(GOPATH)/bin/contextcheck
	@contextcheck $(PKGS)

check-fmt:
	# gofmt check ...
	@bash -c "diff -u <(echo -n) <(gofmt -l -d -s -e $(GO_FILES))"

errcheck: $(GOPATH)/bin/errcheck
	# errcheck check ...
	@errcheck -verbose -blank $(PKGS)

staticcheck: $(GOPATH)/bin/staticcheck
	# staticcheck check ...
	@staticcheck $(PKGS)

revive: $(GOPATH)/bin/revive
	# revive check ...
	@revive -formatter friendly $(PKGS)

pretest: check-fmt lint vet #staticcheck staticcheck is broken upstream, wait till fix

test:
	echo "" > coverage.txt
	for pkg in $(PKGS);	do \
		go test -v -coverprofile=profile.out -covermode=atomic -coverpkg=$${pkg}/... $${pkg} || exit 1; \
		if [ -f profile.out ]; then \
			cat profile.out >> coverage.txt; \
			rm profile.out; \
		fi; \
	done
	sed -i '/mode: atomic/d' coverage.txt
	sed -i '1d' coverage.txt
	sed -i '1s/^/mode: atomic\n/' coverage.txt

kind-setup:
	@echo "Setting up kind cluster"
	@kind delete cluster --name px-object-controller
	@./hack/setup-kind.sh

kind-teardown:
	@kind delete cluster --name px-object-controller

integration-test:
	@set -x
	@echo "Running px-object-controller integration tests"
	@kind load docker-image --name px-object-controller $(PX_OBJECT_CONTROLLER_IMG)
	@cd test/integration && PX_OBJECT_CONTROLLER_IMG=${PX_OBJECT_CONTROLLER_IMG} go test -tags integrationtest -v -kubeconfig=/tmp/px-object-controller-kubeconfig.yaml

test-setup:
	@kubectl apply -f client/config/crd
	@kubectl -n kube-system create secret docker-registry pwxbuild --docker-username=${DOCKER_USER} --docker-password=${DOCKER_PASSWORD}

integration-test-suite: kind-setup test-setup integration-test kind-teardown

codegen:
	@echo "Generating code"
	./client/hack/update-crd.sh
	./client/hack/update-codegen.sh

px-object-controller:
	@echo "Building the cluster px-object-controller binary"
	@cd cmd/px-object-controller && CGO_ENABLED=0 go build $(BUILD_OPTIONS) -o $(BIN)/px-object-controller

sample-app:
	@echo "Building the sample app binary"
	@cd examples/sample-app && CGO_ENABLED=0 go build $(BUILD_OPTIONS) -o $(BIN)/sample-app
	@docker build --tag ggriffiths/pos-sample-app -f examples/sample-app/Dockerfile .


container:
	@echo "Building px-object-controller image $(PX_OBJECT_CONTROLLER_IMG)"
	docker build --tag $(PX_OBJECT_CONTROLLER_IMG) -f cmd/px-object-controller/Dockerfile .

deploy:
	@echo "Pushing px-object-controller image $(PX_OBJECT_CONTROLLER_IMG)"
	docker push $(PX_OBJECT_CONTROLLER_IMG)

cleanconfigs:
	rm -rf bin/configs

mockgen: $(GOPATH)/bin/gomock $(GOPATH)/bin/mockgen

clean:
	@echo "Cleaning up binaries"
	@rm -rf $(BIN)
	@go clean -i $(PKGS)
	@echo "Deleting image "$(PX_OBJECT_CONTROLLER_IMG)
	@docker rmi -f $(PX_OBJECT_CONTROLLER_IMG)
