// Package c1 implements bounded, report-only exact-duplicate opportunity
// analysis. It never alters or suppresses its input.
package c1

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/amahon78/pulsefold/internal/redaction"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcev1 "go.opentelemetry.io/proto/otlp/resource/v1"
)

const (
	CanonicalisationVersion = "c1-otlp-log-v1"
	CandidatePolicyVersion  = "c1-exact-duplicate-report-v1"
	DefaultMaxGroups        = 100_000
	DefaultMaxStateBytes    = 64 << 20
	DefaultMaxIdentityBytes = 1 << 20
)

var ErrCancelled = errors.New("C1 analysis cancelled")

// Config gives every stateful C1 structure explicit bounds. Hash is injectable
// for collision tests; equality is always checked against canonical bytes.
type Config struct {
	MaxGroups        int
	MaxStateBytes    int
	MaxIdentityBytes int
	Hash             func([]byte) string
}

func (c Config) normalised() (Config, error) {
	if c.MaxGroups == 0 {
		c.MaxGroups = DefaultMaxGroups
	}
	if c.MaxStateBytes == 0 {
		c.MaxStateBytes = DefaultMaxStateBytes
	}
	if c.MaxIdentityBytes == 0 {
		c.MaxIdentityBytes = DefaultMaxIdentityBytes
	}
	if c.MaxGroups < 1 || c.MaxStateBytes < 1 || c.MaxIdentityBytes < 1 {
		return Config{}, errors.New("all C1 state limits must be positive")
	}
	if c.Hash == nil {
		c.Hash = defaultHash
	}
	return c, nil
}

// Result contains only counters and versions. It deliberately contains no raw
// telemetry, canonical identity or representative body.
type Result struct {
	CanonicalisationVersion string
	CandidatePolicyVersion  string
	RecordsPresented        uint64
	RecordsAnalysed         uint64
	UniqueIdentities        uint64
	DuplicateGroups         uint64
	DuplicateRecords        uint64
	CounterfactualRecords   uint64
	InputCanonicalBytes     uint64
	CounterfactualBytes     uint64
	RawByteOpportunity      uint64
	ExcludedRecords         uint64
	ExcludedOversized       uint64
	ExcludedStateCap        uint64
	StateBytes              uint64
	StateCapReached         bool
}

type group struct {
	identity []byte
	count    uint64
	bytes    uint64
}

// Analyse calculates the result for one finite replay batch. The replay call
// defines the analysis window; C1 itself has no implicit clock or goroutine.
func Analyse(ctx context.Context, batch redaction.Batch, config Config) (Result, error) {
	config, err := config.normalised()
	if err != nil {
		return Result{}, err
	}
	result := Result{
		CanonicalisationVersion: CanonicalisationVersion,
		CandidatePolicyVersion:  CandidatePolicyVersion,
		RecordsPresented:        uint64(batch.Len()),
	}
	buckets := make(map[string][]*group)
	groupCount := 0

	for index := 0; index < batch.Len(); index++ {
		if err := ctx.Err(); err != nil {
			return finalise(result, buckets), fmt.Errorf("%w: %v", ErrCancelled, err)
		}
		identity, err := Canonical(batch.Record(index))
		if err != nil {
			// Input validation ensures this can only be an unsupported derived
			// representation. Exclude conservatively rather than guessing.
			result.ExcludedRecords++
			result.ExcludedOversized++
			continue
		}
		if len(identity) > config.MaxIdentityBytes {
			result.ExcludedRecords++
			result.ExcludedOversized++
			continue
		}

		bucketKey := config.Hash(identity)
		matched := false
		for _, candidate := range buckets[bucketKey] {
			if bytes.Equal(candidate.identity, identity) {
				candidate.count++
				result.RecordsAnalysed++
				result.InputCanonicalBytes += uint64(len(identity))
				matched = true
				break
			}
		}
		if matched {
			continue
		}

		stateCost := uint64(len(identity) + 128)
		if groupCount >= config.MaxGroups || result.StateBytes+stateCost > uint64(config.MaxStateBytes) {
			result.ExcludedRecords++
			result.ExcludedStateCap++
			result.StateCapReached = true
			continue
		}
		identityCopy := append([]byte(nil), identity...)
		buckets[bucketKey] = append(buckets[bucketKey], &group{identity: identityCopy, count: 1, bytes: uint64(len(identityCopy))})
		groupCount++
		result.RecordsAnalysed++
		result.InputCanonicalBytes += uint64(len(identityCopy))
		result.StateBytes += stateCost
	}
	return finalise(result, buckets), nil
}

func finalise(result Result, buckets map[string][]*group) Result {
	for _, bucket := range buckets {
		for _, group := range bucket {
			result.UniqueIdentities++
			result.CounterfactualRecords++
			result.CounterfactualBytes += group.bytes
			if group.count > 1 {
				result.DuplicateGroups++
				result.DuplicateRecords += group.count - 1
			}
		}
	}
	if result.InputCanonicalBytes >= result.CounterfactualBytes {
		result.RawByteOpportunity = result.InputCanonicalBytes - result.CounterfactualBytes
	}
	return result
}

