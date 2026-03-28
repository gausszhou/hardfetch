package main

import (
	"bytes"
	"flag"
	"os"
	"strings"
	"testing"
)

func TestMainHelp(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"hardfetch", "--help"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Capture both stdout and stderr
	var stdoutBuf, stderrBuf bytes.Buffer
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	stdoutR, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()
	os.Stdout = stdoutW
	os.Stderr = stderrW

	main()
	stdoutW.Close()
	stderrW.Close()

	os.Stdout = oldStdout
	os.Stderr = oldStderr
	stdoutBuf.ReadFrom(stdoutR)
	stderrBuf.ReadFrom(stderrR)

	stdoutOutput := stdoutBuf.String()
	stderrOutput := stderrBuf.String()

	// Combine outputs for checking
	combinedOutput := stdoutOutput + stderrOutput

	if !strings.Contains(combinedOutput, "Usage:") {
		t.Errorf("Help output doesn't contain 'Usage:'")
	}
}

func TestMainVersion(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"hardfetch", "--version"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()
	w.Close()

	os.Stdout = oldStdout
	buf.ReadFrom(r)
	output := buf.String()

	expected := "hardfetch version 0.1.0\nAuthor: Hardfetch Team\nWebsite: https://github.com/yourusername/hardfetch\n"
	if output != expected {
		t.Errorf("Version output doesn't match expected:\nGot: %q\nWant: %q", output, expected)
	}
}

func TestMainDefault(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"hardfetch"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()
	w.Close()

	os.Stdout = oldStdout
	buf.ReadFrom(r)
	output := buf.String()

	// Default output now shows system info, not just version
	// So we need to update the test
	if !strings.Contains(output, "Hostname") && !strings.Contains(output, "hardfetch version") {
		t.Errorf("Default output doesn't contain expected content: %s", output)
	}
}
