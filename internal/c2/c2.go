// Package c2 implements bounded, report-only structural repetition analysis.
// It accepts only redacted records and C1 representatives. It never writes a
// template back to telemetry, extracts parameters or makes a retention choice.
package c2

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/amahon78/pulsefold/internal/c1"
	"github.com/amahon78/pulsefold/internal/redaction"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	"google.golang.org/protobuf/proto"
)

const (
	AlgorithmVersion          = "c2-conservative-drain-style-v1"
	CandidatePolicyVersion    = "c2-report-only-no-parameters-v1"
	DefaultMaxClusters        = 10_000
	DefaultMaxPartitions      = 10_000
	DefaultMaxStateBytes      = 16 << 20
	DefaultMaxBodyBytes       = 64 << 10
	DefaultMaxTemplateBytes   = 64 << 10
	DefaultMaxTokens          = 256
	DefaultMaxSourceTemplates = 10_000
)

var (
	ErrCancelled = errors.New("C2 analysis cancelled")
	uuidToken    = regexp.MustCompile(`(?i)^[{(]?[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}[)}]?[,;.]?$`)
	ipToken      = regexp.MustCompile(`^(?:\d{1,3}\.){3}\d{1,3}(?::\d+)?[,;.]?$`)
	timeToken    = regexp.MustCompile(`^\d{4}-\d\d-\d\d(?:T| )[0-9:.+-]+Z?[,;.]?$`)
	idValue      = regexp.MustCompile(`(?i)^(?:request|trace|span|session|user|transaction|correlation)[._-]?(?:id)?[=:].+$`)
	idPrefix     = regexp.MustCompile(`(?i)^(?:request|trace|span|session|user|transaction|correlation)[._-][a-z0-9-]+[,;.]?$`)
	inlineID     = regexp.MustCompile(`(?i)((?:request|trace|span|session|user|transaction|correlation)[._-]?(?:id)?[=:])[^,\s}\]]+`)
	inlineJSONID = regexp.MustCompile(`(?i)("(?:request|trace|span|session|user|transaction|correlation)[._-]?(?:id)?"\s*:\s*")[^"]+`)
	urlID        = regexp.MustCompile(`/(?:[0-9a-f]{8,}|[0-9]+)`)
	goLocation   = regexp.MustCompile(`\b[A-Za-z0-9_./-]+\.go:\d+\b`)
)

// Config has global bounds. C2 maintains no background state and the bounds
// apply to one finite replay only.
type Config struct {
	MaxClusters        int
	MaxPartitions      int
	MaxStateBytes      int
	MaxBodyBytes       int
	MaxTemplateBytes   int
	MaxTokens          int
	MaxSourceTemplates int
}

func (c Config) normalised() (Config, error) {
	if c.MaxClusters == 0 {
		c.MaxClusters = DefaultMaxClusters
	}
	if c.MaxPartitions == 0 {
		c.MaxPartitions = DefaultMaxPartitions
	}
	if c.MaxStateBytes == 0 {
		c.MaxStateBytes = DefaultMaxStateBytes
	}
	if c.MaxBodyBytes == 0 {
		c.MaxBodyBytes = DefaultMaxBodyBytes
	}
	if c.MaxTokens == 0 {
		c.MaxTokens = DefaultMaxTokens
	}
	if c.MaxTemplateBytes == 0 {
		c.MaxTemplateBytes = DefaultMaxTemplateBytes
	}
	if c.MaxSourceTemplates == 0 {
		c.MaxSourceTemplates = DefaultMaxSourceTemplates
	}
	if c.MaxClusters < 1 || c.MaxPartitions < 1 || c.MaxStateBytes < 1 || c.MaxBodyBytes < 1 || c.MaxTemplateBytes < 1 || c.MaxTokens < 1 || c.MaxSourceTemplates < 1 {
		return Config{}, errors.New("all C2 state limits must be positive")
	}
	return c, nil
}

// Result contains only counters, versions, bounds and opaque assignments for
// in-process evaluation. It deliberately carries no template or telemetry.
type Result struct {
	AlgorithmVersion         string
	CandidatePolicyVersion   string
	RecordsPresented         uint64
	C1RepresentativeRecords  uint64
	RecordsAnalysed          uint64
	CandidateRecords         uint64
	InferredTemplateGroups   uint64
	SourceTemplateRecords    uint64
	SourceTemplateGroups     uint64
	SourceTemplateExcluded   uint64
	AdditionalOpportunity    uint64
	ExcludedRecords          uint64
	ExcludedUnsupportedBody  uint64
	ExcludedOversizedBody    uint64
	ExcludedTooManyTokens    uint64
	ExcludedStateCap         uint64
	Partitions               uint64
	StateBytes               uint64
	StateCapReached          bool
	SourceTemplateCapReached bool
	SourceTemplateStateCap   bool
	MaxClusters              int
	MaxPartitions            int
	MaxStateBytes            int
	MaxBodyBytes             int
	MaxTemplateBytes         int
	MaxTokens                int
	MaxSourceTemplates       int
	Assignments              []Assignment
}