func defaultHash(identity []byte) string {
	sum := sha256.Sum256(identity)
	return string(sum[:])
}

// Canonical returns a complete, deterministic C1 identity for a redacted log.
// Timestamp and observed timestamp are intentionally excluded because C1
// compares recurring log events over a replay window. All remaining log,
// resource and scope context, including type distinctions, is retained.
func Canonical(record redaction.Record) ([]byte, error) {
	resource, resourceSchema, scope, scopeSchema, log := record.Clone()
	if log == nil {
		return nil, errors.New("missing redacted log")
	}
	encoder := canonicalEncoder{}
	encoder.string("pulsefold-c1")
	encoder.resource(resource, resourceSchema)
	encoder.scope(scope, scopeSchema)
	encoder.log(log)
	if encoder.err != nil {
		return nil, encoder.err
	}
	return encoder.buffer.Bytes(), nil
}

type canonicalEncoder struct {
	buffer bytes.Buffer
	err    error
}

func (e *canonicalEncoder) marker(value byte) { e.buffer.WriteByte(value) }

func (e *canonicalEncoder) uint(value uint64) {
	var encoded [10]byte
	size := binary.PutUvarint(encoded[:], value)
	e.buffer.Write(encoded[:size])
}

func (e *canonicalEncoder) string(value string) {
	e.uint(uint64(len(value)))
	e.buffer.WriteString(value)
}

func (e *canonicalEncoder) bytes(value []byte) {
	e.uint(uint64(len(value)))
	e.buffer.Write(value)
}

func (e *canonicalEncoder) resource(resource *resourcev1.Resource, schemaURL string) {
	if resource == nil {
		e.marker(0)
		return
	}
	e.marker(1)
	e.string(schemaURL)
	e.uint(uint64(resource.DroppedAttributesCount))
	e.attributes(resource.Attributes)
}

func (e *canonicalEncoder) scope(scope *commonv1.InstrumentationScope, schemaURL string) {
	if scope == nil {
		e.marker(0)
		return
	}
	e.marker(1)
	e.string(schemaURL)
	e.string(scope.Name)
	e.string(scope.Version)
	e.uint(uint64(scope.DroppedAttributesCount))
	e.attributes(scope.Attributes)
}

func (e *canonicalEncoder) log(log *logsv1.LogRecord) {
	e.uint(uint64(log.SeverityNumber))
	e.string(log.SeverityText)
	if log.Body == nil {
		e.marker(0)
	} else {
		e.marker(1)
		e.value(log.Body)
	}
	e.attributes(log.Attributes)
	e.uint(uint64(log.DroppedAttributesCount))
	e.uint(uint64(log.Flags))
	e.bytes(log.TraceId)
	e.bytes(log.SpanId)
	e.string(log.EventName)
}

func (e *canonicalEncoder) attributes(attributes []*commonv1.KeyValue) {
	entries := make([][]byte, 0, len(attributes))
	for _, attribute := range attributes {
		entry := canonicalEncoder{}
		if attribute == nil {
			entry.marker(0)
		} else {
			entry.marker(1)
			entry.string(attribute.Key)
			entry.value(attribute.Value)
		}
		if entry.err != nil {
			e.err = entry.err
			return
		}
		entries = append(entries, entry.buffer.Bytes())
	}
	sort.Slice(entries, func(first, second int) bool { return bytes.Compare(entries[first], entries[second]) < 0 })
	e.uint(uint64(len(entries)))
	for _, entry := range entries {
		e.bytes(entry)
	}
}

func (e *canonicalEncoder) value(value *commonv1.AnyValue) {
	if value == nil || value.Value == nil {
		e.err = errors.New("unsupported empty AnyValue")
		return
	}
	switch typed := value.Value.(type) {
	case *commonv1.AnyValue_StringValue:
		e.marker(1)
		e.string(typed.StringValue)
	case *commonv1.AnyValue_BoolValue:
		e.marker(2)
		if typed.BoolValue {
			e.marker(1)
		} else {
			e.marker(0)
		}
	case *commonv1.AnyValue_IntValue:
		e.marker(3)
		e.uint(uint64(typed.IntValue<<1) ^ uint64(typed.IntValue>>63))
	case *commonv1.AnyValue_DoubleValue:
		e.marker(4)
		e.uint(math.Float64bits(typed.DoubleValue))
	case *commonv1.AnyValue_ArrayValue:
		if typed.ArrayValue == nil {
			e.err = errors.New("unsupported nil array")
			return
		}
		e.marker(5)
		e.uint(uint64(len(typed.ArrayValue.Values)))
		for _, entry := range typed.ArrayValue.Values {
			e.value(entry)
		}
	case *commonv1.AnyValue_KvlistValue:
		if typed.KvlistValue == nil {
			e.err = errors.New("unsupported nil key-value list")
			return
		}
		e.marker(6)
		e.attributes(typed.KvlistValue.Values)
	case *commonv1.AnyValue_BytesValue:
		e.marker(7)
		e.bytes(typed.BytesValue)
	default:
		e.err = errors.New("unsupported AnyValue type")
	}
}
