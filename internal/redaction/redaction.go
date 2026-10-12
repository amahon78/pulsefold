// Package redaction is the mandatory boundary between untrusted OTLP input and
// every PulseFold-derived representation.
package redaction

import (
	"regexp"

	"github.com/amahon78/pulsefold/internal/input"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcev1 "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/protobuf/proto"
)

const PolicyVersion = "v1-fixture-safety"

var (
	secretKey           = regexp.MustCompile(`(?i)(authorization|cookie|password|secret|token|api[_-]?key|credential)`)
	secretValuePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\bbearer\s+[a-z0-9._~+/=-]+`),
		regexp.MustCompile(`(?i)\b(?:api[_-]?key|password|secret|token)\s*[=:]\s*[^\s,;]+`),
		regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`),
	}
)

// Batch is the only telemetry batch accepted by analysis packages. Its record
// fields are private so callers cannot accidentally pass raw OTLP to analysis.
type Batch struct {
	Source        input.SourceMetadata
	Coverage      input.Coverage
	PolicyVersion string
	records       []Record
}

// Record contains an already-redacted OTLP log. The stored messages are never
// returned directly; Clone returns independent messages to prevent analyser
// mutation of the redaction result or the original input.
type Record struct {
	resource          *resourcev1.Resource
	resourceSchemaURL string
	scope             *commonv1.InstrumentationScope
	scopeSchemaURL    string
	log               *logsv1.LogRecord
}

// Apply clones and redacts every accepted input record. It is deterministic
// and has no logging side effects.
func Apply(replay input.ReplayResult) Batch {
	batch := Batch{
		Source:        replay.Source,
		Coverage:      replay.Coverage,
		PolicyVersion: PolicyVersion,
		records:       make([]Record, 0, len(replay.Records)),
	}
	for _, raw := range replay.Records {
		resource := cloneResource(raw.Resource)
		scope := cloneScope(raw.Scope)
		log := cloneLog(raw.Record)
		redactResource(resource)
		redactScope(scope)
		redactLog(log)
		batch.records = append(batch.records, Record{
			resource:          resource,
			resourceSchemaURL: raw.ResourceSchemaURL,
			scope:             scope,
			scopeSchemaURL:    raw.ScopeSchemaURL,
			log:               log,
		})
	}
	return batch
}

func (b Batch) Len() int { return len(b.records) }

func (b Batch) Record(index int) Record { return b.records[index] }

// Clone returns a private snapshot suitable for deterministic derivation.
func (r Record) Clone() (*resourcev1.Resource, string, *commonv1.InstrumentationScope, string, *logsv1.LogRecord) {
	return cloneResource(r.resource), r.resourceSchemaURL, cloneScope(r.scope), r.scopeSchemaURL, cloneLog(r.log)
}

func cloneResource(value *resourcev1.Resource) *resourcev1.Resource {
	if value == nil {
		return nil
	}
	return proto.Clone(value).(*resourcev1.Resource)
}

func cloneScope(value *commonv1.InstrumentationScope) *commonv1.InstrumentationScope {
	if value == nil {
		return nil
	}
	return proto.Clone(value).(*commonv1.InstrumentationScope)
}

func cloneLog(value *logsv1.LogRecord) *logsv1.LogRecord {
	if value == nil {
		return nil
	}
	return proto.Clone(value).(*logsv1.LogRecord)
}

func redactResource(resource *resourcev1.Resource) {
	if resource != nil {
		redactAttributes(resource.Attributes)
	}
}

func redactScope(scope *commonv1.InstrumentationScope) {
	if scope != nil {
		redactAttributes(scope.Attributes)
	}
}

func redactLog(log *logsv1.LogRecord) {
	if log == nil {
		return
	}
	redactAttributes(log.Attributes)
	redactValue(log.Body)
}

func redactAttributes(attributes []*commonv1.KeyValue) {
	for _, attribute := range attributes {
		if attribute == nil {
			continue
		}
		if secretKey.MatchString(attribute.Key) {
			attribute.Value = stringValue("[REDACTED:ATTRIBUTE]")
			continue
		}
		redactValue(attribute.Value)
	}
}

func redactValue(value *commonv1.AnyValue) {
	if value == nil {
		return
	}
	switch typed := value.Value.(type) {
	case *commonv1.AnyValue_StringValue:
		typed.StringValue = redactString(typed.StringValue)
	case *commonv1.AnyValue_ArrayValue:
		if typed.ArrayValue == nil {
			return
		}
		for _, entry := range typed.ArrayValue.Values {
			redactValue(entry)
		}
	case *commonv1.AnyValue_KvlistValue:
		if typed.KvlistValue != nil {
			redactAttributes(typed.KvlistValue.Values)
		}
	}
}

func redactString(value string) string {
	for _, pattern := range secretValuePatterns {
		if pattern.MatchString(value) {
			return "[REDACTED:VALUE]"
		}
	}
	return value
}

func stringValue(value string) *commonv1.AnyValue {
	return &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: value}}
}
