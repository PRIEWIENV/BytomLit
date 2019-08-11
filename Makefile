GOBIN = $(shell pwd)
GO ?= latest

GO_BUILD_EX_ARGS ?=

all: node

.PHONY: node clean

goget:
	build/env.sh go get -v ./...

node: goget
	build/env.sh go build -o node ${GO_BUILD_EX_ARGS}
	@echo "Done building."
	@echo "Run \"$(GOBIN)/node\" to launch node."

clean:
	rm node -f