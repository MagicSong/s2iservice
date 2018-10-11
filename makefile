# note: call scripts from /scripts
BINARY_NAME=devopshere

.PHONY: all test build clean
all: build
build:
	hack/build-go.sh
run:
	hack/run-go.sh
local-run:
	hack/run-local-go.sh
image:
	hack/build-image.sh
clean:
	rm -rf cmd/server
 