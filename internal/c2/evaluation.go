package c2

// GroundTruth is synthetic benchmark labelling. It is intentionally separate
// from analysis and must never be inferred from a C2 result.
type GroundTruth struct {
	Group string
	Novel bool
	Error bool
}

// Metrics measures grouping quality for truth-labelled synthetic data. Rates
// are zero when their denominator is zero.
type Metrics struct {
	EvaluatedRecords      uint64
	Precision             float64
	Recall                float64
	F1                    float64
	UnsafeMergeGroups     uint64
	UnsafeMergeRecordRate float64
	SplitTruthGroups      uint64
	SplitTruthRecordRate  float64
	NovelPreservationRate float64
	ErrorPreservationRate float64
}

// Evaluate compares opaque C2 assignments with independently supplied truth.
// It uses weighted pairwise grouping quality so C1 exact-group counts remain
// represented after C2 receives only C1 representatives.
func Evaluate(result Result, truth map[int]GroundTruth) Metrics {
	predicted := map[uint64]map[string]uint64{}
	actual := map[string]map[uint64]uint64{}
	var evaluated uint64
	for _, assignment := range result.Assignments {
		label, ok := truth[assignment.InputIndex]
		if !ok || label.Group == "" {
			continue
		}
		if predicted[assignment.Group] == nil {
			predicted[assignment.Group] = map[string]uint64{}
		}
		if actual[label.Group] == nil {
			actual[label.Group] = map[uint64]uint64{}
		}
		predicted[assignment.Group][label.Group] += assignment.Weight
		actual[label.Group][assignment.Group] += assignment.Weight
		evaluated += assignment.Weight
	}

	var truePositive, predictedPairs, actualPairs, unsafeRecords, splitRecords uint64
	for _, labels := range predicted {
		var total uint64
		for _, count := range labels {
			total += count
			truePositive += pairs(count)
		}
		predictedPairs += pairs(total)
		if len(labels) > 1 {
			unsafeRecords += total
		}
	}
	var splitGroups uint64
	for _, groups := range actual {
		var total uint64
		for _, count := range groups {
			total += count
		}
		actualPairs += pairs(total)
		if len(groups) > 1 {
			splitGroups++
			splitRecords += total
		}
	}
	metrics := Metrics{EvaluatedRecords: evaluated, UnsafeMergeRecordRate: rate(unsafeRecords, evaluated), SplitTruthGroups: splitGroups, SplitTruthRecordRate: rate(splitRecords, evaluated)}
	for _, labels := range predicted {
		if len(labels) > 1 {
			metrics.UnsafeMergeGroups++
		}
	}
	metrics.Precision = rate(truePositive, predictedPairs)
	metrics.Recall = rate(truePositive, actualPairs)
	if metrics.Precision+metrics.Recall > 0 {
		metrics.F1 = 2 * metrics.Precision * metrics.Recall / (metrics.Precision + metrics.Recall)
	}
	metrics.NovelPreservationRate = preservationRate(result.Assignments, truth, predicted, func(value GroundTruth) bool { return value.Novel })
	metrics.ErrorPreservationRate = preservationRate(result.Assignments, truth, predicted, func(value GroundTruth) bool { return value.Error })
	return metrics
}

func preservationRate(assignments []Assignment, truth map[int]GroundTruth, predicted map[uint64]map[string]uint64, include func(GroundTruth) bool) float64 {
	var preserved, total uint64
	for _, assignment := range assignments {
		label, ok := truth[assignment.InputIndex]
		if !ok || !include(label) {
			continue
		}
		total += assignment.Weight
		if len(predicted[assignment.Group]) == 1 {
			preserved += assignment.Weight
		}
	}
	return rate(preserved, total)
}

func pairs(count uint64) uint64 {
	if count < 2 {
		return 0
	}
	return count * (count - 1) / 2
}

func rate(numerator, denominator uint64) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}
