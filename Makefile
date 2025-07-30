VERSION          := snapshot
NAME             := demo

GIT_HEAD         := $(shell git rev-parse HEAD)
PACKAGES         := $(shell find . -name *.go | grep -v -E "vendor|tools" | xargs -n1 dirname | sort -u)
MAIN_DIR         := .
TEST_FLAGS       := -race -count=1 -mod=readonly -cover -coverprofile coverprofile.txt
LINK_FLAGS       := -X main.Version=$(VERSION) -X main.GitHead=$(GIT_HEAD)
BUILD_FLAGS      := -mod=readonly -v

.PHONY: download
download:
	@echo Download go.mod dependencies
	@go mod download

BIN   := $(shell pwd)/.bin
TOOLS := $(shell pwd)/tools

BIN_PATH := PATH="$(abspath $(BIN)):$$PATH"

.PHONY: install
install: download
	@echo Installing tools from $(TOOLS)/tools.go
	@cd $(TOOLS) && go mod download
	@cd $(TOOLS) && cat tools.go | grep "_" | awk -F'"' '{print $$2}' | GOBIN=$(BIN) xargs -tI % go install %

.PHONY: default
default: build

.PHONY: generate
generate:
	$(BIN_PATH) go generate $(PACKAGES)

.PHONY: test-generate
test-generate: install generate test

.PHONY: lint
lint: run-lint

.PHONY: test
test: run-lint run-test

.PHONY: build
build:
	@mkdir -p build
	CGO_ENABLED=0 go build $(BUILD_FLAGS) -ldflags="$(LINK_FLAGS)" -o build/$(NAME) $(MAIN_DIR)
	@echo build complete

.PHONY: clean
clean:
	rm -rvf build coverprofile.txt

.PHONY: run-lint
run-lint:
	$(BIN_PATH) golangci-lint --version
	$(BIN_PATH) golangci-lint run $(PACKAGES)

.PHONY: run-test
run-test:
	go test $(TEST_FLAGS) $(PACKAGES)
