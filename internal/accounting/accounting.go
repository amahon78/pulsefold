// Package accounting derives coverage and optional billing opportunity
// estimates from a completed shadow replay. It never observes production
// forwarding, wire compression or a vendor meter.
package accounting

import (
	"errors"
	"math"
	"math/bits"

	"github.com/amahon78/pulsefold/internal/c1"
	"github.com/amahon78/pulsefold/internal/c2"
	"github.com/amahon78/pulsefold/internal/input"
)

// Classification describes what supports a reported value.
type Classification string

const (
	Measured       Classification = "measured"
	Counterfactual Classification = "counterfactual"
	Inferred       Classification = "inferred"
	UserSupplied   Classification = "user_supplied"
	Unknown        Classification = "unknown"
)

// Metric identifies the operator-supplied meter mapping. It is deliberately
// generic and does not encode a vendor's pricing or contract semantics.
type Metric string

const (
	MetricCapturedInputBytes Metric = "captured_input_bytes"
	MetricAcceptedRecords    Metric = "accepted_records"
	MetricC1CanonicalBytes   Metric = "c1_canonical_bytes"
	// These recognised metrics require measurements that V0 does not collect.
	// They are accepted so a report can state that limitation explicitly.
	MetricWireBytes        Metric = "wire_bytes"
	MetricRetainedByteDays Metric = "retained_byte_days"
)

const BillingSchemaVersion = "pulsefold/billing-model/v1"

// Ratio is unavailable rather than zero when no complete, non-zero
// denominator exists. It intentionally contains no floating-point percentage.
type Ratio struct {
	Numerator   uint64 `json:"numerator"`
	Denominator uint64 `json:"denominator"`
	Known       bool   `json:"known"`
	Reason      string `json:"reason,omitempty"`
}

// InputCoverage describes what the isolated replay boundary could observe.
type InputCoverage struct {
	CapturedInputBytes      uint64 `json:"captured_input_bytes"`
	CapturedInputBytesKnown bool   `json:"captured_input_bytes_known"`
	RecordsObserved         uint64 `json:"records_observed"`
	RecordsObservedKnown    bool   `json:"records_observed_known"`
	RecordsAccepted         uint64 `json:"records_accepted"`
	RecordsRejected         uint64 `json:"records_rejected"`
	MalformedRecords        uint64 `json:"malformed_records"`
	UnsupportedRecords      uint64 `json:"unsupported_records"`
	MalformedInput          uint64 `json:"malformed_input"`
	UnsupportedInput        uint64 `json:"unsupported_input"`
	CaptureErrors           uint64 `json:"capture_errors"`
	InputByteLimitHit       bool   `json:"input_byte_limit_hit"`
	AcceptedRecordCoverage  Ratio  `json:"accepted_record_coverage"`
}

// StageCoverage describes one analyser stage without pretending that a state
// overflow or unsupported value was analysed.
type StageCoverage struct {
	Eligible  uint64 `json:"eligible"`
	Analysed  uint64 `json:"analysed"`
	Excluded  uint64 `json:"excluded"`
	Coverage  Ratio  `json:"coverage"`
	Completed bool   `json:"completed"`
}

// Coverage is a first-class description of input and analyser reach.
type Coverage struct {
	Input InputCoverage `json:"input"`
	C1    StageCoverage `json:"c1"`
	C2    StageCoverage `json:"c2"`
}

// Wire is intentionally unknown until a concrete link, protocol, batching
// and compression measurement is supplied by a future measurement adapter.
type Wire struct {
	Classification Classification `json:"classification"`
	Known          bool           `json:"known"`
	Reason         string         `json:"reason"`
}

// Money stores an exact minor-unit amount. Floating-point currency is never
// used in the report or billing arithmetic.
type Money struct {
	Currency   string `json:"currency"`
	MinorUnits int64  `json:"minor_units"`
}

