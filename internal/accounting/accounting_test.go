package accounting

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/amahon78/pulsefold/internal/c1"
	"github.com/amahon78/pulsefold/internal/c2"
	"github.com/amahon78/pulsefold/internal/input"
)

func TestBuildSeparatesMeasuredCoverageCounterfactualAndInferredOpportunity(t *testing.T) {
	replay := completeReplay()
	c1Result := c1.Result{RecordsPresented: 10, RecordsAnalysed: 8, ExcludedRecords: 2, CounterfactualRecords: 3, InputCanonicalBytes: 100, CounterfactualBytes: 40}
	c2Result := c2.Result{C1RepresentativeRecords: 3, RecordsAnalysed: 2, CandidateRecords: 8, InferredTemplateGroups: 1, ExcludedRecords: 4}
	result := Build(replay, c1Result, c2Result, Options{BillingModel: model(MetricAcceptedRecords)})

	if !result.Coverage.Input.AcceptedRecordCoverage.Known || result.Coverage.Input.AcceptedRecordCoverage.Numerator != 10 || result.Coverage.Input.AcceptedRecordCoverage.Denominator != 10 {
		t.Fatalf("input coverage = %#v", result.Coverage.Input)
	}
	if got := result.Coverage.C1; !got.Coverage.Known || got.Eligible != 10 || got.Analysed != 8 || got.Excluded != 2 {
		t.Fatalf("C1 coverage = %#v", got)
	}
	if got := result.Coverage.C2; !got.Coverage.Known || got.Eligible != 3 || got.Analysed != 2 || got.Excluded != 1 {
		t.Fatalf("C2 coverage = %#v", got)
	}
	if result.Wire.Classification != Unknown || result.Wire.Known || result.Wire.Reason != "wire_bytes_not_measured" {
		t.Fatalf("wire result = %#v", result.Wire)
	}
	if !result.Billing.Available || result.Billing.Baseline == nil || result.Billing.C1 == nil || result.Billing.C2 == nil {
		t.Fatalf("billing = %#v", result.Billing)
	}
	if result.Billing.Baseline.Units != 5 || result.Billing.Baseline.Classification != Measured || result.Billing.Baseline.Cost.MinorUnits != 50 {
		t.Errorf("baseline = %#v", result.Billing.Baseline)
	}
	if got := result.Billing.C1; got.Classification != Counterfactual || got.BaselineUnits != 4 || got.CounterfactualUnits != 2 || got.UnitReduction != 2 || got.GrossCostOpportunity.MinorUnits != 20 {
		t.Errorf("C1 opportunity = %#v", got)
	}
	if got := result.Billing.C2; got.Classification != Inferred || got.BaselineUnits != 4 || got.CounterfactualUnits != 1 || got.UnitReduction != 3 || got.GrossCostOpportunity.MinorUnits != 30 {
		t.Errorf("C2 opportunity = %#v", got)
	}
}

func TestBuildReturnsExplicitUnknownsWithoutACompleteDenominatorOrBillingModel(t *testing.T) {
	replay := input.ReplayResult{Coverage: input.Coverage{InputBytesRead: 17, RecordsAccepted: 2, CaptureErrors: 1}}
	result := FromInput(replay)
	if result.Coverage.Input.AcceptedRecordCoverage.Known || result.Coverage.Input.AcceptedRecordCoverage.Reason != "records_observed_unknown" {
		t.Errorf("input ratio = %#v", result.Coverage.Input.AcceptedRecordCoverage)
	}
	if result.Billing.Available || result.Billing.Reason != "billing_model_not_supplied" || result.Billing.Classification != Unknown {
		t.Errorf("billing = %#v", result.Billing)
	}
	if got := ratio(0, 0, "no_records"); got.Known || got.Reason != "no_records" {
		t.Errorf("zero denominator ratio = %#v", got)
	}
}

