SRCS := $(shell find . -type f -name '*.go')

all: run

run: $(SRCS)
	go run ./cmd/minws
