package runner

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/amahon78/pulsefold/internal/accounting"
	"github.com/amahon78/pulsefold/internal/c1"
	"github.com/amahon78/pulsefold/internal/c2"
	"github.com/amahon78/pulsefold/internal/fixture"
	"github.com/amahon78/pulsefold/internal/input"
	"github.com/amahon78/pulsefold/internal/report"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestAnalyseProducesPayloadFreeDeterministicReport(t *testing.T) {
	scenario := fixture.Generate(fixture.Config{Seed: 9, HighCardinalityCount: 2, LargeBodyBytes: 64})
	payload, err := protojson.Marshal(scenario.Request)
	if err != nil {
		t.Fatal(err)
	}
	first := Analyse(context.Background(), source(payload), input.Config{}, zeroC1Config(), zeroC2Config(), accounting.Options{})
	if first.Err != nil {
		t.Fatal(first.Err)
	}
	second := Analyse(context.Background(), source(payload), input.Config{}, zeroC1Config(), zeroC2Config(), accounting.Options{})
	if second.Err != nil {
		t.Fatal(second.Err)
	}
	// created_at is the only volatile field. Everything else, including the
	// content-derived manifest ID and integrity commitment, is deterministic.
	if first.Report.Manifest == nil || second.Report.Manifest == nil || first.Report.Manifest.CreatedAt.IsZero() {
		t.Fatal("report manifest is missing or lacks a creation time")
	}
	if first.Report.Manifest.ManifestID != second.Report.Manifest.ManifestID || first.Report.Manifest.Integrity != second.Report.Manifest.Integrity {
		t.Error("identical replay produced a different manifest identity or commitment")
	}
	normalised := *second.Report.Manifest
	normalised.CreatedAt = first.Report.Manifest.CreatedAt
	second.Report.Manifest = &normalised
	var firstJSON, secondJSON bytes.Buffer
	if err := report.WriteJSON(&firstJSON, first.Report); err != nil {
		t.Fatal(err)
	}
	if err := report.WriteJSON(&secondJSON, second.Report); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstJSON.Bytes(), secondJSON.Bytes()) {
		t.Error("identical replay produced nondeterministic JSON")
	}
	encoded := firstJSON.String()
	for _, forbidden := range []string{"pf_fake_secret_", "Bearer ", "/Users/", "github_pat_", "ghp_"} {
		if strings.Contains(encoded, forbidden) {
			t.Errorf("report leaked forbidden material %q", forbidden)
		}
	}
	if !strings.Contains(encoded, "counterfactual reduction opportunity") || !strings.Contains(encoded, "\"status\": \"complete\"") {
		t.Errorf("report lacks required V0 semantics: %s", encoded)
	}
	if first.Report.C1.Classification != accounting.Counterfactual || first.Report.C2.Classification != accounting.Inferred {
		t.Errorf("analysis classifications = C1 %q, C2 %q", first.Report.C1.Classification, first.Report.C2.Classification)
	}
	if first.Report.Accounting.Coverage.C1.Coverage.Known && first.Report.Accounting.Coverage.C1.Coverage.Numerator != first.Report.C1.RecordsAnalysed {
		t.Errorf("C1 coverage diverged from report counters: %#v", first.Report.Accounting.Coverage.C1)
	}
	if first.Report.Accounting.Wire.Known || first.Report.Accounting.Billing.Available || first.Report.Accounting.Billing.Reason != "billing_model_not_supplied" {
		t.Errorf("default accounting semantics = %#v", first.Report.Accounting)
	}
	if first.Report.Manifest.Completion.State != string(report.StatusComplete) || report.VerifyIntegrity(first.Report) != nil {
		t.Errorf("report manifest = %#v", first.Report.Manifest)
	}
	// Report is a value with pointer fields, so a plain copy aliases C1. The
	// copy below is explicit so the original report is not mutated.
	modified := first.Report
	modifiedC1 := *first.Report.C1
	modified.C1 = &modifiedC1
	modified.C1.RawByteOpportunity++
	if report.VerifyIntegrity(modified) == nil || report.VerifyIntegrity(first.Report) != nil {
		t.Error("integrity did not detect a content change, or the original was mutated")
	}
}

