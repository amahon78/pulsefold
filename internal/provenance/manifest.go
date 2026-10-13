// Package provenance defines the small, payload-free V0 evidence manifest.
// It records how a report was produced, not the telemetry content analysed.
package provenance

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/amahon78/pulsefold/internal/accounting"
	"github.com/amahon78/pulsefold/internal/c1"
	"github.com/amahon78/pulsefold/internal/c2"
	"github.com/amahon78/pulsefold/internal/input"
)

const (
	SchemaVersion                 = "pulsefold/evidence-manifest/v0alpha1"
	IntegrityAlgorithm            = "sha-256"
	IntegrityCanonicalisation     = "pulsefold/report-content-json-v1"
	IntegrityScope                = "report_and_manifest_excluding_integrity_manifest_id_created_at"
	contentLifecycleSchemaVersion = "pulsefold/content-lifecycle/v0alpha1"
)

// Origin describes the trust boundary of metadata. It does not claim that
// every field of a telemetry-derived observation is attacker-controlled. V0
// has no field-level taint tracking.
type Origin string

const (
	OriginSystemGenerated  Origin = "system_generated"
	OriginOperatorSupplied Origin = "operator_supplied"
	OriginTelemetryDerived Origin = "telemetry_derived"
	OriginUnknown          Origin = "unknown"
)

// ContentState describes whether separately managed evidence content can be
// retrieved. V0 emits not_stored only and implements no content store.
type ContentState string

const (
	ContentNotStored ContentState = "not_stored"
	ContentAvailable ContentState = "available"
	ContentExpired   ContentState = "expired"
	ContentErased    ContentState = "erased"
)

// Completion describes the state of the report-producing analysis run.
type Completion struct {
	State       string `json:"state"`
	FailureCode string `json:"failure_code,omitempty"`
}

// Source identifies the source contract without carrying its raw operator
// labels. References are opaque SHA-256 commitments to metadata labels, not
// telemetry content commitments.
type Source struct {
	IdentityReference  string `json:"identity_reference"`
	PartitionReference string `json:"partition_reference,omitempty"`
	Kind               string `json:"kind"`
	Format             string `json:"format"`
	NonAuthoritative   bool   `json:"non_authoritative"`
	Origin             Origin `json:"origin"`
}

// Interval holds the source-observed analysis range already present in the
// replay result. It never substitutes local processing time.
type Interval struct {
	Known bool       `json:"known"`
	Start *time.Time `json:"start,omitempty"`
	End   *time.Time `json:"end,omitempty"`
}

// Bounds captures only C2 limits that affect whether structural results were
// produced. It deliberately excludes templates, parameters and record data.
type Bounds struct {
	MaxClusters        int `json:"max_clusters"`
	MaxPartitions      int `json:"max_partitions"`
	MaxStateBytes      int `json:"max_state_bytes"`
	MaxBodyBytes       int `json:"max_body_bytes"`
	MaxTemplateBytes   int `json:"max_template_bytes"`
	MaxTokens          int `json:"max_tokens"`
	MaxSourceTemplates int `json:"max_source_templates"`
}

// Lineage describes the fixed V0 pipeline rather than a speculative generic
// graph. References are logical stage names scoped to this manifest.
type Lineage struct {
	Capture    CaptureStage    `json:"capture"`
	Redaction  RedactionStage  `json:"redaction"`
	C1         C1Stage         `json:"c1"`
	C2         C2Stage         `json:"c2"`
	Accounting AccountingStage `json:"accounting"`
	Report     ReportStage     `json:"report"`
}

type CaptureStage struct {
	Reference      string                    `json:"reference"`
	Origin         Origin                    `json:"origin"`
	Classification accounting.Classification `json:"classification"`
	Format         string                    `json:"format"`
}

type RedactionStage struct {
	Reference     string `json:"reference"`
	InputRef      string `json:"input_ref"`
	Origin        Origin `json:"origin"`
	PolicyVersion string `json:"policy_version,omitempty"`
}

