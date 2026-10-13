package report

import (
	"bytes"
	"strings"
	"testing"

	"github.com/amahon78/pulsefold/internal/accounting"
	"github.com/amahon78/pulsefold/internal/input"
)

func TestWriteTextSeparatesCoverageC1C2EconomicsAndIntegrity(t *testing.T) {
	value := runPipeline(t, 61, "fixture-source").complete(accounting.Options{BillingModel: testBillingModel()})
	var output bytes.Buffer
	if err := WriteText(&output, value); err != nil {
		t.Fatal(err)
	}
	text := output.String()
	for _, expected := range []string{
		"Accepted records:",
		"C1 exact repetition (counterfactual):",
		"C2 structural repetition (inferred):",
		"C2 is an inferred structural candidate opportunity, not verified or removable telemetry.",
		"Meter: accepted_records",
		"C1 counterfactual opportunity:",
		"Integrity: verified SHA-256 consistency check",
		"Authentication: not provided",
	} {
		if !strings.Contains(text, expected) {
			t.Errorf("summary missing %q: %s", expected, text)
		}
	}
	for _, forbidden := range []string{"pf_fake_secret", "Bearer ", "/Users/", "github_pat_", "ghp_", "Resly"} {
		if strings.Contains(text, forbidden) {
			t.Errorf("summary leaked %q", forbidden)
		}
	}
}

func TestWriteTextHonoursFailureAndUnknownCoverage(t *testing.T) {
	value := InputFailure(input.ReplayResult{Source: input.SourceMetadata{Identity: "fixture-source", Kind: input.SourceOfflineReplay, Format: input.FormatOTLPJSON}, Coverage: input.Coverage{InputByteLimitHit: true}}, "test", "input_byte_limit")
	var output bytes.Buffer
	if err := WriteText(&output, value); err != nil {
		t.Fatal(err)
	}
	text := output.String()
	for _, expected := range []string{"Status: input_failed", "Accepted records: denominator unknown", "Integrity: verified SHA-256 consistency check", "Analysis did not complete: input/input_byte_limit."} {
		if !strings.Contains(text, expected) {
			t.Errorf("failure summary missing %q: %s", expected, text)
		}
	}
}
