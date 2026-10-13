// Package report defines the deterministic, payload-free V0 JSON report.
package report

import (
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/amahon78/pulsefold/internal/accounting"
	"github.com/amahon78/pulsefold/internal/c1"
	"github.com/amahon78/pulsefold/internal/c2"
	"github.com/amahon78/pulsefold/internal/input"
	"github.com/amahon78/pulsefold/internal/provenance"
)

const SchemaVersion = "pulsefold/v0alpha1"

type Status string

const (
	StatusComplete       Status = "complete"
	StatusInputFailed    Status = "input_failed"
	StatusAnalysisFailed Status = "analysis_failed"
)

// Report is intentionally evidence-oriented metadata, not a telemetry export.
// It contains no record body, attribute value or canonical identity. Its only
// hashes are manifest commitments over metadata and aggregate report content.
type Report struct {
	SchemaVersion          string               `json:"schema_version"`
	AnalyserVersion        string               `json:"analyser_version"`
	Status                 Status               `json:"status"`
	CounterfactualNotice   string               `json:"counterfactual_notice"`
	RedactionPolicyVersion string               `json:"redaction_policy_version,omitempty"`
	Source                 Source               `json:"source"`
	AnalysisInterval       AnalysisInterval     `json:"analysis_interval"`
	Coverage               Coverage             `json:"coverage"`
	Accounting             accounting.Result    `json:"accounting"`
	Limitations            []string             `json:"limitations"`
	Manifest               *provenance.Manifest `json:"manifest,omitempty"`
	C1                     *C1                  `json:"c1,omitempty"`
	C2                     *C2                  `json:"c2,omitempty"`
	Failure                *Failure             `json:"failure,omitempty"`
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
	RecordsObserved      uint64 `json:"records_observed"`
	RecordsObservedKnown bool   `json:"records_observed_known"`
	RecordsAccepted      uint64 `json:"records_accepted"`
	RecordsRejected      uint64 `json:"records_rejected"`
	MalformedRecords     uint64 `json:"malformed_records"`
	MalformedInput       uint64 `json:"malformed_input"`
	UnsupportedRecords   uint64 `json:"unsupported_records"`
	UnsupportedInput     uint64 `json:"unsupported_input"`
	CaptureErrors        uint64 `json:"capture_errors"`
	InputBytesRead       uint64 `json:"input_bytes_read"`
	InputBytesKnown      bool   `json:"input_bytes_known"`
	InputByteLimitHit    bool   `json:"input_byte_limit_hit"`
}

type C1 struct {
	Classification          accounting.Classification `json:"classification"`
	CanonicalisationVersion string                    `json:"canonicalisation_version"`
	CandidatePolicyVersion  string                    `json:"candidate_policy_version"`
	RecordsPresented        uint64                    `json:"records_presented"`
	RecordsAnalysed         uint64                    `json:"records_analysed"`
	UniqueIdentities        uint64                    `json:"unique_identities"`
	DuplicateGroups         uint64                    `json:"duplicate_groups"`
	DuplicateRecords        uint64                    `json:"duplicate_records"`
	CounterfactualRecords   uint64                    `json:"counterfactual_records"`
	InputCanonicalBytes     uint64                    `json:"input_canonical_bytes"`
	CounterfactualBytes     uint64                    `json:"counterfactual_canonical_bytes"`
	RawByteOpportunity      uint64                    `json:"canonical_raw_byte_opportunity"`
	ExcludedRecords         uint64                    `json:"excluded_records"`
	ExcludedOversized       uint64                    `json:"excluded_oversized_identity_records"`
	ExcludedStateCap        uint64                    `json:"excluded_state_cap_records"`
	StateBytes              uint64                    `json:"state_bytes"`
	StateCapReached         bool                      `json:"state_cap_reached"`
}