type C1Stage struct {
	Reference               string                    `json:"reference"`
	InputRef                string                    `json:"input_ref"`
	Origin                  Origin                    `json:"origin"`
	Classification          accounting.Classification `json:"classification,omitempty"`
	CanonicalisationVersion string                    `json:"canonicalisation_version,omitempty"`
	CandidatePolicyVersion  string                    `json:"candidate_policy_version,omitempty"`
}

type C2Stage struct {
	Reference              string                    `json:"reference"`
	InputRef               string                    `json:"input_ref"`
	Origin                 Origin                    `json:"origin"`
	Classification         accounting.Classification `json:"classification,omitempty"`
	AlgorithmVersion       string                    `json:"algorithm_version,omitempty"`
	CandidatePolicyVersion string                    `json:"candidate_policy_version,omitempty"`
	PartitionPolicyVersion string                    `json:"partition_policy_version"`
	Bounds                 Bounds                    `json:"bounds"`
}

type AccountingStage struct {
	Reference              string                    `json:"reference"`
	InputRefs              []string                  `json:"input_refs"`
	Origin                 Origin                    `json:"origin"`
	PolicyVersion          string                    `json:"policy_version"`
	BillingOrigin          Origin                    `json:"billing_origin"`
	BillingClassification  accounting.Classification `json:"billing_classification"`
	BillingModelSchema     string                    `json:"billing_model_schema_version,omitempty"`
	BillingModelCommitment string                    `json:"billing_model_commitment,omitempty"`
	WireClassification     accounting.Classification `json:"wire_classification"`
}

type ReportStage struct {
	Reference string   `json:"reference"`
	InputRefs []string `json:"input_refs"`
	Origin    Origin   `json:"origin"`
}

// ContentLifecycle is the initial content state represented by the immutable
// manifest. Later expiry or erasure is represented by a separate immutable
// LifecycleRecord, so provenance does not need rewriting.
type ContentLifecycle struct {
	State ContentState `json:"state"`
}

// Integrity is a content-consistency checksum. It provides neither signer
// identity nor authenticity, and V0 does not implement signatures or PKI.
type Integrity struct {
	Algorithm        string `json:"algorithm"`
	Canonicalisation string `json:"canonicalisation"`
	Scope            string `json:"scope"`
	Commitment       string `json:"commitment"`
	Authenticated    bool   `json:"authenticated"`
}

// Manifest is immutable provenance for one machine-readable report. It has
// no raw telemetry, content locator, representative sample or template text.
type Manifest struct {
	SchemaVersion       string           `json:"schema_version"`
	ManifestID          string           `json:"manifest_id"`
	ReportSchemaVersion string           `json:"report_schema_version"`
	AnalyserVersion     string           `json:"analyser_version"`
	Completion          Completion       `json:"completion"`
	Source              Source           `json:"source"`
	AnalysisInterval    Interval         `json:"analysis_interval"`
	Coverage            CoverageState    `json:"coverage"`
	Content             ContentLifecycle `json:"content"`
	Lineage             Lineage          `json:"lineage"`
	Integrity           Integrity        `json:"integrity"`
	// CreatedAt is volatile local generation time. It is excluded from the
	// integrity commitment and the content-derived ManifestID.
	CreatedAt time.Time `json:"created_at"`
}

// CoverageState summarises the knownness of each coverage denominator in the
// report. The full counts remain in the report accounting object, which the
// integrity commitment covers.
type CoverageState struct {
	InputBytesKnown           bool `json:"input_bytes_known"`
	InputRecordsObservedKnown bool `json:"input_records_observed_known"`
	InputByteLimitHit         bool `json:"input_byte_limit_hit"`
	C1Completed               bool `json:"c1_completed"`
	C1CoverageKnown           bool `json:"c1_coverage_known"`
	C2Completed               bool `json:"c2_completed"`
	C2CoverageKnown           bool `json:"c2_coverage_known"`
	WireKnown                 bool `json:"wire_known"`
}

