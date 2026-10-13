package provenance

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/amahon78/pulsefold/internal/accounting"
	"github.com/amahon78/pulsefold/internal/c1"
	"github.com/amahon78/pulsefold/internal/c2"
	"github.com/amahon78/pulsefold/internal/input"
)

func TestCompleteManifestRecordsPayloadFreeLineage(t *testing.T) {
	replay := testReplay()
	c1Result := c1.Result{CanonicalisationVersion: c1.CanonicalisationVersion, CandidatePolicyVersion: c1.CandidatePolicyVersion}
	c2Result := c2.Result{AlgorithmVersion: c2.AlgorithmVersion, CandidatePolicyVersion: c2.CandidatePolicyVersion, MaxClusters: 10, MaxPartitions: 11, MaxStateBytes: 12, MaxBodyBytes: 13, MaxTemplateBytes: 14, MaxTokens: 15, MaxSourceTemplates: 16}
	accountingResult := accounting.Result{Billing: accounting.Billing{ModelSupplied: true, Classification: accounting.UserSupplied, ModelSchemaVersion: accounting.BillingSchemaVersion, ModelCommitment: "sha256:0123456789012345678901234567890123456789012345678901234567890123"}}
	manifest := Build(replay, "pulsefold/v0alpha1", "test", "v1-fixture-safety", Completion{State: "complete"}, &c1Result, &c2Result, accountingResult)
	AttachIntegrity(&manifest, Commitment([]byte(`{"report":"safe"}`)))
	manifest.CreatedAt = testCreatedAt
	if err := manifest.Validate(); err != nil {
		t.Fatal(err)
	}
	if manifest.Lineage.C1.Classification != accounting.Counterfactual || manifest.Lineage.C2.Classification != accounting.Inferred {
		t.Fatalf("analysis classifications = %#v", manifest.Lineage)
	}
	if manifest.Lineage.Accounting.BillingOrigin != OriginOperatorSupplied || manifest.Lineage.Accounting.BillingClassification != accounting.UserSupplied {
		t.Errorf("billing provenance = %#v", manifest.Lineage.Accounting)
	}
	if manifest.Lineage.C2.PartitionPolicyVersion != c2.PartitionPolicyVersion || manifest.Lineage.C2.Bounds.MaxTokens != 15 {
		t.Errorf("C2 provenance = %#v", manifest.Lineage.C2)
	}
	if manifest.Content.State != ContentNotStored || manifest.Integrity.Authenticated {
		t.Errorf("content or integrity semantics = %#v", manifest)
	}
}

func TestManifestDoesNotExposeRawSourceLabelOrTelemetryLikeText(t *testing.T) {
	replay := testReplay()
	replay.Source.Identity = "/Users/example/private/Bearer pf_fake_secret_manifest"
	replay.Source.Partition = "orders/pf_fake_secret_partition"
	manifest := Build(replay, "pulsefold/v0alpha1", "test", "v1", Completion{State: "complete"}, nil, nil, accounting.Result{})
	AttachIntegrity(&manifest, Commitment([]byte("safe")))
	payload, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	encoded := string(payload)
	for _, forbidden := range []string{"/Users/", "Bearer", "pf_fake_secret", "orders/"} {
		if strings.Contains(encoded, forbidden) {
			t.Fatalf("manifest leaked %q: %s", forbidden, encoded)
		}
	}
}

