.PHONY: all binay test test-unit test-integration test-coverage validate-gofmt validate-golint validate build


GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
DOCKER_IMAGE := docker-dev$(if $(GIT_BRANCH),:$(GIT_BRANCH))
# DOCKER_MOUNT := -v "$(CURDIR)/bin:/usr/src/praetorian/bin" -v "$(CURDIR)/pkg:/usr/src/praetorian/pkg"

# DOCKER_RUN_PRAETORIAN := docker run --rm -it $(DOCKER_MOUNT) "$(DOCKER_IMAGE)"
DOCKER_RUN_PRAETORIAN := docker run --rm -it "$(DOCKER_IMAGE)"

all: validate build test

build:
	docker build -t "$(DOCKER_IMAGE)" .

binary: build
	$(DOCKER_RUN_PRAETORIAN) script/binary


test: test-unit test-integration

test-unit: build
	$(DOCKER_RUN_PRAETORIAN) script/test-unit

test-integration: build
	$(DOCKER_RUN_PRAETORIAN) script/test-integration

test-coverage: test-unit
	$(DOCKER_RUN_PRAETORIAN) script/test-coverage

validate-gofmt: build
	$(DOCKER_RUN_PRAETORIAN) script/validate-gofmt

validate-golint: build
	$(DOCKER_RUN_PRAETORIAN) script/validate-golint

validate: validate-gofmt validate-golint

shell: build
       $(DOCKER_RUN_PRAETORIAN) bash