// LifecycleRecord is an append-only statement about separately managed
// content. It contains no content and does not alter the manifest integrity.
type LifecycleRecord struct {
	SchemaVersion    string       `json:"schema_version"`
	ManifestID       string       `json:"manifest_id"`
	State            ContentState `json:"state"`
	OccurredAt       time.Time    `json:"occurred_at"`
	Reason           string       `json:"reason,omitempty"`
	ContentReference string       `json:"content_reference,omitempty"`
}

// Build creates a manifest before its report integrity value is known. Call
// AttachIntegrity after building the report's non-manifest content.
func Build(replay input.ReplayResult, reportSchemaVersion, analyserVersion, redactionPolicy string, completion Completion, c1Result *c1.Result, c2Result *c2.Result, accountingResult accounting.Result) Manifest {
	manifest := Manifest{
		SchemaVersion: SchemaVersion, ReportSchemaVersion: reportSchemaVersion, AnalyserVersion: analyserVersion,
		Completion: completion, Source: source(replay.Source), AnalysisInterval: interval(replay.Coverage.Interval),
		Coverage: coverageState(accountingResult),
		Content:  ContentLifecycle{State: ContentNotStored},
		Lineage: Lineage{
			Capture:    CaptureStage{Reference: "capture", Origin: OriginTelemetryDerived, Classification: accounting.Measured, Format: replay.Source.Format},
			Redaction:  RedactionStage{Reference: "redaction", InputRef: "capture", Origin: OriginOperatorSupplied, PolicyVersion: redactionPolicy},
			C1:         C1Stage{Reference: "c1", InputRef: "redaction", Origin: OriginTelemetryDerived},
			C2:         C2Stage{Reference: "c2", InputRef: "c1", Origin: OriginTelemetryDerived, PartitionPolicyVersion: c2.PartitionPolicyVersion},
			Accounting: AccountingStage{Reference: "accounting", InputRefs: []string{"capture", "c1", "c2"}, Origin: OriginTelemetryDerived, PolicyVersion: accounting.PolicyVersion, BillingOrigin: billingOrigin(accountingResult.Billing), BillingClassification: accountingResult.Billing.Classification, BillingModelSchema: accountingResult.Billing.ModelSchemaVersion, BillingModelCommitment: accountingResult.Billing.ModelCommitment, WireClassification: accountingResult.Wire.Classification},
			Report:     ReportStage{Reference: "report", InputRefs: []string{"capture", "redaction", "c1", "c2", "accounting"}, Origin: OriginSystemGenerated},
		},
	}
	if c1Result != nil {
		manifest.Lineage.C1.Classification = accounting.Counterfactual
		manifest.Lineage.C1.CanonicalisationVersion = c1Result.CanonicalisationVersion
		manifest.Lineage.C1.CandidatePolicyVersion = c1Result.CandidatePolicyVersion
	}
	if c2Result != nil {
		manifest.Lineage.C2.Classification = accounting.Inferred
		manifest.Lineage.C2.AlgorithmVersion = c2Result.AlgorithmVersion
		manifest.Lineage.C2.CandidatePolicyVersion = c2Result.CandidatePolicyVersion
		manifest.Lineage.C2.Bounds = Bounds{MaxClusters: c2Result.MaxClusters, MaxPartitions: c2Result.MaxPartitions, MaxStateBytes: c2Result.MaxStateBytes, MaxBodyBytes: c2Result.MaxBodyBytes, MaxTemplateBytes: c2Result.MaxTemplateBytes, MaxTokens: c2Result.MaxTokens, MaxSourceTemplates: c2Result.MaxSourceTemplates}
	}
	return manifest
}

func coverageState(result accounting.Result) CoverageState {
	return CoverageState{
		InputBytesKnown: result.Coverage.Input.CapturedInputBytesKnown, InputRecordsObservedKnown: result.Coverage.Input.RecordsObservedKnown,
		InputByteLimitHit: result.Coverage.Input.InputByteLimitHit,
		C1Completed:       result.Coverage.C1.Completed, C1CoverageKnown: result.Coverage.C1.Coverage.Known,
		C2Completed: result.Coverage.C2.Completed, C2CoverageKnown: result.Coverage.C2.Coverage.Known,
		WireKnown: result.Wire.Known,
	}
}

