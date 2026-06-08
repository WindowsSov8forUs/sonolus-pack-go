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

	for _, arg := range []string{"--version", "-V"} {
		t.Run(arg, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := run([]string{arg}, &stdout, &stderr)

			if code != 0 {
				t.Fatalf("exit code = %d; want 0", code)
			}
			if got := stdout.String(); got != "v1.2.3\n" {
				t.Fatalf("stdout = %q; want version", got)
			}
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q; want empty", stderr.String())
			}
		})
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

func TestRunPrintsValidationErrorBeforeFailed(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source")
	output := filepath.Join(dir, "pack")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "info.json"), []byte(`{"title":{"en":1}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"--input", source, "--output", output}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("exit code = %d; want 1", code)
	}
	got := stderr.String()
	errorIndex := bytes.Index([]byte(got), []byte("[ERROR]"))
	failedIndex := bytes.Index([]byte(got), []byte("[FAILED]"))
	if errorIndex < 0 {
		t.Fatalf("stderr = %q; want [ERROR]", got)
	}
	if failedIndex < 0 {
		t.Fatalf("stderr = %q; want [FAILED]", got)
	}
	if errorIndex > failedIndex {
		t.Fatalf("stderr = %q; want [ERROR] before [FAILED]", got)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("output should be removed, stat err: %v", err)
	}
}
