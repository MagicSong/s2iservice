# note: call scripts from /scripts
BINARY_NAME=devopshere

.PHONY: all test build clean
all: build
build:
	scripts/build-go.sh
run:
	scripts/run-go.sh
local-run:
	scripts/run-local-go.sh
image:
	scripts/build-image.sh
clean:
	rm -rf cmd/server
