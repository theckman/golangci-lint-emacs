build:
	go build -o golangci-lint .

install:
	mv golangci-lint ${HOME}/bin/golangci-lint

clean:
	rm golangci-lint

clean_install: clean
	rm ${HOME}/bin/golangci-lint

default: build install

.DEFAULT_GOAL := default

.PHONY: build install default
