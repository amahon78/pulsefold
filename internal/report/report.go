// Package report defines the deterministic, payload-free V0 JSON report.
package report

import (
	"encoding/json"
	"io"
	"time"

	"github.com/amahon78/pulsefold/internal/c1"
	"github.com/amahon78/pulsefold/internal/c2"
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
	C2                     *C2              `json:"c2,omitempty"`
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

// C2 reports bounded structural-candidate counts only. It contains neither a
// template string nor extracted parameters, so a report cannot become a
// second telemetry store.
type C2 struct {
	AlgorithmVersion         string   `json:"algorithm_version"`
	CandidatePolicyVersion   string   `json:"candidate_policy_version"`
	C1RepresentativeRecords  uint64   `json:"c1_representative_records"`
	RecordsAnalysed          uint64   `json:"records_analysed"`
	CandidateRecords         uint64   `json:"candidate_records"`
	InferredTemplateGroups   uint64   `json:"inferred_template_groups"`
	SourceTemplateRecords    uint64   `json:"source_template_records"`
	SourceTemplateGroups     uint64   `json:"source_template_groups"`
	SourceTemplateExcluded   uint64   `json:"source_template_excluded_records"`
	AdditionalOpportunity    uint64   `json:"additional_counterfactual_records"`
	ExcludedRecords          uint64   `json:"excluded_records"`
	ExcludedUnsupportedBody  uint64   `json:"excluded_unsupported_body_records"`
	ExcludedOversizedBody    uint64   `json:"excluded_oversized_body_records"`
	ExcludedTooManyTokens    uint64   `json:"excluded_too_many_token_records"`
	ExcludedStateCap         uint64   `json:"excluded_state_cap_records"`
	Partitions               uint64   `json:"partitions"`
	StateBytes               uint64   `json:"state_bytes"`
	StateCapReached          bool     `json:"state_cap_reached"`
	SourceTemplateCapReached bool     `json:"source_template_cap_reached"`
	SourceTemplateStateCap   bool     `json:"source_template_state_cap_reached"`
	MaxClusters              int      `json:"max_clusters"`
	MaxPartitions            int      `json:"max_partitions"`
	MaxStateBytes            int      `json:"max_state_bytes"`
	MaxBodyBytes             int      `json:"max_body_bytes"`
	MaxTemplateBytes         int      `json:"max_template_bytes"`
	MaxTokens                int      `json:"max_tokens"`
	MaxSourceTemplates       int      `json:"max_source_templates"`
	Limitations              []string `json:"limitations"`
}

// Failure is a static classification. Error text is intentionally omitted so
// parser errors cannot expose raw telemetry in a report or diagnostic path.
type Failure struct {
	Stage string `json:"stage"`
	Code  string `json:"code"`
}

func Complete(replay input.ReplayResult, policyVersion, analyserVersion string, result c1.Result, c2Result c2.Result) Report {
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
	report.C2 = &C2{
		AlgorithmVersion: c2Result.AlgorithmVersion, CandidatePolicyVersion: c2Result.CandidatePolicyVersion,
		C1RepresentativeRecords: c2Result.C1RepresentativeRecords, RecordsAnalysed: c2Result.RecordsAnalysed,
		CandidateRecords: c2Result.CandidateRecords, InferredTemplateGroups: c2Result.InferredTemplateGroups,
		SourceTemplateRecords: c2Result.SourceTemplateRecords, SourceTemplateGroups: c2Result.SourceTemplateGroups, SourceTemplateExcluded: c2Result.SourceTemplateExcluded,
		AdditionalOpportunity: c2Result.AdditionalOpportunity, ExcludedRecords: c2Result.ExcludedRecords,
		ExcludedUnsupportedBody: c2Result.ExcludedUnsupportedBody, ExcludedOversizedBody: c2Result.ExcludedOversizedBody,
		ExcludedTooManyTokens: c2Result.ExcludedTooManyTokens, ExcludedStateCap: c2Result.ExcludedStateCap,
		Partitions: c2Result.Partitions, StateBytes: c2Result.StateBytes, StateCapReached: c2Result.StateCapReached,
		SourceTemplateCapReached: c2Result.SourceTemplateCapReached, SourceTemplateStateCap: c2Result.SourceTemplateStateCap, MaxClusters: c2Result.MaxClusters,
		MaxPartitions: c2Result.MaxPartitions, MaxStateBytes: c2Result.MaxStateBytes, MaxBodyBytes: c2Result.MaxBodyBytes, MaxTemplateBytes: c2Result.MaxTemplateBytes,
		MaxTokens: c2Result.MaxTokens, MaxSourceTemplates: c2Result.MaxSourceTemplates,
		Limitations: []string{
			"C2 is report-only and does not modify telemetry or authorise retention decisions.",
			"C2 does not emit template text or extract raw parameters.",
			"C2 candidate groups are bounded to this replay and are not global evidence identities.",
			"Correctness and cross-instance agreement require independently labelled evaluation data.",
		},
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
		CounterfactualNotice: "Each C1/C2 metric is a redacted, report-only counterfactual reduction opportunity, not observed financial saving or a transformation decision.",
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