// Fraction determines what proportion of a generic meter is indexed. It is
// applied with ceiling arithmetic so it never understates a billable unit.
type Fraction struct {
	Numerator   uint64 `json:"numerator"`
	Denominator uint64 `json:"denominator"`
}

// Rate is an exact fixed price per positive block of meter units.
type Rate struct {
	UnitBlock          uint64 `json:"unit_block"`
	MinorUnitsPerBlock int64  `json:"minor_units_per_block"`
}

// Tier applies a Rate to units up to UpToUnits. A nil UpToUnits is the final,
// unbounded tier and must be last.
type Tier struct {
	UpToUnits *uint64 `json:"up_to_units,omitempty"`
	Rate      Rate    `json:"rate"`
}

// Pricing supports either one fixed rate or ordered marginal tiers.
type Pricing struct {
	Fixed *Rate  `json:"fixed,omitempty"`
	Tiers []Tier `json:"tiers,omitempty"`
}

// Model is supplied by an operator. It represents a local mapping of a
// contract meter, not a vendor integration or a universal price list.
type Model struct {
	SchemaVersion   string   `json:"schema_version"`
	Name            string   `json:"name,omitempty"`
	Currency        string   `json:"currency"`
	Metric          Metric   `json:"metric"`
	IndexedFraction Fraction `json:"indexed_fraction"`
	Pricing         Pricing  `json:"pricing"`
}

// Options controls optional accounting. A nil BillingModel is normal and
// produces an explicit unavailable result rather than a zero estimate.
type Options struct {
	BillingModel *Model
}

// Estimate is a quantity priced using a validated operator model.
type Estimate struct {
	Units          uint64         `json:"units"`
	Classification Classification `json:"classification"`
	Scope          string         `json:"scope"`
	Cost           Money          `json:"cost"`
}

// Opportunity compares a modelled input with its counterfactual output. It
// is never described as a net customer saving because PulseFold costs,
// retention and contract side effects are not represented here.
type Opportunity struct {
	Classification       Classification `json:"classification"`
	BaselineUnits        uint64         `json:"baseline_units"`
	CounterfactualUnits  uint64         `json:"counterfactual_units"`
	UnitReduction        uint64         `json:"unit_reduction"`
	GrossCostOpportunity Money          `json:"gross_observability_cost_opportunity"`
	Scope                string         `json:"scope"`
}

// Billing is either an operator-supplied, validated calculation or an
// explicit unavailable result. It does not report credentials or telemetry.
type Billing struct {
	ModelSupplied  bool           `json:"model_supplied"`
	Classification Classification `json:"classification"`
	Available      bool           `json:"available"`
	Reason         string         `json:"reason,omitempty"`
	Metric         Metric         `json:"metric,omitempty"`
	Baseline       *Estimate      `json:"baseline,omitempty"`
	C1             *Opportunity   `json:"c1,omitempty"`
	C2             *Opportunity   `json:"c2,omitempty"`
}

// Result contains only counts, classifications and optional exact money.
// It intentionally cannot contain telemetry bodies, attribute values or
// identities.
type Result struct {
	Coverage Coverage `json:"coverage"`
	Wire     Wire     `json:"wire"`
	Billing  Billing  `json:"billing"`
}

