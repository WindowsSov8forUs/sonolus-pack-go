package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestParseArgsDefaults(t *testing.T) {
	options, err := parseArgs(nil, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if options.input != "source" || options.output != "pack" {
		t.Fatalf("options = %#v; want source/pack", options)
	}
}

func TestParseArgsShortFlags(t *testing.T) {
	options, err := parseArgs([]string{"-i", "src", "-o", "out"}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if options.input != "src" || options.output != "out" {
		t.Fatalf("options = %#v; want src/out", options)
	}
}

func TestParseArgsLongFlags(t *testing.T) {
	options, err := parseArgs([]string{"--input", "src", "--output", "out"}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if options.input != "src" || options.output != "out" {
		t.Fatalf("options = %#v; want src/out", options)
	}
}

func TestRunVersion(t *testing.T) {
	oldVersion := version
	version = "v1.2.3"
	t.Cleanup(func() { version = oldVersion })

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"--version"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d; want 0", code)
	}
	if got := stdout.String(); got != "v1.2.3\n" {
		t.Fatalf("stdout = %q; want version", got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q; want empty", stderr.String())
	}
}

func TestRunReturnsFailureAndRemovesOutput(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "pack")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"--input", filepath.Join(dir, "missing"), "--output", output}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("exit code = %d; want 1", code)
	}
	if !bytes.Contains(stderr.Bytes(), []byte("[FAILED]")) {
		t.Fatalf("stderr = %q; want [FAILED]", stderr.String())
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("output should be removed, stat err: %v", err)
	}
}
