# golangci-lint-emacs
This is a wrapper around `golangci-lint` to resolve some issues I had while
trying to use `flycheck-golangci-lint` with `lsp`. Specifically, as best I could
tell, the way the different checkers were executed resulted in loss of important
information.

In particular, I'd either get linting errors in my editor or compilation
failures. I could not find a configuration that permitted both. By writing this
wrapper around `golangci-lint`, I can invoke `go build` and format the syntax
errors to look like linter failures. This tricks my editor in to displaying
those failures too, and so I've gotten back compilation errors and kept linting
failures.

## License
The source code in the repo is released in to the Public Domain, so you can use
it however you want. This issue was driving me fucking insane, and so I hope it
can be helpful for someone else.

If you do make any changes you'd find useful, I encourage you to raise a PR.

## Building / Installing
### Build
```shell
make build
```

### Install in ~/bin/
```shell
make install
```

### Easy Mode (build + install)
```shell
make
```

## Notes
This project was built for use on my personal systems, and may not be suitable
for use on yours. The following assumptions are present:

* `$GOPATH` is `$HOME/go`
* `$HOME/go/bin/golangci-lint` is where the real `golangci-lint` is installed
* `$HOME/bin` is on the `PATH` before `$HOME/go/bin`, so this wrapper should be installed in `$HOME/bin`
