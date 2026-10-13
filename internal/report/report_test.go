package report

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/amahon78/pulsefold/internal/accounting"
	"github.com/amahon78/pulsefold/internal/c1"
	"github.com/amahon78/pulsefold/internal/c2"
	"github.com/amahon78/pulsefold/internal/fixture"
	"github.com/amahon78/pulsefold/internal/input"
	"github.com/amahon78/pulsefold/internal/provenance"
	"github.com/amahon78/pulsefold/internal/redaction"
	"google.golang.org/protobuf/encoding/protojson"
)

type pipeline struct {
	replay input.ReplayResult
	c1     c1.Result
	c2     c2.Result
	policy string
}

func runPipeline(t testing.TB, seed int64, label string) pipeline {
	t.Helper()
	scenario := fixture.Generate(fixture.Config{Seed: seed, HighCardinalityCount: int(seed%5) + 1, LargeBodyBytes: 64})
	payload, err := protojson.Marshal(scenario.Request)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := input.Replay(context.Background(), input.ReaderSource{Reader: bytes.NewReader(payload), Info: input.SourceMetadata{Identity: label, Partition: "synthetic", Kind: input.SourceOfflineReplay, Format: input.FormatOTLPJSON}}, input.Config{})
	if err != nil {
		t.Fatal(err)
	}
	batch := redaction.Apply(replay)
	c1Result, err := c1.Analyse(context.Background(), batch, c1.Config{})
	if err != nil {
		t.Fatal(err)
	}
	c2Result, err := c2.Analyse(context.Background(), batch, c1Result, c2.Config{})
	if err != nil {
		t.Fatal(err)
	}
	return pipeline{replay: replay, c1: c1Result, c2: c2Result, policy: batch.PolicyVersion}
}

func (p pipeline) complete(options accounting.Options) Report {
	return Complete(p.replay, p.policy, "test", p.c1, p.c2, accounting.Build(p.replay, p.c1, p.c2, options))
}

func fixedClock(t *testing.T, at time.Time) {
	t.Helper()
	previous := now
	now = func() time.Time { return at }
	t.Cleanup(func() { now = previous })
}

func testBillingModel() *accounting.Model {
	return &accounting.Model{SchemaVersion: accounting.BillingSchemaVersion, Currency: "TST", Metric: accounting.MetricAcceptedRecords, IndexedFraction: accounting.Fraction{Numerator: 1, Denominator: 1}, Pricing: accounting.Pricing{Fixed: &accounting.Rate{UnitBlock: 1, MinorUnitsPerBlock: 1}}}
}

func TestCompleteManifestCoversRequiredProvenance(t *testing.T) {
	for _, withBilling := range []bool{false, true} {
		options := accounting.Options{}
		if withBilling {
			options.BillingModel = testBillingModel()
		}
		value := runPipeline(t, 21, "fixture-source").complete(options)
		manifest := value.Manifest
		if manifest == nil || VerifyIntegrity(value) != nil {
			t.Fatalf("billing=%v: manifest missing or unverifiable", withBilling)
		}
		lineage := manifest.Lineage
		checks := map[string]bool{
			"versions":         manifest.SchemaVersion == provenance.SchemaVersion && manifest.ReportSchemaVersion == SchemaVersion && manifest.AnalyserVersion == "test",
			"source":           manifest.Source.IdentityReference != "" && manifest.Source.PartitionReference != "" && manifest.Source.Kind == string(input.SourceOfflineReplay) && manifest.Source.Format == input.FormatOTLPJSON,
			"interval":         manifest.AnalysisInterval.Known && manifest.AnalysisInterval.Start != nil,
			"completion":       manifest.Completion == provenance.Completion{State: "complete"},
			"coverage":         manifest.Coverage.InputBytesKnown && manifest.Coverage.C1Completed && manifest.Coverage.C2Completed && !manifest.Coverage.WireKnown,
			"content":          manifest.Content.State == provenance.ContentNotStored,
			"capture":          lineage.Capture.Classification == accounting.Measured && lineage.Capture.Origin == provenance.OriginTelemetryDerived,
			"redaction":        lineage.Redaction.PolicyVersion == redaction.PolicyVersion && lineage.Redaction.InputRef == "capture" && lineage.Redaction.Origin == provenance.OriginOperatorSupplied,
			"c1":               lineage.C1.Classification == accounting.Counterfactual && lineage.C1.CanonicalisationVersion == c1.CanonicalisationVersion && lineage.C1.CandidatePolicyVersion == c1.CandidatePolicyVersion && lineage.C1.InputRef == "redaction",
			"c2":               lineage.C2.Classification == accounting.Inferred && lineage.C2.AlgorithmVersion == c2.AlgorithmVersion && lineage.C2.PartitionPolicyVersion == c2.PartitionPolicyVersion && lineage.C2.Bounds.MaxClusters == c2.DefaultMaxClusters && lineage.C2.InputRef == "c1",
			"accounting":       lineage.Accounting.PolicyVersion == accounting.PolicyVersion && lineage.Accounting.WireClassification == accounting.Unknown,
			"report":           reflect.DeepEqual(lineage.Report.InputRefs, []string{"capture", "redaction", "c1", "c2", "accounting"}),
			"unauthentic":      !manifest.Integrity.Authenticated,
			"created volatile": !manifest.CreatedAt.IsZero(),
		}
		if withBilling {
			checks["billing"] = lineage.Accounting.BillingOrigin == provenance.OriginOperatorSupplied && lineage.Accounting.BillingClassification == accounting.UserSupplied && lineage.Accounting.BillingModelSchema == accounting.BillingSchemaVersion && strings.HasPrefix(lineage.Accounting.BillingModelCommitment, "sha256:")
		} else {
			checks["billing"] = lineage.Accounting.BillingOrigin == provenance.OriginUnknown && lineage.Accounting.BillingClassification == accounting.Unknown && lineage.Accounting.BillingModelCommitment == ""
		}
		for name, ok := range checks {
			if !ok {
				t.Errorf("billing=%v: %s provenance incorrect: %#v", withBilling, name, manifest)
			}
		}
	}
}

