// Command pulsefold runs the bounded V0 offline analyser.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/amahon78/pulsefold/internal/c1"
	"github.com/amahon78/pulsefold/internal/c2"
	"github.com/amahon78/pulsefold/internal/input"
	"github.com/amahon78/pulsefold/internal/report"
	"github.com/amahon78/pulsefold/internal/runner"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(arguments []string, stdout, stderr io.Writer) int {
	if len(arguments) != 2 || arguments[0] != "analyse" {
		fmt.Fprintln(stderr, "usage: pulsefold analyse <otlp-json-file>")
		return 2
	}
	source := input.FileSource{Path: arguments[1], Info: input.SourceMetadata{
		Identity: filepath.Base(arguments[1]), Kind: input.SourceOfflineReplay, Format: input.FormatOTLPJSON,
	}}
	outcome := runner.Analyse(context.Background(), source, input.Config{}, c1Config(), c2Config())
	if err := report.WriteJSON(stdout, outcome.Report); err != nil {
		fmt.Fprintln(stderr, "pulsefold: unable to write report")
		return 1
	}
	if outcome.Err != nil {
		fmt.Fprintln(stderr, "pulsefold: analysis did not complete")
		return 1
	}
	return 0
}

func c1Config() c1.Config {
	return c1.Config{}
}

func c2Config() c2.Config {
	return c2.Config{}
}