// FromInput produces coverage for a replay that did not complete analysis.
func FromInput(replay input.ReplayResult) Result {
	inputCoverage := InputCoverage{
		CapturedInputBytes: replay.Coverage.InputBytesRead, CapturedInputBytesKnown: replay.Coverage.InputBytesKnown,
		RecordsObserved: replay.Coverage.RecordsObserved, RecordsObservedKnown: replay.Coverage.RecordsObservedKnown,
		RecordsAccepted: replay.Coverage.RecordsAccepted, RecordsRejected: replay.Coverage.RecordsRejected,
		MalformedRecords: replay.Coverage.MalformedRecords, UnsupportedRecords: replay.Coverage.UnsupportedRecords,
		MalformedInput: replay.Coverage.MalformedInput, UnsupportedInput: replay.Coverage.UnsupportedInput,
		CaptureErrors: replay.Coverage.CaptureErrors, InputByteLimitHit: replay.Coverage.InputByteLimitHit,
	}
	if replay.Coverage.RecordsObservedKnown {
		inputCoverage.AcceptedRecordCoverage = ratio(replay.Coverage.RecordsAccepted, replay.Coverage.RecordsObserved, "no_records_observed")
	} else {
		inputCoverage.AcceptedRecordCoverage = Ratio{Known: false, Reason: "records_observed_unknown"}
	}
	return Result{
		Coverage: Coverage{
			Input: inputCoverage,
			C1:    StageCoverage{Coverage: Ratio{Known: false, Reason: "c1_not_completed"}},
			C2:    StageCoverage{Coverage: Ratio{Known: false, Reason: "c2_not_completed"}},
		},
		Wire:    Wire{Classification: Unknown, Known: false, Reason: "wire_bytes_not_measured"},
		Billing: Billing{Classification: Unknown, Available: false, Reason: "billing_model_not_supplied"},
	}
}

// Build derives complete-stage coverage and applies an optional validated
// billing model. Invalid or incompatible models remain unavailable.
func Build(replay input.ReplayResult, c1Result c1.Result, c2Result c2.Result, options Options) Result {
	result := FromInput(replay)
	result.Coverage.C1 = stage(c1Result.RecordsPresented, c1Result.RecordsAnalysed, c1Result.ExcludedRecords, true, "")
	c2Excluded := subtract(c2Result.C1RepresentativeRecords, c2Result.RecordsAnalysed)
	result.Coverage.C2 = stage(c2Result.C1RepresentativeRecords, c2Result.RecordsAnalysed, c2Excluded, true, "")
	result.Billing = calculateBilling(replay, c1Result, c2Result, options.BillingModel)
	return result
}

func stage(eligible, analysed, excluded uint64, complete bool, reason string) StageCoverage {
	return StageCoverage{Eligible: eligible, Analysed: analysed, Excluded: excluded, Completed: complete, Coverage: ratio(analysed, eligible, reasonOr(reason, "no_eligible_records"))}
}

func ratio(numerator, denominator uint64, zeroReason string) Ratio {
	if denominator == 0 {
		return Ratio{Numerator: numerator, Denominator: denominator, Known: false, Reason: zeroReason}
	}
	return Ratio{Numerator: numerator, Denominator: denominator, Known: true}
}