func TestFailureAndPartialManifests(t *testing.T) {
	p := runPipeline(t, 22, "fixture-source")
	inputFailed := InputFailure(input.ReplayResult{Source: input.SourceMetadata{Identity: "invalid_source", Kind: input.SourceOfflineReplay, Format: "unsupported"}, Coverage: input.Coverage{UnsupportedInput: 1}}, "test", "unsupported_input")
	analysisFailed := AnalysisFailure(p.replay, p.policy, "test", "cancelled")
	for _, value := range []Report{inputFailed, analysisFailed} {
		if VerifyIntegrity(value) != nil || value.Manifest.Completion.FailureCode == "" || value.Manifest.Lineage.C1.Classification != "" || value.Manifest.Lineage.C2.AlgorithmVersion != "" || value.Manifest.Coverage.C1Completed {
			t.Errorf("failure manifest = %#v", value.Manifest)
		}
	}
	if inputFailed.Manifest.Lineage.Redaction.PolicyVersion != "" || analysisFailed.Manifest.Lineage.Redaction.PolicyVersion != redaction.PolicyVersion {
		t.Error("failure manifests misreport the redaction stage")
	}
	// A partial capture: byte limit reached before the request completed.
	partial, _ := input.Replay(context.Background(), input.ReaderSource{Reader: strings.NewReader(strings.Repeat(" ", 2048)), Info: input.SourceMetadata{Identity: "partial", Kind: input.SourceOfflineReplay, Format: input.FormatOTLPJSON}}, input.Config{MaxInputBytes: 16})
	value := InputFailure(partial, "test", "input_byte_limit")
	if VerifyIntegrity(value) != nil || !value.Manifest.Coverage.InputByteLimitHit || value.Manifest.Coverage.InputRecordsObservedKnown {
		t.Errorf("partial manifest = %#v", value.Manifest.Coverage)
	}
}

func TestIntegrityIsDeterministicExcludesVolatileFieldsAndDetectsChanges(t *testing.T) {
	fixedClock(t, time.Date(2026, 10, 13, 5, 0, 0, 0, time.UTC))
	first := runPipeline(t, 23, "fixture-source").complete(accounting.Options{})
	fixedClock(t, time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC))
	second := runPipeline(t, 23, "fixture-source").complete(accounting.Options{})
	if first.Manifest.Integrity != second.Manifest.Integrity || first.Manifest.ManifestID != second.Manifest.ManifestID || first.Manifest.CreatedAt.Equal(second.Manifest.CreatedAt) {
		t.Fatal("replay commitment depends on created_at or is not deterministic")
	}
	if other := runPipeline(t, 24, "fixture-source").complete(accounting.Options{}); other.Manifest.Integrity == first.Manifest.Integrity {
		// The commitment covers report content (aggregate counts), not raw
		// telemetry. Seeds 23 and 24 differ in high-cardinality record count.
		t.Error("different report content produced the same commitment")
	}
	if relabelled := runPipeline(t, 23, "other-source").complete(accounting.Options{}); relabelled.Manifest.Integrity == first.Manifest.Integrity {
		t.Error("different source produced the same commitment")
	}
	if billed := runPipeline(t, 23, "fixture-source").complete(accounting.Options{BillingModel: testBillingModel()}); billed.Manifest.Integrity == first.Manifest.Integrity {
		t.Error("billing model did not affect the commitment")
	}
	tampers := map[string]func(*Report){
		"report c1":        func(r *Report) { c := *r.C1; c.DuplicateRecords++; r.C1 = &c },
		"report coverage":  func(r *Report) { r.Coverage.RecordsAccepted++ },
		"report notice":    func(r *Report) { r.CounterfactualNotice = "saving" },
		"manifest lineage": func(r *Report) { m := r.Manifest.Clone(); m.Lineage.C2.AlgorithmVersion = "other"; r.Manifest = &m },
		"manifest content": func(r *Report) {
			m := r.Manifest.Clone()
			m.Content.State = provenance.ContentAvailable
			r.Manifest = &m
		},
		"manifest id": func(r *Report) { m := r.Manifest.Clone(); m.ManifestID = "pfm_v0_x"; r.Manifest = &m },
		"no manifest": func(r *Report) { r.Manifest = nil },
	}
	for name, tamper := range tampers {
		value := first
		tamper(&value)
		if VerifyIntegrity(value) == nil {
			t.Errorf("%s: tampering was not detected", name)
		}
	}
	if VerifyIntegrity(first) != nil {
		t.Error("tamper cases mutated the original report")
	}
	// Changing only the volatile creation time keeps the commitment valid.
	value := first
	m := first.Manifest.Clone()
	m.CreatedAt = m.CreatedAt.Add(time.Hour)
	value.Manifest = &m
	if VerifyIntegrity(value) != nil {
		t.Error("created_at is not excluded from the commitment")
	}
}

