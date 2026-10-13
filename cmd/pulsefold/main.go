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
	billingPath, inputPath, format, ok := parseArguments(arguments, stderr)
	if !ok {
		return 2
	}
	options := accounting.Options{}
	if billingPath != "" {
		billingFile, err := os.Open(billingPath)
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
	source := input.FileSource{Path: inputPath, Info: input.SourceMetadata{
		Identity: sourceLabel(inputPath), Kind: input.SourceOfflineReplay, Format: input.FormatOTLPJSON,
	}}
	outcome := runner.Analyse(context.Background(), source, input.Config{}, c1Config(), c2Config(), options)
	var err error
	switch format {
	case outputText:
		err = report.WriteText(stdout, outcome.Report)
	case outputJSON:
		err = report.WriteJSON(stdout, outcome.Report)
	}
	if err != nil {
		fmt.Fprintln(stderr, "pulsefold: unable to write report")
		return 1
	}
	if outcome.Err != nil {
		fmt.Fprintln(stderr, "pulsefold: analysis did not complete")
		return 1
	}
	return 0
}

type outputFormat string

const (
	outputText outputFormat = "text"
	outputJSON outputFormat = "json"
)

// parsedArguments keeps the small CLI intentionally dependency-free. The
// billing path is retained only until it is opened and parsed locally.
type parsedArguments struct {
	billingPath string
}

func parseArguments(arguments []string, stderr io.Writer) (string, string, outputFormat, bool) {
	if len(arguments) == 0 || arguments[0] != "analyse" {
		writeUsage(stderr)
		return "", "", "", false
	}
	parsed := parsedArguments{}
	format := outputText
	inputPath := ""
	for index := 1; index < len(arguments); index++ {
		argument := arguments[index]
		switch argument {
		case "--billing":
			if parsed.billingPath != "" || index+1 >= len(arguments) {
				writeUsage(stderr)
				return "", "", "", false
			}
			index++
			parsed.billingPath = arguments[index]
		case "--format":
			if index+1 >= len(arguments) {
				writeUsage(stderr)
				return "", "", "", false
			}
			index++
			format = outputFormat(arguments[index])
			if format != outputText && format != outputJSON {
				writeUsage(stderr)
				return "", "", "", false
			}
		default:
			if inputPath != "" || argument == "" || argument[0] == '-' {
				writeUsage(stderr)
				return "", "", "", false
			}
			inputPath = argument
		}
	}
	if inputPath == "" {
		writeUsage(stderr)
		return "", "", "", false
	}
	return parsed.billingPath, inputPath, format, true
}

func writeUsage(stderr io.Writer) {
	fmt.Fprintln(stderr, "usage: pulsefold analyse <otlp-json-file> [--billing <billing-json-file>] [--format text|json]")
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