func TestAnalyseManifestCarriesBillingProvenance(t *testing.T) {
	scenario := fixture.Generate(fixture.Config{Seed: 15, HighCardinalityCount: 0, LargeBodyBytes: 0})
	payload, err := protojson.Marshal(scenario.Request)
	if err != nil {
		t.Fatal(err)
	}
	model := &accounting.Model{SchemaVersion: accounting.BillingSchemaVersion, Currency: "TST", Metric: accounting.MetricAcceptedRecords, IndexedFraction: accounting.Fraction{Numerator: 1, Denominator: 1}, Pricing: accounting.Pricing{Fixed: &accounting.Rate{UnitBlock: 1, MinorUnitsPerBlock: 1}}}
	outcome := Analyse(context.Background(), source(payload), input.Config{}, zeroC1Config(), zeroC2Config(), accounting.Options{BillingModel: model})
	if outcome.Err != nil || outcome.Report.Manifest == nil {
		t.Fatalf("billing provenance outcome = %#v", outcome)
	}
	stage := outcome.Report.Manifest.Lineage.Accounting
	if stage.BillingOrigin != "operator_supplied" || stage.BillingClassification != accounting.UserSupplied || stage.BillingModelSchema != accounting.BillingSchemaVersion || stage.BillingModelCommitment == "" {
		t.Errorf("billing manifest lineage = %#v", stage)
	}
}

func TestUnsafeSourceLabelDoesNotLeakIntoFailureReportOrManifest(t *testing.T) {
	unsafe := "/Users/example/private/pf_fake_secret_source"
	outcome := Analyse(context.Background(), input.ReaderSource{Reader: strings.NewReader(`{}`), Info: input.SourceMetadata{Identity: unsafe, Kind: input.SourceOfflineReplay, Format: input.FormatOTLPJSON}}, input.Config{}, zeroC1Config(), zeroC2Config(), accounting.Options{})
	if outcome.Err == nil || outcome.Report.Manifest == nil {
		t.Fatalf("unsafe source outcome = %#v", outcome)
	}
	var output bytes.Buffer
	if err := report.WriteJSON(&output, outcome.Report); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), unsafe) || strings.Contains(output.String(), "pf_fake_secret_source") {
		t.Errorf("unsafe source leaked into report: %s", output.String())
	}
}

func TestAnalyseReturnsSafeFailureReportForMalformedInput(t *testing.T) {
	secret := "Bearer pf_fake_secret_malformed"
	outcome := Analyse(context.Background(), input.ReaderSource{
		Reader: strings.NewReader(`{"resourceLogs":` + secret),
		Info:   input.SourceMetadata{Identity: "malformed", Kind: input.SourceOfflineReplay, Format: input.FormatOTLPJSON},
	}, input.Config{}, zeroC1Config(), zeroC2Config(), accounting.Options{})
	if outcome.Err == nil || outcome.Report.Status != report.StatusInputFailed || outcome.Report.Failure.Code != "malformed_input" {
		t.Fatalf("unexpected malformed outcome: %#v", outcome.Report)
	}
	var output bytes.Buffer
	if err := report.WriteJSON(&output, outcome.Report); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), secret) {
		t.Error("input failure report leaked raw telemetry")
	}
	if outcome.Report.Accounting.Coverage.C1.Coverage.Known || outcome.Report.Accounting.Coverage.Input.RecordsObservedKnown {
		t.Errorf("input failure accounting claimed unavailable analysis: %#v", outcome.Report.Accounting)
	}
	if outcome.Report.Manifest == nil || outcome.Report.Manifest.Completion.State != string(report.StatusInputFailed) || outcome.Report.Manifest.Completion.FailureCode != "malformed_input" {
		t.Errorf("input failure manifest = %#v", outcome.Report.Manifest)
	}
}