func TestBillingMetricsDoNotConflateCapturedBytesAndCanonicalBytes(t *testing.T) {
	replay := completeReplay()
	c1Result := c1.Result{RecordsPresented: 10, RecordsAnalysed: 10, CounterfactualRecords: 2, InputCanonicalBytes: 100, CounterfactualBytes: 20}
	c2Result := c2.Result{}

	bytesResult := Build(replay, c1Result, c2Result, Options{BillingModel: model(MetricCapturedInputBytes)})
	if bytesResult.Billing.Baseline == nil || bytesResult.Billing.Baseline.Units != 50 || bytesResult.Billing.C1 != nil || bytesResult.Billing.Reason != "c1_counterfactual_not_derivable_for_captured_input_bytes" {
		t.Errorf("captured-byte billing = %#v", bytesResult.Billing)
	}
	canonicalResult := Build(replay, c1Result, c2Result, Options{BillingModel: model(MetricC1CanonicalBytes)})
	if canonicalResult.Billing.Baseline == nil || canonicalResult.Billing.C1 == nil || canonicalResult.Billing.Baseline.Units != 50 || canonicalResult.Billing.C1.CounterfactualUnits != 10 || canonicalResult.Billing.C2 != nil {
		t.Errorf("canonical billing = %#v", canonicalResult.Billing)
	}
}

func TestBuildPreservesZeroAndCompleteDuplicateOpportunitiesWithoutExtrapolation(t *testing.T) {
	zero := Build(input.ReplayResult{Coverage: input.Coverage{InputBytesKnown: true, RecordsObservedKnown: true}}, c1.Result{}, c2.Result{}, Options{BillingModel: model(MetricAcceptedRecords)})
	if zero.Billing.Baseline == nil || zero.Billing.Baseline.Units != 0 || zero.Billing.Baseline.Cost.MinorUnits != 0 || zero.Coverage.C1.Coverage.Known {
		t.Errorf("zero accounting = %#v", zero)
	}
	repeated := Build(completeReplay(), c1.Result{RecordsPresented: 10, RecordsAnalysed: 10, CounterfactualRecords: 1}, c2.Result{C1RepresentativeRecords: 1, RecordsAnalysed: 1, CandidateRecords: 10, InferredTemplateGroups: 1}, Options{BillingModel: model(MetricAcceptedRecords)})
	if got := repeated.Billing.C1; got == nil || got.BaselineUnits != 5 || got.CounterfactualUnits != 1 || got.UnitReduction != 4 {
		t.Errorf("complete duplicate opportunity = %#v", got)
	}
	if repeated.Billing.C1.Scope != "c1_analysed_subset" {
		t.Errorf("counterfactual scope was extrapolated: %#v", repeated.Billing.C1)
	}
}

func TestBuildReportsIncompatibleMeasuredBillingMetric(t *testing.T) {
	incompatible := model(MetricRetainedByteDays)
	result := Build(completeReplay(), c1.Result{}, c2.Result{}, Options{BillingModel: incompatible})
	if result.Billing.Available || result.Billing.Reason != "retained_volume_not_measured" || !result.Billing.ModelSupplied {
		t.Errorf("incompatible model result = %#v", result.Billing)
	}
	unknown := model(Metric("not_a_meter"))
	if err := unknown.Validate(); err == nil {
		t.Fatal("unknown billing metric was accepted")
	}
}

func TestTieredPricingAndZeroInputUseExactIntegerArithmetic(t *testing.T) {
	ten := uint64(10)
	model := Model{SchemaVersion: BillingSchemaVersion, Currency: "TST", Metric: MetricAcceptedRecords, IndexedFraction: Fraction{Numerator: 1, Denominator: 1}, Pricing: Pricing{Tiers: []Tier{
		{UpToUnits: &ten, Rate: Rate{UnitBlock: 1, MinorUnitsPerBlock: 1}},
		{Rate: Rate{UnitBlock: 1, MinorUnitsPerBlock: 2}},
	}}}
	if err := model.Validate(); err != nil {
		t.Fatal(err)
	}
	cost, err := model.price(15)
	if err != nil || cost != 20 {
		t.Fatalf("tiered price = %d, %v", cost, err)
	}
	zero, err := estimate(&model, 0, Measured, "complete_captured_input")
	if err != nil || zero.Units != 0 || zero.Cost.MinorUnits != 0 {
		t.Fatalf("zero estimate = %#v, %v", zero, err)
	}
}

