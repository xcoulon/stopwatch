# Makefile for the `stopwatch` project

# tools
CUR_DIR=$(shell pwd)
INSTALL_PREFIX=$(CUR_DIR)/bin
VENDOR_DIR=vendor
SOURCE_DIR ?= .
BINARY_PATH=$(INSTALL_PREFIX)/stopwatch

# Call this function with $(call log-info,"Your message")
define log-info =
@echo "INFO: $(1)"
endef


.PHONY: help
# Based on https://gist.github.com/rcmachado/af3db315e31383502660
## Display this help text.
help:/
	$(info Available targets)
	$(info -----------------)
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		helpCommand = substr($$1, 0, index($$1, ":")-1); \
		if (helpMessage) { \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			gsub(/##/, "\n                                     ", helpMessage); \
		} else { \
			helpMessage = "(No documentation)"; \
		} \
		printf "%-35s - %s\n", helpCommand, helpMessage; \
		lastLine = "" \
	} \
	{ hasComment = match(lastLine, /^## (.*)/); \
          if(hasComment) { \
            lastLine=lastLine$$0; \
	  } \
          else { \
	    lastLine = $$0 \
          } \
        }' $(MAKEFILE_LIST)

.PHONY: deps
## Download build dependencies.
deps: 
	dep ensure -v

$(INSTALL_PREFIX):
# Build artifacts dir
	@mkdir -p $(INSTALL_PREFIX)

.PHONY: prebuild-checks
## Check that all tools where found
prebuild-checks: $(INSTALL_PREFIX)

.PHONY: test
## run all tests excluding fixtures and vendored packages
test: deps 
	STOPWATCH_LOG_LEVEL=info \
	STOPWATCH_ENABLE_DB_LOGS=false \
	STOPWATCH_CLEAN_TEST_DATA=true \
	STOPWATCH_POSTGRES_DATABASE=test \
	go test -p 1 -v ./...


.PHONY: build
## build the binary executable from CLI
build: $(INSTALL_PREFIX) deps
	$(eval BUILD_COMMIT:=$(shell git rev-parse --short HEAD))
	$(eval BUILD_TAG:=$(shell git tag --contains $(BUILD_COMMIT)))
	$(eval BUILD_TIME:=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ'))
	@echo "building $(BINARY_PATH) (commit:$(BUILD_COMMIT) / tag:$(BUILD_TAG) / time:$(BUILD_TIME))"
	@go build -ldflags \
	  " -X github.com/vatriathlon/stopwatch/configuration.BuildCommit=$(BUILD_COMMIT)\
	    -X github.com/vatriathlon/stopwatch/configuration.BuildTag=$(BUILD_TAG) \
	    -X github.com/vatriathlon/stopwatch/configuration.BuildTime=$(BUILD_TIME)" \
	-o $(BINARY_PATH) \
	main.go

.PHONY: start-database
start-database:
	docker-compose up -d db

.PHONY: dev
## run with fresh
dev: start-database


.PHONY: lint
## run golangci-lint against project
lint:
	@golangci-lint run -E gofmt,golint,megacheck,misspell ./...


PARSER_DIFF_STATUS :=
