build:
	go build -ldflags="-s -w" -o golangci-lint .

install:
	mv golangci-lint ${HOME}/bin/golangci-lint

clean:
	rm -f golangci-lint

clean_install: clean
	rm ${HOME}/bin/golangci-lint

default: build install

.DEFAULT_GOAL := default

.PHONY: build install default
