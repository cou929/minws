SRCS := $(shell find . -type f -name '*.go')

all: run

run: $(SRCS)
	go run ./cmd/minws

.PHONY: local-server
local-server:
	go run ./tools/localsvr/ ./tools/test-client/

.PHONY: dev
dev:
	make local-server &
	make run
