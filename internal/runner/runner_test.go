package runner

import (
	"bytes"
	"context"
	"strings"
	"testing"

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
	first := Analyse(context.Background(), source(payload), input.Config{}, zeroC1Config(), zeroC2Config())
	if first.Err != nil {
		t.Fatal(first.Err)
	}
	second := Analyse(context.Background(), source(payload), input.Config{}, zeroC1Config(), zeroC2Config())
	if second.Err != nil {
		t.Fatal(second.Err)
	}
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
	if strings.Contains(encoded, "pf_fake_secret_") || strings.Contains(encoded, "Bearer ") {
		t.Error("report leaked a fixture secret")
	}
	if !strings.Contains(encoded, "counterfactual reduction opportunity") || !strings.Contains(encoded, "\"status\": \"complete\"") {
		t.Errorf("report lacks required V0 semantics: %s", encoded)
	}
}

func TestAnalyseReturnsSafeFailureReportForMalformedInput(t *testing.T) {
	secret := "Bearer pf_fake_secret_malformed"
	outcome := Analyse(context.Background(), input.ReaderSource{
		Reader: strings.NewReader(`{"resourceLogs":` + secret),
		Info:   input.SourceMetadata{Identity: "malformed", Kind: input.SourceOfflineReplay, Format: input.FormatOTLPJSON},
	}, input.Config{}, zeroC1Config(), zeroC2Config())
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
}

func TestAnalyseC2ReportDoesNotContainTemplateOrRawTelemetry(t *testing.T) {
	scenario := fixture.GenerateTemplateScenario(fixture.Config{Seed: 11})
	payload, err := protojson.Marshal(scenario.Request)
	if err != nil {
		t.Fatal(err)
	}
	outcome := Analyse(context.Background(), source(payload), input.Config{}, zeroC1Config(), zeroC2Config())
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

func source(payload []byte) input.ReaderSource {
	return input.ReaderSource{Reader: bytes.NewReader(payload), Info: input.SourceMetadata{Identity: "runner-fixture", Partition: "test", Kind: input.SourceOfflineReplay, Format: input.FormatOTLPJSON}}
}

func zeroC1Config() c1.Config {
	return c1.Config{}
}

func zeroC2Config() c2.Config {
	return c2.Config{}
}
