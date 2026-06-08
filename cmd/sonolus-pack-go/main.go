package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/WindowsSov8forUs/sonolus-pack-go/packer"
	"github.com/WindowsSov8forUs/sonolus-pack-go/schema"
)

var version = "dev"

type logger struct {
	w io.Writer
}

func (l logger) Info(args ...any) {
	fmt.Fprintln(l.w, append([]any{"[INFO]"}, args...)...)
}

func (l logger) Warning(args ...any) {
	fmt.Fprintln(l.w, append([]any{"[WARNING]"}, args...)...)
}

type cliOptions struct {
	input   string
	output  string
	version bool
}

func parseArgs(args []string, stderr io.Writer) (cliOptions, error) {
	fs := flag.NewFlagSet("sonolus-pack", flag.ContinueOnError)
	fs.SetOutput(stderr)

	options := cliOptions{input: "source", output: "pack"}
	fs.StringVar(&options.input, "input", options.input, "input directory")
	fs.StringVar(&options.input, "i", options.input, "input directory")
	fs.StringVar(&options.output, "output", options.output, "output directory")
	fs.StringVar(&options.output, "o", options.output, "output directory")
	fs.BoolVar(&options.version, "version", false, "print version")
	fs.BoolVar(&options.version, "V", false, "print version")

	if err := fs.Parse(args); err != nil {
		return cliOptions{}, err
	}
	return options, nil
}

func run(args []string, stdout, stderr io.Writer) int {
	options, err := parseArgs(args, stderr)
	if err != nil {
		return 2
	}
	if options.version {
		fmt.Fprintln(stdout, version)
		return 0
	}

	fmt.Fprintln(stdout, "[INFO]", "Packing:", displayPath(options.input))
	fmt.Fprintln(stdout)

	if err := packer.Pack(context.Background(), packer.Options{
		Input:  options.input,
		Output: options.output,
		Logger: logger{w: stdout},
	}); err != nil {
		var validationErrs *schema.ValidationErrors
		var validationErr *schema.ValidationError
		fmt.Fprintln(stdout)
		if errors.As(err, &validationErrs) {
			for _, item := range validationErrs.Items {
				printValidationError(stderr, item)
			}
		} else if errors.As(err, &validationErr) {
			printValidationError(stderr, *validationErr)
		}
		fmt.Fprintln(stderr, "[FAILED]", err)
		return 1
	}

	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "[SUCCESS]", "Packed to:", options.output)
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func printValidationError(stderr io.Writer, err schema.ValidationError) {
	fmt.Fprintf(stderr, "[ERROR] %s: %s, got %s (%s)\n", err.Path, err.Message, validationValue(err.Value), err.JSONPath)
}

func validationValue(value json.RawMessage) string {
	if len(value) == 0 {
		return "undefined"
	}
	return string(value)
}

func displayPath(path string) string {
	return filepath.ToSlash(path)
}
