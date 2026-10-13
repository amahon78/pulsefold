package report

import (
	"fmt"
	"io"

	"github.com/amahon78/pulsefold/internal/accounting"
)

// WriteText writes a deterministic terminal-oriented V0 shadow summary. It
// exposes aggregate report facts only and never renders telemetry content.
func WriteText(writer io.Writer, value Report) error {
	if _, err := fmt.Fprintln(writer, "PulseFold V0 shadow analysis"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(writer, "Status: %s\n", value.Status); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(writer, "Source: %s (%s, %s)\n", value.Source.Identity, value.Source.Kind, value.Source.Format); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "\nCoverage:"); err != nil {
		return err
	}
	inputCoverage := value.Accounting.Coverage.Input
	if inputCoverage.RecordsObservedKnown {
		if _, err := fmt.Fprintf(writer, "  Accepted records: %d of %d observed\n", inputCoverage.RecordsAccepted, inputCoverage.RecordsObserved); err != nil {
			return err
		}
	} else if _, err := fmt.Fprintln(writer, "  Accepted records: denominator unknown"); err != nil {
		return err
	}
	if inputCoverage.CapturedInputBytesKnown {
		if _, err := fmt.Fprintf(writer, "  Captured input bytes: %d\n", inputCoverage.CapturedInputBytes); err != nil {
			return err
		}
	} else if _, err := fmt.Fprintln(writer, "  Captured input bytes: unknown"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(writer, "  Rejected records: %d (malformed %d, unsupported %d)\n", inputCoverage.RecordsRejected, inputCoverage.MalformedRecords, inputCoverage.UnsupportedRecords); err != nil {
		return err
	}
	if value.C1 != nil {
		if _, err := fmt.Fprintln(writer, "\nC1 exact repetition (counterfactual):"); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(writer, "  Analysed records: %d of %d eligible\n  Exact duplicate records: %d\n  Canonical-byte opportunity: %d of %d canonical bytes\n", value.Accounting.Coverage.C1.Analysed, value.Accounting.Coverage.C1.Eligible, value.C1.DuplicateRecords, value.C1.RawByteOpportunity, value.C1.InputCanonicalBytes); err != nil {
			return err
		}
	}
	if value.C2 != nil {
		if _, err := fmt.Fprintln(writer, "\nC2 structural repetition (inferred):"); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(writer, "  Analysed representatives: %d of %d eligible\n  Candidate records beyond C1: %d\n  Inferred template groups: %d\n", value.Accounting.Coverage.C2.Analysed, value.Accounting.Coverage.C2.Eligible, value.C2.AdditionalOpportunity, value.C2.InferredTemplateGroups); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(writer, "\nEconomics:"); err != nil {
		return err
	}
	if err := writeEconomics(writer, value.Accounting.Billing); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "\nProvenance:"); err != nil {
		return err
	}
	if value.Manifest == nil {
		if _, err := fmt.Fprintln(writer, "  Integrity: unavailable"); err != nil {
			return err
		}
	} else if err := VerifyIntegrity(value); err != nil {
		if _, err := fmt.Fprintln(writer, "  Integrity: invalid"); err != nil {
			return err
		}
	} else if _, err := fmt.Fprintf(writer, "  Manifest: %s\n  Integrity: verified SHA-256 consistency check\n  Authentication: not provided\n", value.Manifest.ManifestID); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "\nLimitations:"); err != nil {
		return err
	}
	for _, limitation := range value.Limitations {
		if _, err := fmt.Fprintf(writer, "  - %s\n", limitation); err != nil {
			return err
		}
	}
	return nil
}

func writeEconomics(writer io.Writer, billing accounting.Billing) error {
	if !billing.Available {
		_, err := fmt.Fprintf(writer, "  Estimated gross observability cost opportunity: unavailable (%s)\n", billing.Reason)
		return err
	}
	if _, err := fmt.Fprintf(writer, "  Meter: %s\n", billing.Metric); err != nil {
		return err
	}
	if billing.Baseline != nil {
		if _, err := fmt.Fprintf(writer, "  Baseline: %d units, %s %d minor units\n", billing.Baseline.Units, billing.Baseline.Cost.Currency, billing.Baseline.Cost.MinorUnits); err != nil {
			return err
		}
	}
	for _, entry := range []struct {
		label       string
		opportunity *accounting.Opportunity
	}{
		{label: "C1 counterfactual opportunity", opportunity: billing.C1},
		{label: "C2 inferred potential", opportunity: billing.C2},
	} {
		if entry.opportunity == nil {
			continue
		}
		if _, err := fmt.Fprintf(writer, "  %s: %d units, %s %d minor units\n", entry.label, entry.opportunity.UnitReduction, entry.opportunity.GrossCostOpportunity.Currency, entry.opportunity.GrossCostOpportunity.MinorUnits); err != nil {
			return err
		}
	}
	return nil
}
