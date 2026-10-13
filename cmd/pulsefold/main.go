// Command pulsefold runs the bounded V0 offline analyser.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/amahon78/pulsefold/internal/accounting"
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
	if (len(arguments) != 2 && len(arguments) != 4) || arguments[0] != "analyse" || (len(arguments) == 4 && arguments[2] != "--billing") {
		fmt.Fprintln(stderr, "usage: pulsefold analyse <otlp-json-file> [--billing <billing-json-file>]")
		return 2
	}
	options := accounting.Options{}
	if len(arguments) == 4 {
		billingFile, err := os.Open(arguments[3])
		if err != nil {
			fmt.Fprintln(stderr, "pulsefold: unable to read billing configuration")
			return 2
		}
		model, err := accounting.LoadModelJSON(billingFile)
		closeErr := billingFile.Close()
		if err != nil || closeErr != nil {
			fmt.Fprintln(stderr, "pulsefold: billing configuration is invalid")
			return 2
		}
		options.BillingModel = &model
	}
	source := input.FileSource{Path: arguments[1], Info: input.SourceMetadata{
		Identity: sourceLabel(arguments[1]), Kind: input.SourceOfflineReplay, Format: input.FormatOTLPJSON,
	}}
	outcome := runner.Analyse(context.Background(), source, input.Config{}, c1Config(), c2Config(), options)
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

// fallbackSourceLabel replaces a file name that is not a safe report label.
// The CLI never echoes a directory, and an unsafe base name such as one with
// whitespace, credentials or a Windows path does not fail the analysis.
const fallbackSourceLabel = "offline-replay-file"

func sourceLabel(path string) string {
	label := filepath.Base(path)
	if !input.SafeSourceLabel(label) {
		return fallbackSourceLabel
	}
	return label
}

func c1Config() c1.Config {
	return c1.Config{}
}

func c2Config() c2.Config {
	return c2.Config{}
}