// CommitmentView returns the manifest fields covered by the integrity
// commitment. Integrity, ManifestID and CreatedAt are cleared so the
// commitment is neither self-referential nor dependent on wall-clock time.
// The returned value shares no mutable memory with manifest.
func (manifest Manifest) CommitmentView() Manifest {
	view := manifest.Clone()
	view.Integrity = Integrity{}
	view.ManifestID = ""
	view.CreatedAt = time.Time{}
	return view
}

// Clone returns a deep copy. Manifest contains pointer and slice fields, so
// a plain struct copy would alias them.
func (manifest Manifest) Clone() Manifest {
	clone := manifest
	if manifest.AnalysisInterval.Start != nil {
		start := *manifest.AnalysisInterval.Start
		clone.AnalysisInterval.Start = &start
	}
	if manifest.AnalysisInterval.End != nil {
		end := *manifest.AnalysisInterval.End
		clone.AnalysisInterval.End = &end
	}
	clone.Lineage.Accounting.InputRefs = append([]string(nil), manifest.Lineage.Accounting.InputRefs...)
	clone.Lineage.Report.InputRefs = append([]string(nil), manifest.Lineage.Report.InputRefs...)
	return clone
}

// AttachIntegrity records the deterministic report-content commitment and
// derives the manifest's content-derived identity. It is not authentication.
func AttachIntegrity(manifest *Manifest, integrity Integrity) {
	manifest.Integrity = integrity
	manifest.ManifestID = manifestID(integrity.Commitment)
}

// Commitment returns a standard SHA-256 commitment over already canonical
// report bytes. It is deliberately generic so report owns its hash scope.
func Commitment(payload []byte) Integrity {
	sum := sha256.Sum256(payload)
	return Integrity{Algorithm: IntegrityAlgorithm, Canonicalisation: IntegrityCanonicalisation, Scope: IntegrityScope, Commitment: "sha256:" + hex.EncodeToString(sum[:]), Authenticated: false}
}

// NewLifecycleRecord validates a later lifecycle declaration. V0 itself only
// emits not_stored in manifests and does not resolve ContentReference values.
func NewLifecycleRecord(manifestID string, state ContentState, occurredAt time.Time, reason, contentReference string) (LifecycleRecord, error) {
	record := LifecycleRecord{SchemaVersion: contentLifecycleSchemaVersion, ManifestID: manifestID, State: state, OccurredAt: occurredAt.UTC(), Reason: reason, ContentReference: contentReference}
	if err := record.Validate(); err != nil {
		return LifecycleRecord{}, err
	}
	return record, nil
}

// Retrievable is false for every unavailable state. It does not attempt a
// lookup, and an available record does not imply a V0 content store exists.
func (record LifecycleRecord) Retrievable() bool {
	return record.State == ContentAvailable && record.ContentReference != ""
}

func (manifest Manifest) Validate() error {
	if manifest.SchemaVersion != SchemaVersion || manifest.ManifestID == "" || manifest.ReportSchemaVersion == "" || manifest.AnalyserVersion == "" {
		return errors.New("invalid evidence manifest")
	}
	if err := validContentState(manifest.Content.State); err != nil {
		return err
	}
	if err := validIntegrity(manifest.Integrity); err != nil {
		return err
	}
	if manifest.Source.IdentityReference == "" || manifest.Source.Origin != OriginOperatorSupplied {
		return errors.New("invalid manifest source")
	}
	if manifest.ManifestID != manifestID(manifest.Integrity.Commitment) || manifest.CreatedAt.IsZero() {
		return errors.New("invalid manifest identity")
	}
	switch manifest.Completion.State {
	case "complete", "input_failed", "analysis_failed":
	default:
		return errors.New("invalid manifest completion")
	}
	if (manifest.Completion.State == "complete") != (manifest.Completion.FailureCode == "") {
		return errors.New("invalid manifest completion")
	}
	lineage := manifest.Lineage
	if lineage.Capture.Reference != "capture" || lineage.Redaction.Reference != "redaction" || lineage.Redaction.InputRef != "capture" ||
		lineage.C1.Reference != "c1" || lineage.C1.InputRef != "redaction" || lineage.C2.Reference != "c2" || lineage.C2.InputRef != "c1" ||
		lineage.Accounting.Reference != "accounting" || lineage.Report.Reference != "report" {
		return errors.New("invalid manifest lineage")
	}
	return nil
}

