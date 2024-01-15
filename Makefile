BIN ?= $(CURDIR)/bin

.PHONY: build
build:
	go build -v -o $(BIN)/mprotc ./main.go


.PHONY: test
test:
	go test ./...
