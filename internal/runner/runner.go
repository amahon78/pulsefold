// Package runner wires the fixed V0 input, redaction and report-only C1/C2 path.
package runner

import (
	"context"
	"errors"

	"github.com/amahon78/pulsefold/internal/accounting"
	"github.com/amahon78/pulsefold/internal/c1"
	"github.com/amahon78/pulsefold/internal/c2"
	"github.com/amahon78/pulsefold/internal/input"
	"github.com/amahon78/pulsefold/internal/redaction"
	"github.com/amahon78/pulsefold/internal/report"
)

const AnalyserVersion = "devel"

// Outcome always carries a safe machine-readable report. Err determines the
// caller's exit status and must not be printed with raw parser detail.
type Outcome struct {
	Report report.Report
	Err    error
}

func Analyse(ctx context.Context, source input.Source, inputConfig input.Config, c1Config c1.Config, c2Config c2.Config, accountingOptions accounting.Options) Outcome {
	replay, err := input.Replay(ctx, source, inputConfig)
	if err != nil {
		return Outcome{Report: report.InputFailure(replay, AnalyserVersion, inputFailureCode(err)), Err: err}
	}
	redacted := redaction.Apply(replay)
	result, err := c1.Analyse(ctx, redacted, c1Config)
	if err != nil {
		return Outcome{Report: report.AnalysisFailure(replay, redacted.PolicyVersion, AnalyserVersion, analysisFailureCode(err)), Err: err}
	}
	structural, err := c2.Analyse(ctx, redacted, result, c2Config)
	if err != nil {
		return Outcome{Report: report.AnalysisFailure(replay, redacted.PolicyVersion, AnalyserVersion, analysisFailureCode(err)), Err: err}
	}
	return Outcome{Report: report.Complete(replay, redacted.PolicyVersion, AnalyserVersion, result, structural, accounting.Build(replay, result, structural, accountingOptions))}
}

func inputFailureCode(err error) string {
	switch {
	case errors.Is(err, input.ErrMalformedInput):
		return "malformed_input"
	case errors.Is(err, input.ErrUnsupportedInput):
		return "unsupported_input"
	case errors.Is(err, input.ErrInputTooLarge):
		return "input_byte_limit"
	case errors.Is(err, context.Canceled):
		return "cancelled"
	default:
		return "capture_failed"
	}
}

func analysisFailureCode(err error) string {
	if errors.Is(err, c1.ErrCancelled) || errors.Is(err, c2.ErrCancelled) || errors.Is(err, context.Canceled) {
		return "cancelled"
	}
	return "analysis_failed"
}