func TestLifecycleRecordsDoNotMutateManifestOrClaimUnavailableContentIsRetrievable(t *testing.T) {
	manifest := completeManifest(t)
	original := manifest
	available, err := NewLifecycleRecord(manifest.ManifestID, ContentAvailable, time.Date(2026, 10, 13, 1, 2, 3, 0, time.UTC), "", "content_ref_opaque")
	if err != nil || !available.Retrievable() {
		t.Fatalf("available lifecycle = %#v, %v", available, err)
	}
	for _, state := range []ContentState{ContentExpired, ContentErased} {
		record, err := NewLifecycleRecord(manifest.ManifestID, state, time.Date(2026, 10, 13, 1, 2, 4, 0, time.UTC), "retention_policy", "")
		if err != nil || record.Retrievable() {
			t.Fatalf("unavailable lifecycle %q = %#v, %v", state, record, err)
		}
	}
	if !reflect.DeepEqual(manifest, original) || manifest.Content.State != ContentNotStored {
		t.Errorf("lifecycle mutated immutable manifest: %#v", manifest)
	}
	if _, err := NewLifecycleRecord(manifest.ManifestID, ContentErased, time.Time{}, "", "content_ref_opaque"); err == nil {
		t.Error("invalid erased lifecycle record was accepted")
	}
	notStored, err := NewLifecycleRecord(manifest.ManifestID, ContentNotStored, testCreatedAt, "", "")
	if err != nil || notStored.Retrievable() {
		t.Errorf("not_stored lifecycle = %#v, %v", notStored, err)
	}
	for _, invalid := range []struct {
		state            ContentState
		reason, location string
	}{
		{ContentErased, "erased because Bearer pf_fake_secret", ""},
		{ContentExpired, "retention", "/Users/example/content.bin"},
		{ContentAvailable, "", "https://user:pass@example.invalid/x"},
		{ContentAvailable, "", ""},
		{ContentNotStored, "", "content_ref_opaque"},
		{ContentState("archived"), "", ""},
	} {
		if record, err := NewLifecycleRecord(manifest.ManifestID, invalid.state, testCreatedAt, invalid.reason, invalid.location); err == nil || record.Retrievable() {
			t.Errorf("invalid lifecycle accepted: %#v", invalid)
		}
	}
}

func TestManifestJSONRoundTripAndStrictParsing(t *testing.T) {
	manifest := completeManifest(t)
	payload, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := ParseManifestJSON(bytes.NewReader(payload))
	if err != nil || !reflect.DeepEqual(decoded, manifest) {
		t.Fatalf("manifest round trip = %#v, %v", decoded, err)
	}
	for _, malformed := range [][]byte{
		[]byte(`{"schema_version":"` + SchemaVersion + `","unknown":"pf_fake_secret"}`),
		[]byte(strings.Replace(string(payload), `"commitment":"sha256:`, `"commitment":"sha256:broken`, 1)),
		[]byte(`not-json`),
		bytes.Repeat([]byte("x"), maxManifestBytes+1),
	} {
		if _, err := ParseManifestJSON(bytes.NewReader(malformed)); err == nil || strings.Contains(err.Error(), "pf_fake_secret") {
			t.Errorf("unsafe parse result: %v", err)
		}
	}
}

func BenchmarkBuildManifest(b *testing.B) {
	replay := testReplay()
	c1Result := c1.Result{CanonicalisationVersion: c1.CanonicalisationVersion, CandidatePolicyVersion: c1.CandidatePolicyVersion}
	c2Result := c2.Result{AlgorithmVersion: c2.AlgorithmVersion, CandidatePolicyVersion: c2.CandidatePolicyVersion, MaxClusters: 10, MaxPartitions: 10, MaxStateBytes: 100, MaxBodyBytes: 100, MaxTemplateBytes: 100, MaxTokens: 10, MaxSourceTemplates: 10}
	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		manifest := Build(replay, "pulsefold/v0alpha1", "test", "v1", Completion{State: "complete"}, &c1Result, &c2Result, accounting.Result{})
		AttachIntegrity(&manifest, Commitment([]byte("safe")))
	}
}

func completeManifest(t *testing.T) Manifest {
	t.Helper()
	manifest := Build(testReplay(), "pulsefold/v0alpha1", "test", "v1", Completion{State: "complete"}, nil, nil, accounting.Result{})
	AttachIntegrity(&manifest, Commitment([]byte("safe")))
	manifest.CreatedAt = testCreatedAt
	if err := manifest.Validate(); err != nil {
		t.Fatal(err)
	}
	return manifest
}

var testCreatedAt = time.Date(2026, 10, 13, 5, 0, 0, 0, time.UTC)

func testReplay() input.ReplayResult {
	start := time.Date(2026, 10, 13, 1, 0, 0, 0, time.UTC)
	end := start.Add(time.Minute)
	return input.ReplayResult{Source: input.SourceMetadata{Identity: "fixture-source", Partition: "fixture", Kind: input.SourceOfflineReplay, Format: input.FormatOTLPJSON}, Coverage: input.Coverage{Interval: input.Interval{Known: true, Start: start, End: end}}}
}