func (record LifecycleRecord) Validate() error {
	if record.SchemaVersion != contentLifecycleSchemaVersion || record.ManifestID == "" || record.OccurredAt.IsZero() {
		return errors.New("invalid content lifecycle record")
	}
	if err := validContentState(record.State); err != nil {
		return err
	}
	switch record.State {
	case ContentAvailable:
		if record.ContentReference == "" || record.Reason != "" {
			return errors.New("invalid available content lifecycle record")
		}
	case ContentNotStored:
		if record.ContentReference != "" || record.Reason != "" {
			return errors.New("invalid not-stored content lifecycle record")
		}
	case ContentExpired, ContentErased:
		if record.ContentReference != "" || !validCode(record.Reason) {
			return errors.New("invalid unavailable content lifecycle record")
		}
	}
	if record.ContentReference != "" && !validCode(record.ContentReference) {
		return errors.New("invalid content reference")
	}
	return nil
}

// validCode accepts a short lower-case machine code. Lifecycle reasons and
// opaque content references are not free-form text and cannot carry content.
func validCode(value string) bool {
	if value == "" || len(value) > 64 {
		return false
	}
	for index := 0; index < len(value); index++ {
		character := value[index]
		if (character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') || character == '_' || character == '-' {
			continue
		}
		return false
	}
	return true
}

func source(metadata input.SourceMetadata) Source {
	return Source{IdentityReference: metadataReference("source-identity", metadata.Identity), PartitionReference: optionalReference("source-partition", metadata.Partition), Kind: string(metadata.Kind), Format: metadata.Format, NonAuthoritative: metadata.NonAuthoritative, Origin: OriginOperatorSupplied}
}

func interval(value input.Interval) Interval {
	output := Interval{Known: value.Known}
	if value.Known {
		start, end := value.Start, value.End
		output.Start, output.End = &start, &end
	}
	return output
}

func metadataReference(namespace, value string) string {
	sum := sha256.Sum256([]byte(namespace + "\x00" + value))
	return "ref:" + hex.EncodeToString(sum[:])
}

func optionalReference(namespace, value string) string {
	if value == "" {
		return ""
	}
	return metadataReference(namespace, value)
}

func manifestID(commitment string) string {
	value := strings.TrimPrefix(commitment, "sha256:")
	return "pfm_v0_" + value
}

func billingOrigin(billing accounting.Billing) Origin {
	if billing.ModelSupplied {
		return OriginOperatorSupplied
	}
	return OriginUnknown
}

func validContentState(state ContentState) error {
	switch state {
	case ContentNotStored, ContentAvailable, ContentExpired, ContentErased:
		return nil
	default:
		return errors.New("invalid content state")
	}
}

func validIntegrity(value Integrity) error {
	if value.Algorithm != IntegrityAlgorithm || value.Canonicalisation != IntegrityCanonicalisation || value.Scope != IntegrityScope || value.Authenticated {
		return errors.New("invalid manifest integrity")
	}
	commitment := strings.TrimPrefix(value.Commitment, "sha256:")
	if len(commitment) != sha256.Size*2 || value.Commitment == commitment {
		return errors.New("invalid manifest commitment")
	}
	if _, err := hex.DecodeString(commitment); err != nil {
		return fmt.Errorf("invalid manifest commitment")
	}
	return nil
}
