.PHONY: default build test style docker binaries clean

DOCKER       ?= docker
GO           ?= go
GOFMT        ?= $(GO)fmt
APP          := carousel
DOCKER_ORG   := xmidt
FIRST_GOPATH := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
BINARY    	 := $(FIRST_GOPATH)/bin/$(APP)

VERSION ?= $(shell git describe --tag --always --dirty)
PROGVER ?= $(shell git describe --tags `git rev-list --tags --max-count=1` | tail -1 | sed 's/v\(.*\)/\1/')
BUILDTIME = $(shell date -u '+%c')
GITCOMMIT = $(shell git rev-parse --short HEAD)
GOBUILDFLAGS = -ldflags "-X 'main.BuildTime=$(BUILDTIME)' -X main.GitCommit=$(GITCOMMIT) -X main.Version=$(VERSION)" -o $(APP)

default: build

generate:
	$(GO) get github.com/abice/go-enum
	$(GO) generate ./...
	$(GO) mod vendor

test:
	$(GO) test -v -race  -coverprofile=coverage.txt $$(go list ./... | grep -v integration)
	$(GO) test -v -race  -json $$(go list ./... | grep -v integration) > report.json

acceptance:
	# check if terraform is installed
	@which terraform > /dev/null
	-mkdir integration/plugin
	-$(GO) build -buildmode=plugin -o integration/plugin/evenHostValidator.so example/main.go
	$(GO) test -v ./integration/

style:
	! $(GOFMT) -d $$(find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

check:
	golangci-lint run -n | tee errors.txt

build:
	$(GO) build $(GOBUILDFLAGS) ./cmd

release: build
	upx $(APP)

docker:
	-$(DOCKER) rmi "$(DOCKER_ORG)/$(APP):$(VERSION)"
	-$(DOCKER) rmi "$(DOCKER_ORG)/$(APP):latest"
	$(DOCKER) build -t "$(DOCKER_ORG)/$(APP):$(VERSION)" -t "$(DOCKER_ORG)/$(APP):latest" .

binaries: generate
	mkdir -p ./.ignore
	GOOS=darwin GOARCH=amd64 $(GO) build -o ./.ignore/$(APP)-$(PROGVER).darwin-amd64 -ldflags "-X 'main.BuildTime=$(BUILDTIME)' -X main.GitCommit=$(GITCOMMIT) -X main.Version=$(VERSION)" ./cmd
	GOOS=linux  GOARCH=amd64 $(GO) build -o ./.ignore/$(APP)-$(PROGVER).linux-amd64 -ldflags "-X 'main.BuildTime=$(BUILDTIME)' -X main.GitCommit=$(GITCOMMIT) -X main.Version=$(VERSION)" ./cmd


clean:
	-rm -r .ignore/ $(APP) errors.txt report.json coverage.txt