func TestIdentifierSemantics(t *testing.T) {
	first, second := completeManifest(t), completeManifest(t)
	if first.Source.IdentityReference != second.Source.IdentityReference || first.ManifestID != second.ManifestID {
		t.Error("source reference or manifest ID is not stable across identical replay")
	}
	if first.ManifestID != "pfm_v0_"+strings.TrimPrefix(first.Integrity.Commitment, "sha256:") {
		t.Errorf("manifest ID is not content-derived: %q", first.ManifestID)
	}
	other := testReplay()
	other.Source.Identity = "other-source"
	changed := Build(other, "pulsefold/v0alpha1", "test", "v1", Completion{State: "complete"}, nil, nil, accounting.Result{})
	if changed.Source.IdentityReference == first.Source.IdentityReference || changed.Source.PartitionReference != first.Source.PartitionReference {
		t.Error("source references are not label-specific")
	}
	if strings.Contains(first.Source.IdentityReference, "fixture") || !strings.HasPrefix(first.Source.IdentityReference, "ref:") {
		t.Errorf("source reference is not opaque: %q", first.Source.IdentityReference)
	}
	// The identity and partition namespaces are separate.
	if metadataReference("source-identity", "x") == metadataReference("source-partition", "x") {
		t.Error("reference namespaces collide")
	}
}

func TestManifestValidationRejectsInconsistentManifests(t *testing.T) {
	mutations := map[string]func(*Manifest){
		"authenticated":   func(m *Manifest) { m.Integrity.Authenticated = true },
		"id mismatch":     func(m *Manifest) { m.ManifestID = "pfm_v0_tampered" },
		"no created time": func(m *Manifest) { m.CreatedAt = time.Time{} },
		"bad completion":  func(m *Manifest) { m.Completion.State = "done" },
		"failed no code":  func(m *Manifest) { m.Completion.State = "input_failed" },
		"complete code":   func(m *Manifest) { m.Completion.FailureCode = "malformed_input" },
		"lineage":         func(m *Manifest) { m.Lineage.C2.InputRef = "capture" },
		"content state":   func(m *Manifest) { m.Content.State = "archived" },
		"algorithm":       func(m *Manifest) { m.Integrity.Algorithm = "md5" },
	}
	for name, mutate := range mutations {
		manifest := completeManifest(t)
		mutate(&manifest)
		if err := manifest.Validate(); err == nil {
			t.Errorf("%s: invalid manifest accepted", name)
		}
	}
}

func TestCloneAndCommitmentViewDoNotAlias(t *testing.T) {
	manifest := completeManifest(t)
	original := manifest.Clone()
	clone := manifest.Clone()
	*clone.AnalysisInterval.Start = clone.AnalysisInterval.Start.Add(time.Hour)
	clone.Lineage.Report.InputRefs[0] = "tampered"
	clone.Lineage.Accounting.InputRefs[0] = "tampered"
	view := manifest.CommitmentView()
	*view.AnalysisInterval.End = view.AnalysisInterval.End.Add(time.Hour)
	if !reflect.DeepEqual(manifest, original) {
		t.Errorf("clone or commitment view aliased the manifest: %#v", manifest)
	}
	if view.ManifestID != "" || view.Integrity != (Integrity{}) || !view.CreatedAt.IsZero() {
		t.Errorf("commitment view retained volatile or self-referential fields: %#v", view)
	}
}

func TestManifestWithoutAnalysisOmitsAnalysisVersions(t *testing.T) {
	manifest := completeManifest(t)
	if manifest.Lineage.C1.Classification != "" || manifest.Lineage.C2.AlgorithmVersion != "" || manifest.Lineage.Accounting.BillingOrigin != OriginUnknown {
		t.Errorf("unavailable stages claimed provenance: %#v", manifest.Lineage)
	}
}

// Property: references are fixed-shape, deterministic, and never echo input.
func TestSourceReferenceProperty(t *testing.T) {
	random := rand.New(rand.NewSource(1006))
	for iteration := 0; iteration < 2000; iteration++ {
		raw := make([]byte, random.Intn(300))
		random.Read(raw)
		value := string(raw)
		reference := metadataReference("source-identity", value)
		if reference != metadataReference("source-identity", value) || len(reference) != len("ref:")+64 {
			t.Fatalf("reference is not deterministic or fixed-size: %q", reference)
		}
		if _, err := hex.DecodeString(strings.TrimPrefix(reference, "ref:")); err != nil {
			t.Fatalf("reference is not hex: %q", reference)
		}
		if len(value) >= 8 && strings.Contains(reference, value) {
			t.Fatalf("reference echoed input")
		}
	}
}