// C2 reports bounded structural-candidate counts only. It contains neither a
// template string nor extracted parameters, so a report cannot become a
// second telemetry store.
type C2 struct {
	Classification           accounting.Classification `json:"classification"`
	AlgorithmVersion         string                    `json:"algorithm_version"`
	CandidatePolicyVersion   string                    `json:"candidate_policy_version"`
	C1RepresentativeRecords  uint64                    `json:"c1_representative_records"`
	RecordsAnalysed          uint64                    `json:"records_analysed"`
	CandidateRecords         uint64                    `json:"candidate_records"`
	InferredTemplateGroups   uint64                    `json:"inferred_template_groups"`
	SourceTemplateRecords    uint64                    `json:"source_template_records"`
	SourceTemplateGroups     uint64                    `json:"source_template_groups"`
	SourceTemplateExcluded   uint64                    `json:"source_template_excluded_records"`
	AdditionalOpportunity    uint64                    `json:"additional_counterfactual_records"`
	ExcludedRecords          uint64                    `json:"excluded_records"`
	ExcludedUnsupportedBody  uint64                    `json:"excluded_unsupported_body_records"`
	ExcludedOversizedBody    uint64                    `json:"excluded_oversized_body_records"`
	ExcludedTooManyTokens    uint64                    `json:"excluded_too_many_token_records"`
	ExcludedStateCap         uint64                    `json:"excluded_state_cap_records"`
	Partitions               uint64                    `json:"partitions"`
	StateBytes               uint64                    `json:"state_bytes"`
	StateCapReached          bool                      `json:"state_cap_reached"`
	SourceTemplateCapReached bool                      `json:"source_template_cap_reached"`
	SourceTemplateStateCap   bool                      `json:"source_template_state_cap_reached"`
	MaxClusters              int                       `json:"max_clusters"`
	MaxPartitions            int                       `json:"max_partitions"`
	MaxStateBytes            int                       `json:"max_state_bytes"`
	MaxBodyBytes             int                       `json:"max_body_bytes"`
	MaxTemplateBytes         int                       `json:"max_template_bytes"`
	MaxTokens                int                       `json:"max_tokens"`
	MaxSourceTemplates       int                       `json:"max_source_templates"`
	Limitations              []string                  `json:"limitations"`
}

// Failure is a static classification. Error text is intentionally omitted so
// parser errors cannot expose raw telemetry in a report or diagnostic path.
type Failure struct {
	Stage string `json:"stage"`
	Code  string `json:"code"`
}

func Complete(replay input.ReplayResult, policyVersion, analyserVersion string, result c1.Result, c2Result c2.Result, accountingResult accounting.Result) Report {
	report := base(replay, analyserVersion)
	report.Status = StatusComplete
	report.RedactionPolicyVersion = policyVersion
	report.Accounting = accountingResult
	report.C1 = &C1{
		Classification:          accounting.Counterfactual,
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
		Classification:   accounting.Inferred,
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
	report.Limitations = limitations(report)
	attachManifest(&report, replay, policyVersion, &result, &c2Result)
	return report
}

func InputFailure(replay input.ReplayResult, analyserVersion, code string) Report {
	report := base(replay, analyserVersion)
	report.Status = StatusInputFailed
	report.Failure = &Failure{Stage: "input", Code: code}
	report.Limitations = limitations(report)
	attachManifest(&report, replay, "", nil, nil)
	return report
}

func AnalysisFailure(replay input.ReplayResult, policyVersion, analyserVersion, code string) Report {
	report := base(replay, analyserVersion)
	report.Status = StatusAnalysisFailed
	report.RedactionPolicyVersion = policyVersion
	report.Failure = &Failure{Stage: "analysis", Code: code}
	report.Limitations = limitations(report)
	attachManifest(&report, replay, policyVersion, nil, nil)
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
		Accounting:       accounting.FromInput(replay),
	}
}

func coverage(value input.Coverage) Coverage {
	return Coverage{
		RecordsObserved: value.RecordsObserved, RecordsObservedKnown: value.RecordsObservedKnown, RecordsAccepted: value.RecordsAccepted,
		RecordsRejected: value.RecordsRejected, MalformedRecords: value.MalformedRecords, MalformedInput: value.MalformedInput,
		UnsupportedRecords: value.UnsupportedRecords, UnsupportedInput: value.UnsupportedInput,
		CaptureErrors: value.CaptureErrors, InputBytesRead: value.InputBytesRead, InputBytesKnown: value.InputBytesKnown,
		InputByteLimitHit: value.InputByteLimitHit,
	}
}

