package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

func homeDir() string {
	// try to get the homedir from the environment
	if h := os.Getenv("HOME"); h != "" {
		return h
	}

	u, err := user.Current()
	if err == nil && len(u.HomeDir) > 0 {
		return u.HomeDir
	}

	// reasonable default fallbacks for me
	switch runtime.GOOS {
	case "linux":
		return "/home/theckman"
	default:
		return "/Users/theckman"
	}
}

func gopath() string {
	g := os.Getenv("GOPATH")
	if len(g) > 0 {
		return g
	}

	return filepath.Join(homeDir(), "go")
}

// cleans up the go build output to look like linter errors
func printCleanOutput(r io.Reader, w io.Writer) {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		t := scanner.Text()

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

		fmt.Fprintln(w, strings.TrimPrefix(t, "./"))
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

	// check that `go build` compiles
	failedB, err := builder(goBin, "build", path, "-o", "/dev/null")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to build source: %v\n", err)
		os.Exit(2)
	}

	// check that `go test` compiles
	failedT, err := builder(goBin, "test", path, "-c")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to build test source: %v\n", err)
		os.Exit(2)
	}

	if failedB || failedT {
		os.Exit(1)
	}
}

func main() {
	build(os.Args[len(os.Args)-1])

	bin := filepath.Join(gopath(), "/bin/golangci-lint")

	// hand off to the real golangci-lint
	// TODO(theckman): consider importing golangci-lint directly and invoking their library code
	//                 their package main is tiny!!
	err := syscall.Exec(bin, os.Args, os.Environ()) // #nosec
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to execve golangci-lint: %v", err)
		os.Exit(2)
	}
}
