// Package report defines the deterministic, payload-free V0 JSON report.
package report

import (
	"encoding/json"
	"io"
	"time"

	"github.com/amahon78/pulsefold/internal/c1"
	"github.com/amahon78/pulsefold/internal/input"
)

const SchemaVersion = "pulsefold/v0alpha1"

type Status string

const (
	StatusComplete       Status = "complete"
	StatusInputFailed    Status = "input_failed"
	StatusAnalysisFailed Status = "analysis_failed"
)

// Report is intentionally evidence-oriented metadata, not a telemetry export.
// It contains no record body, attribute value, canonical identity or hash.
type Report struct {
	SchemaVersion          string           `json:"schema_version"`
	AnalyserVersion        string           `json:"analyser_version"`
	Status                 Status           `json:"status"`
	CounterfactualNotice   string           `json:"counterfactual_notice"`
	RedactionPolicyVersion string           `json:"redaction_policy_version,omitempty"`
	Source                 Source           `json:"source"`
	AnalysisInterval       AnalysisInterval `json:"analysis_interval"`
	Coverage               Coverage         `json:"coverage"`
	C1                     *C1              `json:"c1,omitempty"`
	Failure                *Failure         `json:"failure,omitempty"`
}

type Source struct {
	Identity         string `json:"identity"`
	Partition        string `json:"partition,omitempty"`
	Kind             string `json:"kind"`
	Format           string `json:"format"`
	NonAuthoritative bool   `json:"non_authoritative"`
}

type AnalysisInterval struct {
	Known bool       `json:"known"`
	Start *time.Time `json:"start,omitempty"`
	End   *time.Time `json:"end,omitempty"`
}

type Coverage struct {
	RecordsObserved    uint64 `json:"records_observed"`
	RecordsAccepted    uint64 `json:"records_accepted"`
	RecordsRejected    uint64 `json:"records_rejected"`
	MalformedRecords   uint64 `json:"malformed_records"`
	MalformedInput     uint64 `json:"malformed_input"`
	UnsupportedRecords uint64 `json:"unsupported_records"`
	UnsupportedInput   uint64 `json:"unsupported_input"`
	CaptureErrors      uint64 `json:"capture_errors"`
	InputBytesRead     uint64 `json:"input_bytes_read"`
	InputByteLimitHit  bool   `json:"input_byte_limit_hit"`
}

type C1 struct {
	CanonicalisationVersion string `json:"canonicalisation_version"`
	CandidatePolicyVersion  string `json:"candidate_policy_version"`
	RecordsPresented        uint64 `json:"records_presented"`
	RecordsAnalysed         uint64 `json:"records_analysed"`
	UniqueIdentities        uint64 `json:"unique_identities"`
	DuplicateGroups         uint64 `json:"duplicate_groups"`
	DuplicateRecords        uint64 `json:"duplicate_records"`
	CounterfactualRecords   uint64 `json:"counterfactual_records"`
	InputCanonicalBytes     uint64 `json:"input_canonical_bytes"`
	CounterfactualBytes     uint64 `json:"counterfactual_canonical_bytes"`
	RawByteOpportunity      uint64 `json:"canonical_raw_byte_opportunity"`
	ExcludedRecords         uint64 `json:"excluded_records"`
	ExcludedOversized       uint64 `json:"excluded_oversized_identity_records"`
	ExcludedStateCap        uint64 `json:"excluded_state_cap_records"`
	StateBytes              uint64 `json:"state_bytes"`
	StateCapReached         bool   `json:"state_cap_reached"`
}

// Failure is a static classification. Error text is intentionally omitted so
// parser errors cannot expose raw telemetry in a report or diagnostic path.
type Failure struct {
	Stage string `json:"stage"`
	Code  string `json:"code"`
}

func Complete(replay input.ReplayResult, policyVersion, analyserVersion string, result c1.Result) Report {
	report := base(replay, analyserVersion)
	report.Status = StatusComplete
	report.RedactionPolicyVersion = policyVersion
	report.C1 = &C1{
		CanonicalisationVersion: result.CanonicalisationVersion,
		CandidatePolicyVersion:  result.CandidatePolicyVersion,
		RecordsPresented:        result.RecordsPresented,
		RecordsAnalysed:         result.RecordsAnalysed,
		UniqueIdentities:        result.UniqueIdentities,
		DuplicateGroups:         result.DuplicateGroups,
		DuplicateRecords:        result.DuplicateRecords,
		CounterfactualRecords:   result.CounterfactualRecords,
		InputCanonicalBytes:     result.InputCanonicalBytes,
		CounterfactualBytes:     result.CounterfactualBytes,
		RawByteOpportunity:      result.RawByteOpportunity,
		ExcludedRecords:         result.ExcludedRecords,
		ExcludedOversized:       result.ExcludedOversized,
		ExcludedStateCap:        result.ExcludedStateCap,
		StateBytes:              result.StateBytes,
		StateCapReached:         result.StateCapReached,
	}
	return report
}

func InputFailure(replay input.ReplayResult, analyserVersion, code string) Report {
	report := base(replay, analyserVersion)
	report.Status = StatusInputFailed
	report.Failure = &Failure{Stage: "input", Code: code}
	return report
}

func AnalysisFailure(replay input.ReplayResult, policyVersion, analyserVersion, code string) Report {
	report := base(replay, analyserVersion)
	report.Status = StatusAnalysisFailed
	report.RedactionPolicyVersion = policyVersion
	report.Failure = &Failure{Stage: "analysis", Code: code}
	return report
}

func WriteJSON(writer io.Writer, report Report) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

func base(replay input.ReplayResult, analyserVersion string) Report {
	interval := replay.Coverage.Interval
	outputInterval := AnalysisInterval{Known: interval.Known}
	if interval.Known {
		start, end := interval.Start, interval.End
		outputInterval.Start, outputInterval.End = &start, &end
	}
	return Report{
		SchemaVersion:        SchemaVersion,
		AnalyserVersion:      analyserVersion,
		CounterfactualNotice: "C1 metrics are a counterfactual reduction opportunity, not observed financial saving.",
		Source: Source{
			Identity: replay.Source.Identity, Partition: replay.Source.Partition,
			Kind: string(replay.Source.Kind), Format: replay.Source.Format,
			NonAuthoritative: replay.Source.NonAuthoritative,
		},
		AnalysisInterval: outputInterval,
		Coverage:         coverage(replay.Coverage),
	}
}

func coverage(value input.Coverage) Coverage {
	return Coverage{
		RecordsObserved: value.RecordsObserved, RecordsAccepted: value.RecordsAccepted,
		RecordsRejected: value.RecordsRejected, MalformedRecords: value.MalformedRecords, MalformedInput: value.MalformedInput,
		UnsupportedRecords: value.UnsupportedRecords, UnsupportedInput: value.UnsupportedInput,
		CaptureErrors: value.CaptureErrors, InputBytesRead: value.InputBytesRead,
		InputByteLimitHit: value.InputByteLimitHit,
	}
}
