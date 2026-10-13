package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/amahon78/pulsefold/internal/accounting"
	"github.com/amahon78/pulsefold/internal/c1"
	"github.com/amahon78/pulsefold/internal/c2"
	"github.com/amahon78/pulsefold/internal/input"
	"github.com/amahon78/pulsefold/internal/report"
)

// adversarialPayload builds one OTLP/JSON request with string bodies only.
func adversarialPayload(t *testing.T, bodies []string) []byte {
	t.Helper()
	records := make([]string, 0, len(bodies))
	for _, body := range bodies {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		records = append(records, fmt.Sprintf(`{"timeUnixNano":"1700000000000000000","severityNumber":9,"body":{"stringValue":%s}}`, encoded))
	}
	return []byte(`{"resourceLogs":[{"resource":{"attributes":[{"key":"service.name","value":{"stringValue":"svc"}}]},"scopeLogs":[{"scope":{"name":"s"},"logRecords":[` + strings.Join(records, ",") + `]}]}]}`)
}

func analyseAdversarial(t *testing.T, bodies []string, options accounting.Options) report.Report {
	t.Helper()
	outcome := Analyse(context.Background(), input.ReaderSource{Reader: bytes.NewReader(adversarialPayload(t, bodies)), Info: input.SourceMetadata{Identity: "adversarial", Kind: input.SourceOfflineReplay, Format: input.FormatOTLPJSON}}, input.Config{}, c1.Config{}, c2.Config{}, options)
	if outcome.Err != nil || outcome.Report.C1 == nil || outcome.Report.C2 == nil {
		t.Fatalf("adversarial replay did not complete: %v", outcome.Err)
	}
	return outcome.Report
}

// Regression for the V0 adversarial review: C2's "beyond C1" opportunity
// previously used C1-weighted candidate records, so purely exact repetition
// was claimed a second time as structural opportunity. C2 must claim only
// representatives that collapse into an existing inferred group.
func TestC2OpportunityDoesNotReclaimC1ExactRepetition(t *testing.T) {
	exact := make([]string, 100)
	for index := range exact {
		exact[index] = "identical heartbeat"
	}
	value := analyseAdversarial(t, exact, accounting.Options{})
	if value.C1.DuplicateRecords != 99 || value.C2.AdditionalOpportunity != 0 {
		t.Fatalf("exact-only input: C1 duplicates=%d, C2 beyond C1=%d; want 99 and 0", value.C1.DuplicateRecords, value.C2.AdditionalOpportunity)
	}

	// Two exact groups of 50 that share a structure: C2 may claim exactly one
	// additional representative, never the 98 records C1 already claimed.
	structural := make([]string, 0, 100)
	for index := 0; index < 50; index++ {
		structural = append(structural, "request_id=aaaa served", "request_id=bbbb served")
	}
	value = analyseAdversarial(t, structural, accounting.Options{})
	if value.C1.DuplicateRecords != 98 || value.C2.AdditionalOpportunity != 1 || value.C2.InferredTemplateGroups != 1 {
		t.Fatalf("structural input: C1 duplicates=%d, C2 beyond C1=%d, groups=%d; want 98, 1 and 1", value.C1.DuplicateRecords, value.C2.AdditionalOpportunity, value.C2.InferredTemplateGroups)
	}
	var text bytes.Buffer
	if err := report.WriteText(&text, value); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text.String(), "Candidate records beyond C1: 1\n") {
		t.Errorf("text report misstates C2 opportunity beyond C1:\n%s", text.String())
	}
}

// The billing C2 potential is measured over the C1-weighted candidate subset
// and therefore overlaps the C1 opportunity. A reader adding the two lines
// would claim more than the baseline, so the overlap must be a committed
// limitation whenever both are priced.
func TestPricedC1AndC2OpportunitiesDeclareOverlap(t *testing.T) {
	bodies := make([]string, 1000)
	for index := range bodies {
		bodies[index] = "identical heartbeat"
	}
	model := accounting.Model{SchemaVersion: accounting.BillingSchemaVersion, Currency: "USD", Metric: accounting.MetricAcceptedRecords, IndexedFraction: accounting.Fraction{Numerator: 1, Denominator: 1}, Pricing: accounting.Pricing{Fixed: &accounting.Rate{UnitBlock: 1, MinorUnitsPerBlock: 10}}}
	value := analyseAdversarial(t, bodies, accounting.Options{BillingModel: &model})
	billing := value.Accounting.Billing
	if billing.Baseline == nil || billing.C1 == nil || billing.C2 == nil {
		t.Fatalf("billing was not priced: %#v", billing)
	}
	if billing.C1.GrossCostOpportunity.MinorUnits+billing.C2.GrossCostOpportunity.MinorUnits <= billing.Baseline.Cost.MinorUnits {
		t.Skip("C1 and C2 no longer overlap; revisit this limitation")
	}
	found := false
	for _, limitation := range value.Limitations {
		if strings.Contains(limitation, "must not be added") {
			found = true
		}
	}
	if !found {
		t.Errorf("overlapping C1 and C2 cost opportunities lack a limitation: %q", value.Limitations)
	}
	if err := report.VerifyIntegrity(value); err != nil {
		t.Errorf("report with overlap limitation failed integrity: %v", err)
	}
}
