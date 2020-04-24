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
// returns the lines to output (to be joined by newlines)
func cleanOutput(r io.Reader) []string {
	scanner := bufio.NewScanner(r)

	var lines []string

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

		lines = append(lines, strings.TrimPrefix(t, "./"))
	}

	return lines
}

// go build
func build(path string) (lines []string, buildFailed bool, err error) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		return nil, false, fmt.Errorf("failed to get go binary path: %w", err)
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
				return cleanOutput(buf), true, nil
			case 0:
				return nil, false, nil
			default:
				return cleanOutput(buf), false, err
			}
		}

		return nil, false, err
	}

	return nil, false, nil
}

func main() {
	output, failed, err := build(os.Args[len(os.Args)-1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to invoke go build: %v", err)
		os.Exit(2)
	}

	// go build experienced compilation failures
	// treat it like a linter failure
	if failed {
		fmt.Println(strings.Join(output, "\n"))
		os.Exit(1)
	}

	// hand off to the real golangci-lint
	// TODO(theckman): consider importing golangci-lint directly and invoking their library code
	//                 their package main is tiny!!
	err = syscall.Exec(filepath.Join(homeDir(), "/go/bin/golangci-lint"), os.Args, os.Environ()) // #nosec
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to execve golangci-lint: %v", err)
		os.Exit(2)
	}
}