func TestReportJSONRoundTripPreservesIntegrity(t *testing.T) {
	value := runPipeline(t, 25, "fixture-source").complete(accounting.Options{BillingModel: testBillingModel()})
	var encoded bytes.Buffer
	if err := WriteJSON(&encoded, value); err != nil {
		t.Fatal(err)
	}
	var decoded Report
	if err := json.Unmarshal(encoded.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	if VerifyIntegrity(decoded) != nil {
		t.Error("decoded report no longer verifies")
	}
	manifestJSON, err := json.Marshal(decoded.Manifest)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := provenance.ParseManifestJSON(bytes.NewReader(manifestJSON))
	if err != nil || parsed.ManifestID != value.Manifest.ManifestID {
		t.Errorf("manifest did not round trip: %v", err)
	}
}

func TestReportDoesNotAliasAnalysisInputs(t *testing.T) {
	p := runPipeline(t, 26, "fixture-source")
	options := accounting.Options{BillingModel: testBillingModel()}
	value := p.complete(options)
	before, err := IntegrityCommitment(value)
	if err != nil {
		t.Fatal(err)
	}
	// Mutate every caller-owned input after construction.
	p.c1.DuplicateRecords += 100
	p.c2.MaxClusters++
	p.replay.Source.Identity = "mutated"
	p.replay.Coverage.Interval.Start = p.replay.Coverage.Interval.Start.Add(time.Hour)
	options.BillingModel.Pricing.Fixed.MinorUnitsPerBlock = 99
	options.BillingModel.Currency = "XXX"
	after, err := IntegrityCommitment(value)
	if err != nil || after != before || VerifyIntegrity(value) != nil {
		t.Error("report shares mutable memory with its inputs")
	}
	// Separately built reports share no pointers.
	other := runPipeline(t, 26, "fixture-source").complete(accounting.Options{})
	if value.Manifest == other.Manifest || value.C1 == other.C1 || value.Manifest.AnalysisInterval.Start == value.AnalysisInterval.Start || &value.Manifest.Lineage.Report.InputRefs[0] == &other.Manifest.Lineage.Report.InputRefs[0] {
		t.Error("reports or manifest and report share pointers")
	}
}

func TestReportAndManifestDoNotLeakSensitiveMaterial(t *testing.T) {
	value := runPipeline(t, 27, "fixture-source").complete(accounting.Options{BillingModel: testBillingModel()})
	var encoded bytes.Buffer
	if err := WriteJSON(&encoded, value); err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"pf_fake_secret", "Bearer ", "/Users/", "/home/", "C:\\", "github_pat_", "ghp_", "GITHUB_TOKEN", "PATH="} {
		if strings.Contains(encoded.String(), forbidden) {
			t.Errorf("report leaked %q", forbidden)
		}
	}
}

func BenchmarkCompleteReportWithManifest(b *testing.B) {
	p := runPipeline(b, 31, "fixture-source")
	accountingResult := accounting.Build(p.replay, p.c1, p.c2, accounting.Options{BillingModel: testBillingModel()})
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_ = Complete(p.replay, p.policy, "test", p.c1, p.c2, accountingResult)
	}
}

func BenchmarkVerifyIntegrity(b *testing.B) {
	value := runPipeline(b, 32, "fixture-source").complete(accounting.Options{})
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if err := VerifyIntegrity(value); err != nil {
			b.Fatal(err)
		}
	}
}