// Assignment links an input position to an opaque inferred-template ordinal.
// It is for deterministic local evaluation and is not serialised in reports.
type Assignment struct {
	InputIndex int
	Group      uint64
	Weight     uint64
}

// Analyse evaluates structural candidates after C1. C1 representatives are
// the only records presented to the learner, with their exact-group counts
// retained as weights. This preserves the fixed V0 stage order.
func Analyse(ctx context.Context, batch redaction.Batch, c1Result c1.Result, config Config) (Result, error) {
	config, err := config.normalised()
	if err != nil {
		return Result{}, err
	}
	result := Result{
		AlgorithmVersion: AlgorithmVersion, CandidatePolicyVersion: CandidatePolicyVersion,
		RecordsPresented: uint64(batch.Len()), C1RepresentativeRecords: uint64(len(c1Result.Representatives)),
		MaxClusters: config.MaxClusters, MaxPartitions: config.MaxPartitions, MaxStateBytes: config.MaxStateBytes,
		MaxBodyBytes: config.MaxBodyBytes, MaxTemplateBytes: config.MaxTemplateBytes, MaxTokens: config.MaxTokens, MaxSourceTemplates: config.MaxSourceTemplates,
	}
	partitions := make(map[string]uint64)
	clusters := make(map[string]uint64)
	sourceTemplates := make(map[string]struct{})

	for _, representative := range c1Result.Representatives {
		if err := ctx.Err(); err != nil {
			return result, fmt.Errorf("%w: %v", ErrCancelled, err)
		}
		if representative.InputIndex < 0 || representative.InputIndex >= batch.Len() {
			return result, errors.New("C1 representative index is outside the redacted batch")
		}
		record := batch.Record(representative.InputIndex)
		contextID, err := c1.ContextCanonical(record)
		if err != nil {
			result.excludeUnsupported(representative.Count)
			continue
		}
		_, _, _, _, log := record.Clone()
		if log == nil {
			result.excludeUnsupported(representative.Count)
			continue
		}
		partitionKey := partitionKey(contextID, uint64(log.SeverityNumber), log.SeverityText)
		partitionID, found := partitions[partitionKey]
		if !found {
			if len(partitions) >= config.MaxPartitions {
				result.excludeStateCap(representative.Count)
				continue
			}
			cost := uint64(len(contextID) + 96)
			if result.StateBytes+cost > uint64(config.MaxStateBytes) {
				result.excludeStateCap(representative.Count)
				continue
			}
			partitionID = uint64(len(partitions) + 1)
			partitions[partitionKey] = partitionID
			result.Partitions++
			result.StateBytes += cost
		}

		if template, ok := sourceTemplate(log.Attributes); ok {
			result.SourceTemplateRecords += representative.Count
			if len(template) > config.MaxTemplateBytes {
				result.SourceTemplateExcluded += representative.Count
			} else {
				key := strconv.FormatUint(partitionID, 10) + "\x00" + template
				if _, found := sourceTemplates[key]; !found {
					if len(sourceTemplates) >= config.MaxSourceTemplates {
						result.SourceTemplateCapReached = true
						result.SourceTemplateExcluded += representative.Count
					} else {
						cost := uint64(len(key) + 80)
						if result.StateBytes+cost > uint64(config.MaxStateBytes) {
							result.SourceTemplateStateCap = true
							result.SourceTemplateExcluded += representative.Count
						} else {
							sourceTemplates[key] = struct{}{}
							result.SourceTemplateGroups++
							result.StateBytes += cost
						}
					}
				}
			}
		}

		body, ok := stringBody(log.Body)
		if !ok {
			result.excludeUnsupported(representative.Count)
			continue
		}
		if len(body) > config.MaxBodyBytes {
			result.ExcludedRecords += representative.Count
			result.ExcludedOversizedBody += representative.Count
			continue
		}
		tokens := templateTokens(body)
		if len(tokens) == 0 || len(tokens) > config.MaxTokens {
			result.ExcludedRecords += representative.Count
			result.ExcludedTooManyTokens += representative.Count
			continue
		}
		template := strings.Join(tokens, " ")
		attributes, err := attributeShape(log.Attributes)
		if err != nil {
			result.excludeUnsupported(representative.Count)
			continue
		}
		clusterKey := strconv.FormatUint(partitionID, 10) + "\x00" + template + "\x00" + string(attributes)
		group, found := clusters[clusterKey]
		if !found {
			if len(clusters) >= config.MaxClusters {
				result.excludeStateCap(representative.Count)
				continue
			}
			cost := uint64(len(clusterKey) + 80)
			if result.StateBytes+cost > uint64(config.MaxStateBytes) {
				result.excludeStateCap(representative.Count)
				continue
			}
			group = uint64(len(clusters) + 1)
			clusters[clusterKey] = group
			result.InferredTemplateGroups++
			result.StateBytes += cost
		}
		result.RecordsAnalysed++
		result.CandidateRecords += representative.Count
		result.Assignments = append(result.Assignments, Assignment{InputIndex: representative.InputIndex, Group: group, Weight: representative.Count})
	}
	if result.CandidateRecords >= result.InferredTemplateGroups {
		result.AdditionalOpportunity = result.CandidateRecords - result.InferredTemplateGroups
	}
	return result, nil
}

