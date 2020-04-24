package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
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

	// reasonable default fallbacks for me
	switch runtime.GOOS {
	case "linux":
		return "/home/theckman"
	default:
		return "/Users/theckman"
	}
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

// go build
// if there is a failure it does not return control to the program
func build(path string) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get go binary path: %v", err)
		os.Exit(2)
	}

	buf := &bytes.Buffer{}
	cmd := exec.Command(goBin, "build", "-o", "/dev/null", path) // #nosec
	cmd.Stdout = buf
	cmd.Stderr = buf

	// parse the error code to guess whether it was syntax related
	// ExitCode 2 looks to be that
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			switch ee.ProcessState.ExitCode() {
			case 2, 1:
				printCleanOutput(buf, os.Stdout)
				os.Exit(1)

			case 0: // not possible, but just in case...
				return

			default:
				printOutput(buf, os.Stderr)
				os.Exit(2)
			}
		}

		fmt.Fprintf(os.Stderr, "failed to run go build: %v", err)
		os.Exit(2)
	}
}

func main() {
	build(os.Args[len(os.Args)-1])

	bin := filepath.Join(homeDir(), "/go/bin/golangci-lint")

	// hand off to the real golangci-lint
	// TODO(theckman): consider importing golangci-lint directly and invoking their library code
	//                 their package main is tiny!!
	err := syscall.Exec(bin, os.Args, os.Environ()) // #nosec
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to execve golangci-lint: %v", err)
		os.Exit(2)
	}
}
