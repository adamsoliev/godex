BINDIR := bin
BINARY := godex
CMD := ./cmd/godex

.PHONY: build run test fmt clean

build:
	mkdir -p $(BINDIR)
	go build -o $(BINDIR)/$(BINARY) $(CMD)

run: build
	$(BINDIR)/$(BINARY)

test:
	go test ./...

fmt:
	gofmt -w ./cmd ./internal

clean:
	rm -rf $(BINDIR)
