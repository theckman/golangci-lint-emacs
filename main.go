package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/golangci/golangci-lint/pkg/commands"
	"github.com/golangci/golangci-lint/pkg/exitcodes"
)

// cleans up the go build output to look like linter errors
func printCleanOutput(r io.Reader, w io.Writer) {
	scanner := bufio.NewScanner(r)

	var allowNext bool

	for scanner.Scan() {
		t := scanner.Text()
		if allowNext {
			allowNext = false
			fmt.Fprintf(w, " %s", strings.TrimSpace(t))
			continue
		}

		if strings.HasPrefix(t, "#") {
			continue
		}

		if strings.HasPrefix(t, "\t") {
			continue
		}

		if strings.HasPrefix(t, "can't load package:") {
			continue
		}

		if strings.HasSuffix(t, "too many errors") {
			continue
		}

		if strings.HasSuffix(t, " in assignment:") {
			allowNext = true
		}

		fmt.Fprint(w, strings.TrimPrefix(t, "./"))

		if !allowNext {
			fmt.Fprintln(w)
		}
	}
}

func printOutput(r io.Reader, w io.Writer) {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		fmt.Fprintln(w, scanner.Text())
	}
}

// builder is a generic go command runner to check that the source builds
//
// if the program failed to build, and looks like a syntax error of sorts, it
// prints that to os.Stdout and returns failed: true, err: nil
//
// if the go command failed for another reason, it echoes that out to Stderr
// and does an os.Exit(2)
//
// if it doesn't even seem like it called the Go command, it returns that back up the stack
func builder(goBin, mode, path string, flags ...string) (failed bool, err error) {
	buf := &bytes.Buffer{}

	f := make([]string, 0, len(flags)+2)
	f = append(f, mode)
	f = append(f, flags...)
	f = append(f, path)

	cmd := exec.Command(goBin, f...) // #nosec
	cmd.Stdout = buf
	cmd.Stderr = buf

	// parse the error code to guess whether it was syntax related
	// ExitCode 2 looks to be that
	if err = cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			switch ee.ProcessState.ExitCode() {
			case 2, 1:
				printCleanOutput(buf, os.Stdout)
				return true, nil

			case 0: // not possible, but just in case...
				return false, nil

			default:
				printOutput(buf, os.Stderr)
				os.Exit(2)
			}
		}

		return true, fmt.Errorf("failed to run go build: %v", err)
	}

	return false, nil
}

// go build
// if there is a failure it does not return control to the program
func build(path string) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get go binary path: %v", err)
		os.Exit(2)
	}

	/*
		// check that `go build` compiles
		failedB, err := builder(goBin, "build", path, "-o", os.DevNull)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to build source: %v\n", err)
			os.Exit(2)
		}
	*/

	// check that `go test` compiles
	failedT, err := builder(goBin, "test", path, "-c", "-o", os.DevNull)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to build test source: %v\n", err)
		os.Exit(2)
	}

	// if failedB || failedT {
	if failedT {
		os.Exit(1)
	}
}

func main() {
	build(os.Args[len(os.Args)-1])

	e := commands.NewExecutor("golangci-lint-emacs", "?", "")

	if err := e.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "failed executing command with error %v\n", err)
		os.Exit(exitcodes.Failure)
	}
}
