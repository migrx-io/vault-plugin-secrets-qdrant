REPO_DIR := $(shell basename $(CURDIR))
PLUGIN_DIR := $(GOPATH)/vault-plugins
PLUGIN_NAME := $(shell command ls cmd/)

.PHONY: default
default: build

.PHONY: build
build:
	@CGO_ENABLED=0 go build -o bin/$(PLUGIN_NAME) cmd/$(PLUGIN_NAME)/main.go

.PHONY: fmt
fmt:
	@gofmt -l -w .

.PHONY: setup-env
setup-env: build
	@cd bootstrap && docker compose -f ./docker-compose.yml down -t 1 && docker compose -f ./docker-compose.yml up --build -d

.PHONY: teardown-env
teardown-env:
	@cd bootstrap && docker-compose -f ./docker-compose.yml down -t 1
	@docker rmi -f bootstrap-vault

.PHONY: e2e
e2e:
	@docker build --network=host --progress=plain --no-cache -f tests/Dockerfile -t vault-jwt-e2e-test .

.PHONY: clean
clean: teardown-env
	@rm -rf bin/*
	@cd bootstrap && rm -rf qdrant_data
