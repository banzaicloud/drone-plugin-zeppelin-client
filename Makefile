EXECUTABLE ?= pipeline_client
IMAGE ?= banzaicloud/$(EXECUTABLE)
COMMIT ?= $(shell git rev-parse --short HEAD)

LD_FLAGS = -X "main.version=$(COMMIT)"
PACKAGES = $(shell go list ./... | grep -v /vendor/)

.PHONY: _no-target-specified
_no-target-specified:
	$(error Please specify the target to make - `make list` shows targets.)

.PHONY: list
list:
	@$(MAKE) -pRrn : -f $(MAKEFILE_LIST) 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | sort

all: clean deps fmt vet build

clean:
	go clean -i ./...

deps:
	go get ./...

fmt:
	go fmt $(PACKAGES)

vet:
	go vet $(PACKAGES)

docker:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '-s -w $(LD_FLAGS)' -o $(EXECUTABLE)
	docker build --rm -t $(IMAGE) .
	docker tag $(IMAGE):latest $(IMAGE):$(COMMIT)

push:
	docker push $(IMAGE):latest
	docker push $(IMAGE):$(COMMIT)

$(EXECUTABLE): $(wildcard *.go)
	go build -ldflags '-s -w $(LD_FLAGS)' -o $(EXECUTABLE)

build: $(EXECUTABLE)