func reasonOr(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func calculateBilling(replay input.ReplayResult, c1Result c1.Result, c2Result c2.Result, model *Model) Billing {
	if model == nil {
		return Billing{Classification: Unknown, Available: false, Reason: "billing_model_not_supplied"}
	}
	if err := model.Validate(); err != nil {
		return Billing{ModelSupplied: true, Classification: UserSupplied, Available: false, Reason: "billing_model_invalid"}
	}
	billing := Billing{ModelSupplied: true, Classification: UserSupplied, Available: true, Metric: model.Metric}
	switch model.Metric {
	case MetricCapturedInputBytes:
		if !replay.Coverage.InputBytesKnown {
			billing.Available = false
			billing.Reason = "captured_input_bytes_unknown"
			return billing
		}
		baseline, err := estimate(model, replay.Coverage.InputBytesRead, Measured, "complete_captured_input")
		if err != nil {
			return unavailableBilling(err)
		}
		billing.Baseline = &baseline
		billing.Reason = "c1_counterfactual_not_derivable_for_captured_input_bytes"
	case MetricAcceptedRecords:
		if !replay.Coverage.RecordsObservedKnown {
			billing.Available = false
			billing.Reason = "accepted_record_baseline_unknown"
			return billing
		}
		baseline, err := estimate(model, replay.Coverage.RecordsAccepted, Measured, "complete_captured_input")
		if err != nil {
			return unavailableBilling(err)
		}
		billing.Baseline = &baseline
		c1Opportunity, err := opportunity(model, c1Result.RecordsAnalysed, c1Result.CounterfactualRecords, Counterfactual, "c1_analysed_subset")
		if err != nil {
			return unavailableBilling(err)
		}
		billing.C1 = &c1Opportunity
		c2Opportunity, err := opportunity(model, c2Result.CandidateRecords, c2Result.InferredTemplateGroups, Inferred, "c2_candidate_subset")
		if err != nil {
			return unavailableBilling(err)
		}
		billing.C2 = &c2Opportunity
	case MetricC1CanonicalBytes:
		baseline, err := estimate(model, c1Result.InputCanonicalBytes, Measured, "c1_analysed_subset")
		if err != nil {
			return unavailableBilling(err)
		}
		billing.Baseline = &baseline
		c1Opportunity, err := opportunity(model, c1Result.InputCanonicalBytes, c1Result.CounterfactualBytes, Counterfactual, "c1_analysed_subset")
		if err != nil {
			return unavailableBilling(err)
		}
		billing.C1 = &c1Opportunity
		billing.Reason = "c2_counterfactual_bytes_not_derivable"
	case MetricWireBytes:
		billing.Available = false
		billing.Reason = "wire_bytes_not_measured"
	case MetricRetainedByteDays:
		billing.Available = false
		billing.Reason = "retained_volume_not_measured"
	default:
		billing.Available = false
		billing.Reason = "billing_metric_unsupported"
	}
	return billing
}

func unavailableBilling(err error) Billing {
	_ = err
	return Billing{ModelSupplied: true, Classification: UserSupplied, Available: false, Reason: "billing_calculation_overflow"}
}

func estimate(model *Model, quantity uint64, classification Classification, scope string) (Estimate, error) {
	units, err := indexedUnits(quantity, model.IndexedFraction)
	if err != nil {
		return Estimate{}, err
	}
	cost, err := model.price(units)
	if err != nil {
		return Estimate{}, err
	}
	return Estimate{Units: units, Classification: classification, Scope: scope, Cost: Money{Currency: model.Currency, MinorUnits: cost}}, nil
}

func opportunity(model *Model, baseline, counterfactual uint64, classification Classification, scope string) (Opportunity, error) {
	baselineUnits, err := indexedUnits(baseline, model.IndexedFraction)
	if err != nil {
		return Opportunity{}, err
	}
	counterfactualUnits, err := indexedUnits(counterfactual, model.IndexedFraction)
	if err != nil {
		return Opportunity{}, err
	}
	baselineCost, err := model.price(baselineUnits)
	if err != nil {
		return Opportunity{}, err
	}
	counterfactualCost, err := model.price(counterfactualUnits)
	if err != nil {
		return Opportunity{}, err
	}
	return Opportunity{
		Classification: classification, BaselineUnits: baselineUnits, CounterfactualUnits: counterfactualUnits,
		UnitReduction:        subtract(baselineUnits, counterfactualUnits),
		GrossCostOpportunity: Money{Currency: model.Currency, MinorUnits: subtractSigned(baselineCost, counterfactualCost)}, Scope: scope,
	}, nil
}

// Validate rejects configurations that could silently turn an unsupported
// meter, impossible fraction or unsafe currency calculation into an estimate.
func (m Model) Validate() error {
	if m.SchemaVersion != BillingSchemaVersion || !validCurrency(m.Currency) {
		return errors.New("invalid billing model schema")
	}
	if m.Metric != MetricCapturedInputBytes && m.Metric != MetricAcceptedRecords && m.Metric != MetricC1CanonicalBytes && m.Metric != MetricWireBytes && m.Metric != MetricRetainedByteDays {
		return errors.New("unsupported billing metric")
	}
	if m.IndexedFraction.Denominator == 0 || m.IndexedFraction.Numerator > m.IndexedFraction.Denominator {
		return errors.New("invalid indexed fraction")
	}
	if (m.Pricing.Fixed == nil) == (len(m.Pricing.Tiers) == 0) {
		return errors.New("pricing must specify exactly one fixed rate or tier list")
	}
	if m.Pricing.Fixed != nil {
		return validRate(*m.Pricing.Fixed)
	}
	var previous uint64
	for index, tier := range m.Pricing.Tiers {
		if err := validRate(tier.Rate); err != nil {
			return err
		}
		if tier.UpToUnits == nil {
			if index != len(m.Pricing.Tiers)-1 {
				return errors.New("unbounded tier must be last")
			}
			continue
		}
		if *tier.UpToUnits == 0 || *tier.UpToUnits <= previous {
			return errors.New("tier bounds must increase")
		}
		previous = *tier.UpToUnits
		if index == len(m.Pricing.Tiers)-1 {
			return errors.New("last tier must be unbounded")
		}
	}
	return nil
}

func validCurrency(value string) bool {
	if len(value) != 3 {
		return false
	}
	for _, character := range value {
		if character < 'A' || character > 'Z' {
			return false
		}
	}
	return true
}

func validRate(rate Rate) error {
	if rate.UnitBlock == 0 || rate.MinorUnitsPerBlock < 0 {
		return errors.New("invalid billing rate")
	}
	return nil
}

func (m Model) price(units uint64) (int64, error) {
	if m.Pricing.Fixed != nil {
		return priceRate(units, *m.Pricing.Fixed)
	}
	var total int64
	var lower uint64
	remaining := units
	for _, tier := range m.Pricing.Tiers {
		if remaining == 0 {
			break
		}
		span := remaining
		if tier.UpToUnits != nil {
			width := *tier.UpToUnits - lower
			if span > width {
				span = width
			}
			lower = *tier.UpToUnits
		}
		cost, err := priceRate(span, tier.Rate)
		if err != nil || total > math.MaxInt64-cost {
			return 0, errors.New("billing cost overflow")
		}
		total += cost
		remaining -= span
	}
	if remaining != 0 {
		return 0, errors.New("billing tier unavailable")
	}
	return total, nil
}

func priceRate(units uint64, rate Rate) (int64, error) {
	blocks, err := ceilDiv(units, rate.UnitBlock)
	if err != nil {
		return 0, err
	}
	hi, lo := bits.Mul64(blocks, uint64(rate.MinorUnitsPerBlock))
	if hi != 0 || lo > math.MaxInt64 {
		return 0, errors.New("billing cost overflow")
	}
	return int64(lo), nil
}

func indexedUnits(quantity uint64, fraction Fraction) (uint64, error) {
	whole := quantity / fraction.Denominator
	remainder := quantity % fraction.Denominator
	hi, lo := bits.Mul64(whole, fraction.Numerator)
	if hi != 0 {
		return 0, errors.New("billing quantity overflow")
	}
	partial, err := ceilDivProduct(remainder, fraction.Numerator, fraction.Denominator)
	if err != nil || lo > math.MaxUint64-partial {
		return 0, errors.New("billing quantity overflow")
	}
	return lo + partial, nil
}

func ceilDivProduct(a, b, divisor uint64) (uint64, error) {
	hi, lo := bits.Mul64(a, b)
	if hi >= divisor {
		return 0, errors.New("billing quantity overflow")
	}
	quotient, remainder := bits.Div64(hi, lo, divisor)
	if remainder > 0 {
		if quotient == math.MaxUint64 {
			return 0, errors.New("billing quantity overflow")
		}
		quotient++
	}
	return quotient, nil
}

func ceilDiv(value, divisor uint64) (uint64, error) {
	if divisor == 0 {
		return 0, errors.New("division by zero")
	}
	if value == 0 {
		return 0, nil
	}
	return 1 + (value-1)/divisor, nil
}

func subtract(first, second uint64) uint64 {
	if first < second {
		return 0
	}
	return first - second
}

func subtractSigned(first, second int64) int64 {
	if first < second {
		return 0
	}
	return first - second
}
