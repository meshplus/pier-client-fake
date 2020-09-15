SHELL := /bin/bash
CURRENT_PATH = $(shell pwd)

GO  = GO111MODULE=on go

ifeq (docker,$(firstword $(MAKECMDGOALS)))
  DOCKER_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  $(eval $(DOCKER_ARGS):;@:)
endif

help: Makefile
	@echo "Choose a command run:"
	@sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/ /'

## make eth: build ethereum client plugin
fake:
	mkdir -p build
	$(GO) build -o build/fake-client ./*.go

## make linter: Run golanci-lint
linter:
	golangci-lint run -E goimports --skip-dirs-use-default -D staticcheck