// limitations returns concise, deterministic statements that constrain the
// interpretation of this particular shadow report. They are report content,
// rather than diagnostics, and are therefore covered by its commitment.
func limitations(value Report) []string {
	items := []string{
		"Shadow analysis only: PulseFold did not modify, suppress, compact or forward telemetry.",
		"The report stores no evidence content and its integrity commitment is not authentication.",
	}
	if value.Status != StatusComplete && value.Failure != nil {
		items = append(items, "Analysis did not complete: "+value.Failure.Stage+"/"+value.Failure.Code+".")
	}
	if !value.Accounting.Coverage.Input.RecordsObservedKnown {
		items = append(items, "Observed-record coverage is unknown: "+value.Accounting.Coverage.Input.AcceptedRecordCoverage.Reason+".")
	}
	if value.Accounting.Coverage.Input.InputByteLimitHit {
		items = append(items, "Input capture reached its byte limit; no opportunity is extrapolated beyond the captured input.")
	}
	if value.C1 != nil && stageIncomplete(value.Accounting.Coverage.C1) {
		items = append(items, "C1 analysis coverage is incomplete; only the reported eligible and analysed counts support its opportunity.")
	}
	if value.C2 != nil && stageIncomplete(value.Accounting.Coverage.C2) {
		items = append(items, "C2 analysis coverage is incomplete; no structural opportunity is extrapolated beyond its analysed representatives.")
	}
	if value.C1 != nil && value.C1.ExcludedRecords > 0 {
		items = append(items, "C1 excluded records outside its bounded analysis state.")
	}
	if value.C2 != nil {
		items = append(items, "C2 is an inferred structural candidate opportunity, not verified or removable telemetry.")
		if value.C2.ExcludedRecords > 0 {
			items = append(items, "C2 excluded records outside its supported body types or bounded analysis state.")
		}
	}
	if !value.Accounting.Wire.Known {
		items = append(items, "Wire bytes are unknown because this offline analyser does not measure transport encoding, batching or compression.")
	}
	if !value.Accounting.Billing.Available {
		items = append(items, "Estimated gross observability cost opportunity is unavailable: "+value.Accounting.Billing.Reason+".")
	} else {
		items = append(items, "Economic estimates use an operator-supplied generic billing model and are not net or realised savings.")
		if value.Accounting.Billing.C1 != nil && value.Accounting.Billing.C2 != nil {
			items = append(items, "C1 and C2 cost opportunities overlap: the C2 potential includes the C1 exact-repetition reduction and must not be added to it.")
		}
	}
	return items
}

func stageIncomplete(value accounting.StageCoverage) bool {
	return !value.Completed || !value.Coverage.Known || value.Analysed < value.Eligible
}

// now is the volatile manifest generation clock. It is excluded from the
// integrity commitment and replaced only by tests.
var now = time.Now

// ErrIntegrityMismatch reports that a report and its manifest commitment do
// not agree. It is integrity checking, not authentication: anyone able to
// modify a report can recompute the commitment.
var ErrIntegrityMismatch = errors.New("report integrity commitment does not match")

// commitmentPayload is the pulsefold/report-content-json-v1 canonical form:
// compact encoding/json output of this schema version's typed structs, with
// fields in declaration order, no maps, UTC RFC 3339 times and the report's
// manifest replaced by its CommitmentView.
type commitmentPayload struct {
	Canonicalisation string              `json:"canonicalisation"`
	Report           Report              `json:"report"`
	Manifest         provenance.Manifest `json:"manifest"`
}

// IntegrityCommitment produces the deterministic V0 report checksum. It
// covers all report content and the deterministic manifest fields. The
// manifest integrity, content-derived ID and creation time are excluded to
// avoid self-reference and wall-clock dependence. value is not modified.
func IntegrityCommitment(value Report) (provenance.Integrity, error) {
	payload := commitmentPayload{Canonicalisation: provenance.IntegrityCanonicalisation, Report: value}
	payload.Report.Manifest = nil
	if value.Manifest != nil {
		payload.Manifest = value.Manifest.CommitmentView()
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return provenance.Integrity{}, err
	}
	return provenance.Commitment(encoded), nil
}

// VerifyIntegrity recomputes the commitment and checks the manifest ID. It
// detects accidental or uncoordinated changes only; it does not prove who
// produced the report.
func VerifyIntegrity(value Report) error {
	if value.Manifest == nil {
		return ErrIntegrityMismatch
	}
	integrity, err := IntegrityCommitment(value)
	if err != nil || integrity != value.Manifest.Integrity {
		return ErrIntegrityMismatch
	}
	if err := value.Manifest.Validate(); err != nil {
		return ErrIntegrityMismatch
	}
	return nil
}

func attachManifest(report *Report, replay input.ReplayResult, policyVersion string, c1Result *c1.Result, c2Result *c2.Result) {
	completion := provenance.Completion{State: string(report.Status)}
	if report.Failure != nil {
		completion.FailureCode = report.Failure.Code
	}
	manifest := provenance.Build(replay, SchemaVersion, report.AnalyserVersion, policyVersion, completion, c1Result, c2Result, report.Accounting)
	report.Manifest = &manifest
	integrity, err := IntegrityCommitment(*report)
	if err != nil {
		// Report contains only typed JSON values, so this is unreachable for
		// this schema. Do not emit a manifest that claims a commitment.
		report.Manifest = nil
		return
	}
	provenance.AttachIntegrity(&manifest, integrity)
	manifest.CreatedAt = now().UTC()
}