func TestAnalyseFailureMarksAnalysisCoverageUnavailable(t *testing.T) {
	scenario := fixture.Generate(fixture.Config{Seed: 13, HighCardinalityCount: 2, LargeBodyBytes: 32})
	payload, err := protojson.Marshal(scenario.Request)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	config := c1.Config{Hash: func(value []byte) string {
		cancel()
		return "deterministic-test-bucket"
	}}
	outcome := Analyse(ctx, source(payload), input.Config{}, config, zeroC2Config(), accounting.Options{})
	if outcome.Err == nil || outcome.Report.Status != report.StatusAnalysisFailed || outcome.Report.Failure.Code != "cancelled" {
		t.Fatalf("unexpected interrupted outcome: %#v", outcome.Report)
	}
	if outcome.Report.Accounting.Coverage.C1.Coverage.Known || outcome.Report.Accounting.Coverage.C2.Coverage.Known || outcome.Report.Accounting.Coverage.C1.Coverage.Reason != "c1_not_completed" {
		t.Errorf("analysis failure coverage = %#v", outcome.Report.Accounting.Coverage)
	}
	if outcome.Report.Manifest == nil || outcome.Report.Manifest.Completion.State != string(report.StatusAnalysisFailed) || outcome.Report.Manifest.Completion.FailureCode != "cancelled" {
		t.Errorf("analysis failure manifest = %#v", outcome.Report.Manifest)
	}
}

func TestAnalyseC2ReportDoesNotContainTemplateOrRawTelemetry(t *testing.T) {
	scenario := fixture.GenerateTemplateScenario(fixture.Config{Seed: 11})
	payload, err := protojson.Marshal(scenario.Request)
	if err != nil {
		t.Fatal(err)
	}
	outcome := Analyse(context.Background(), source(payload), input.Config{}, zeroC1Config(), zeroC2Config(), accounting.Options{})
	if outcome.Err != nil || outcome.Report.C2 == nil || outcome.Report.C2.SourceTemplateRecords == 0 {
		t.Fatalf("unexpected C2 outcome: %#v", outcome)
	}
	var output bytes.Buffer
	if err := report.WriteJSON(&output, outcome.Report); err != nil {
		t.Fatal(err)
	}
	encoded := output.String()
	for _, forbidden := range []string{"login accepted", "pf_fake_secret_", "Bearer ", "gateway failed"} {
		if strings.Contains(encoded, forbidden) {
			t.Fatalf("C2 report leaked telemetry text %q", forbidden)
		}
	}
}

func TestAnalyseMakesBoundedStageExclusionsVisibleInShadowReport(t *testing.T) {
	scenario := fixture.GenerateTemplateScenario(fixture.Config{Seed: 72})
	payload, err := protojson.Marshal(scenario.Request)
	if err != nil {
		t.Fatal(err)
	}
	for name, config := range map[string]struct {
		c1Config c1.Config
		c2Config c2.Config
		expected string
	}{
		"C1 state cap": {c1Config: c1.Config{MaxGroups: 1}, c2Config: zeroC2Config(), expected: "C1 analysis coverage is incomplete"},
		"C2 state cap": {c1Config: zeroC1Config(), c2Config: c2.Config{MaxClusters: 1}, expected: "C2 analysis coverage is incomplete"},
	} {
		t.Run(name, func(t *testing.T) {
			outcome := Analyse(context.Background(), source(payload), input.Config{}, config.c1Config, config.c2Config, accounting.Options{})
			if outcome.Err != nil || outcome.Report.C1 == nil || outcome.Report.C2 == nil {
				t.Fatalf("outcome = %#v", outcome)
			}
			var output bytes.Buffer
			if err := report.WriteText(&output, outcome.Report); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(output.String(), config.expected) {
				t.Errorf("bounded exclusion was not visible: %s", output.String())
			}
		})
	}
}

func BenchmarkAnalyseAndRenderShadowReport(b *testing.B) {
	scenario := fixture.GenerateTemplateScenario(fixture.Config{Seed: 71, HighCardinalityCount: 32, LargeBodyBytes: 512})
	payload, err := protojson.Marshal(scenario.Request)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.SetBytes(int64(len(payload)))
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		outcome := Analyse(context.Background(), source(payload), input.Config{}, zeroC1Config(), zeroC2Config(), accounting.Options{})
		if outcome.Err != nil {
			b.Fatal(outcome.Err)
		}
		var output bytes.Buffer
		if err := report.WriteText(&output, outcome.Report); err != nil {
			b.Fatal(err)
		}
	}
}

func source(payload []byte) input.ReaderSource {
	return input.ReaderSource{Reader: bytes.NewReader(payload), Info: input.SourceMetadata{Identity: "runner-fixture", Partition: "test", Kind: input.SourceOfflineReplay, Format: input.FormatOTLPJSON}}
}

func zeroC1Config() c1.Config {
	return c1.Config{}
}

func zeroC2Config() c2.Config {
	return c2.Config{}
}
