GO111MODULE?=on
export GO111MODULE

BINARY_PATH := bin/stopwatch

.PHONY: build
## build the binary executable from CLI
build: $(INSTALL_PREFIX)
	$(eval BUILD_COMMIT:=$(shell git rev-parse --short HEAD))
	$(eval BUILD_TAG:=$(shell git tag --contains $(BUILD_COMMIT)))
	$(eval BUILD_TIME:=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ'))
	@echo "building $(BINARY_PATH) (commit:$(BUILD_COMMIT) / tag:$(BUILD_TAG) / time:$(BUILD_TIME))"
	@go build -ldflags \
	  " -X github.com/vatriathlon/stopwatch/pkg/configuration.BuildCommit=$(BUILD_COMMIT)\
	    -X github.com/vatriathlon/stopwatch/pkg/configuration.BuildTag=$(BUILD_TAG) \
	    -X github.com/vatriathlon/stopwatch/pkg/configuration.BuildTime=$(BUILD_TIME)" \
	-o $(BINARY_PATH) \
	main.go