func TestInvalidModelsAndArithmeticOverflowAreUnavailable(t *testing.T) {
	invalid := model(MetricAcceptedRecords)
	invalid.IndexedFraction = Fraction{Numerator: 2, Denominator: 1}
	if err := invalid.Validate(); err == nil {
		t.Fatal("invalid fraction was accepted")
	}
	overflow := model(MetricAcceptedRecords)
	overflow.Pricing.Fixed.MinorUnitsPerBlock = math.MaxInt64
	if _, err := estimate(overflow, 4, Measured, "test"); err == nil {
		t.Fatal("overflow was accepted")
	}
	broken := model(MetricAcceptedRecords)
	broken.Pricing.Fixed.UnitBlock = 0
	if err := broken.Validate(); err == nil {
		t.Fatal("zero unit block was accepted")
	}
	brokenCurrency := model(MetricAcceptedRecords)
	brokenCurrency.Currency = "bad"
	if err := brokenCurrency.Validate(); err == nil {
		t.Fatal("unsafe currency was accepted")
	}
}

func TestLoadModelJSONIsStrictBoundedAndPayloadFree(t *testing.T) {
	valid, err := json.Marshal(model(MetricAcceptedRecords))
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadModelJSON(bytes.NewReader(valid))
	if err != nil || loaded.Metric != MetricAcceptedRecords {
		t.Fatalf("valid model = %#v, %v", loaded, err)
	}
	for _, payload := range [][]byte{
		[]byte(`{"schema_version":"pulsefold/billing-model/v1","currency":"TST","metric":"accepted_records","indexed_fraction":{"numerator":1,"denominator":1},"pricing":{"fixed":{"unit_block":1,"minor_units_per_block":1}},"unknown":"Bearer pf_fake_secret"}`),
		[]byte(`{"schema_version":"pulsefold/billing-model/v1"}`),
		bytes.Repeat([]byte("x"), maxBillingConfigBytes+1),
	} {
		_, err := LoadModelJSON(bytes.NewReader(payload))
		if err == nil || strings.Contains(err.Error(), "pf_fake_secret") {
			t.Errorf("unsafe configuration result for %q: %v", payload[:min(len(payload), 16)], err)
		}
	}
}

func TestSyntheticBillingFixtures(t *testing.T) {
	tests := []struct {
		name   string
		metric Metric
	}{
		{name: "captured-bytes-fixed.json", metric: MetricCapturedInputBytes},
		{name: "records-indexed.json", metric: MetricAcceptedRecords},
		{name: "records-tiered.json", metric: MetricAcceptedRecords},
		{name: "retained-volume-unmeasured.json", metric: MetricRetainedByteDays},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload, err := os.ReadFile(filepath.Join("testdata", test.name))
			if err != nil {
				t.Fatal(err)
			}
			loaded, err := LoadModelJSON(bytes.NewReader(payload))
			if err != nil || loaded.Metric != test.metric {
				t.Fatalf("fixture model = %#v, %v", loaded, err)
			}
		})
	}
}

func BenchmarkBuildAccounting(b *testing.B) {
	replay := completeReplay()
	c1Result := c1.Result{RecordsPresented: 10_000, RecordsAnalysed: 10_000, CounterfactualRecords: 1_000, InputCanonicalBytes: 100_000, CounterfactualBytes: 10_000}
	c2Result := c2.Result{C1RepresentativeRecords: 1_000, RecordsAnalysed: 900, CandidateRecords: 10_000, InferredTemplateGroups: 700}
	options := Options{BillingModel: model(MetricAcceptedRecords)}
	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		_ = Build(replay, c1Result, c2Result, options)
	}
}

func completeReplay() input.ReplayResult {
	return input.ReplayResult{Coverage: input.Coverage{InputBytesRead: 100, InputBytesKnown: true, RecordsObserved: 10, RecordsObservedKnown: true, RecordsAccepted: 10}}
}

func model(metric Metric) *Model {
	return &Model{SchemaVersion: BillingSchemaVersion, Currency: "TST", Metric: metric, IndexedFraction: Fraction{Numerator: 1, Denominator: 2}, Pricing: Pricing{Fixed: &Rate{UnitBlock: 1, MinorUnitsPerBlock: 10}}}
}

func min(first, second int) int {
	if first < second {
		return first
	}
	return second
}