func (r *Result) excludeUnsupported(weight uint64) {
	r.ExcludedRecords += weight
	r.ExcludedUnsupportedBody += weight
}

func (r *Result) excludeStateCap(weight uint64) {
	r.ExcludedRecords += weight
	r.ExcludedStateCap += weight
	r.StateCapReached = true
}

func partitionKey(contextID []byte, severity uint64, severityText string) string {
	return string(contextID) + "\x00" + fmt.Sprintf("%d:%s", severity, severityText)
}

func sourceTemplate(attributes []*commonv1.KeyValue) (string, bool) {
	for _, attribute := range attributes {
		if attribute != nil && attribute.Key == "log.record.template" && attribute.Value != nil {
			if value, ok := attribute.Value.Value.(*commonv1.AnyValue_StringValue); ok && value != nil && value.StringValue != "" {
				return value.StringValue, true
			}
		}
	}
	return "", false
}

func stringBody(value *commonv1.AnyValue) (string, bool) {
	if value == nil {
		return "", false
	}
	typed, ok := value.Value.(*commonv1.AnyValue_StringValue)
	if !ok || typed == nil {
		return "", false
	}
	return typed.StringValue, true
}

// templateTokens is deliberately more conservative than a general-purpose
// Drain cluster merger. It masks only strongly recognisable dynamic values.
// Values such as HTTP status or error codes remain literal, preventing a
// same-shape message from silently crossing a semantic boundary.
func templateTokens(body string) []string {
	body = inlineJSONID.ReplaceAllString(body, "${1}<*>")
	body = inlineID.ReplaceAllString(body, "${1}<*>")
	body = urlID.ReplaceAllString(body, "/<*>")
	body = goLocation.ReplaceAllString(body, "<*>")
	tokens := strings.Fields(body)
	for index, token := range tokens {
		if isDynamicToken(token) {
			tokens[index] = "<*>"
		}
	}
	return tokens
}

func isDynamicToken(token string) bool {
	trimmed := strings.Trim(token, "[](){}\"'")
	if uuidToken.MatchString(trimmed) || ipToken.MatchString(trimmed) || timeToken.MatchString(trimmed) || idValue.MatchString(trimmed) || idPrefix.MatchString(trimmed) {
		return true
	}
	return false
}

func attributeShape(attributes []*commonv1.KeyValue) ([]byte, error) {
	entries := make([][]byte, 0, len(attributes))
	for _, attribute := range attributes {
		if attribute == nil || attribute.Key == "log.record.template" {
			continue
		}
		copy := proto.Clone(attribute).(*commonv1.KeyValue)
		if isDynamicAttribute(copy.Key) {
			copy.Value = &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "<*>"}}
		}
		encoded, err := proto.MarshalOptions{Deterministic: true}.Marshal(copy)
		if err != nil {
			return nil, err
		}
		entries = append(entries, encoded)
	}
	sort.Slice(entries, func(first, second int) bool { return bytes.Compare(entries[first], entries[second]) < 0 })
	var output []byte
	for _, entry := range entries {
		output = append(output, entry...)
		output = append(output, 0)
	}
	return output, nil
}

func isDynamicAttribute(key string) bool {
	switch strings.ToLower(key) {
	case "request.id", "request_id", "trace.id", "trace_id", "span.id", "span_id", "user.id", "user_id", "session.id", "session_id", "transaction.id", "transaction_id", "correlation.id", "correlation_id":
		return true
	default:
		return false
	}
}

// SameAssignments reports whether two runs assigned the same records to the
// same inferred groups, modulo independently allocated group ordinals.
func SameAssignments(first, second Result) bool {
	if len(first.Assignments) != len(second.Assignments) {
		return false
	}
	firstByIndex := make(map[int]uint64, len(first.Assignments))
	secondByIndex := make(map[int]uint64, len(second.Assignments))
	for _, assignment := range first.Assignments {
		firstByIndex[assignment.InputIndex] = assignment.Group
	}
	for _, assignment := range second.Assignments {
		secondByIndex[assignment.InputIndex] = assignment.Group
	}
	return equivalentPartitions(firstByIndex, secondByIndex)
}

func equivalentPartitions(first, second map[int]uint64) bool {
	if len(first) != len(second) {
		return false
	}
	for leftIndex, leftGroup := range first {
		rightGroup, ok := second[leftIndex]
		if !ok {
			return false
		}
		for otherIndex, otherLeftGroup := range first {
			otherRightGroup := second[otherIndex]
			if (leftGroup == otherLeftGroup) != (rightGroup == otherRightGroup) {
				return false
			}
		}
	}
	return true
}

// CanonicalTemplate returns an opaque-safe equality check used by tests. It
// never logs its input and is not included in a report.
func CanonicalTemplate(body string) []byte {
	return bytes.Clone([]byte(strings.Join(templateTokens(body), " ")))
}
