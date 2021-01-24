SRCS := $(shell find server -type f -name '*.go')

all: run

run: $(SRCS)
	go run ./server/cmd/minws
