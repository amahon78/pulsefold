// Package input defines the untrusted OTLP input boundary.
//
// It deliberately exposes raw OTLP records only to the immediately following
// redaction boundary. Analysis packages must consume redacted records instead.
package input

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	collectlogsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcev1 "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	// FormatOTLPJSON is an OTLP JSON ExportLogsServiceRequest encoded with the
	// protobuf JSON mapping. One file contains exactly one request.
	FormatOTLPJSON = "otlp-json-v1"

	// DefaultMaxInputBytes bounds a replay before OTLP decoding. It prevents an
	// offline input from allocating unbounded memory while preserving a useful
	// local replay size.
	DefaultMaxInputBytes int64 = 64 << 20
)

var (
	ErrMalformedInput   = errors.New("malformed OTLP log input")
	ErrUnsupportedInput = errors.New("unsupported OTLP log input")
	ErrInputTooLarge    = errors.New("OTLP log input exceeds configured byte limit")
)

// SourceKind describes the source delivery contract, not its transport.
type SourceKind string

const (
	SourceOfflineReplay     SourceKind = "offline_replay"
	SourceBestEffortReplica SourceKind = "best_effort_replica"
)

// SourceMetadata is retained in coverage and later report provenance. A
// replica must explicitly declare itself non-authoritative.
type SourceMetadata struct {
	Identity         string
	Partition        string
	Kind             SourceKind
	Format           string
	NonAuthoritative bool
}

// Source supplies a bounded input stream. The interface intentionally does
// not imply a network listener or a production forwarding role.
type Source interface {
	Metadata() SourceMetadata
	Open(context.Context) (io.ReadCloser, error)
}

// FileSource is the offline replay implementation. Identity should be an
// operator-safe label, not a telemetry value.
type FileSource struct {
	Path string
	Info SourceMetadata
}

func (s FileSource) Metadata() SourceMetadata {
	info := s.Info
	if info.Identity == "" {
		info.Identity = filepath.Base(s.Path)
	}
	if info.Kind == "" {
		info.Kind = SourceOfflineReplay
	}
	if info.Format == "" {
		info.Format = FormatOTLPJSON
	}
	return info
}

func (s FileSource) Open(context.Context) (io.ReadCloser, error) {
	return os.Open(s.Path)
}

// ReaderSource allows callers and tests to replay a known reader without
// defining a transport. It is not safe to reuse concurrently.
type ReaderSource struct {
	Reader io.Reader
	Info   SourceMetadata
}

func (s ReaderSource) Metadata() SourceMetadata { return s.Info }

func (s ReaderSource) Open(context.Context) (io.ReadCloser, error) {
	if s.Reader == nil {
		return nil, errors.New("input reader is nil")
	}
	return io.NopCloser(s.Reader), nil
}

// Config controls bounded replay behaviour.
type Config struct {
	MaxInputBytes int64
}

func (c Config) normalised() (Config, error) {
	if c.MaxInputBytes == 0 {
		c.MaxInputBytes = DefaultMaxInputBytes
	}
	if c.MaxInputBytes < 1 {
		return Config{}, fmt.Errorf("max input bytes must be positive")
	}
	return c, nil
}

// Coverage is conservative. Rejected, malformed, unsupported, oversized and
// failed input is never counted as accepted or analysed.
type Coverage struct {
	RecordsObserved    uint64
	RecordsAccepted    uint64
	RecordsRejected    uint64
	MalformedRecords   uint64
	MalformedInput     uint64
	UnsupportedRecords uint64
	UnsupportedInput   uint64
	CaptureErrors      uint64
	InputBytesRead     uint64
	InputByteLimitHit  bool
	Interval           Interval
}

// Interval is derived only from non-zero log timestamps. It never substitutes
// wall-clock processing time, keeping offline replay deterministic.
type Interval struct {
	Start time.Time
	End   time.Time
	Known bool
}

// RawLog is an untrusted record. It is solely the hand-off type from input to
// redaction. No analyser package should accept RawLog.
type RawLog struct {
	Resource          *resourcev1.Resource
	ResourceSchemaURL string
	Scope             *commonv1.InstrumentationScope
	ScopeSchemaURL    string
	Record            *logsv1.LogRecord
}

// ReplayResult is the complete outcome of a source attempt. Coverage is
// populated even when a format, source or cancellation error is returned.
type ReplayResult struct {
	Source   SourceMetadata
	Coverage Coverage
	Records  []RawLog
}

// Replay reads a single bounded OTLP JSON ExportLogsServiceRequest. It never
// writes diagnostics containing telemetry bodies or attributes.
func Replay(ctx context.Context, source Source, config Config) (ReplayResult, error) {
	config, err := config.normalised()
	if err != nil {
		return ReplayResult{}, err
	}
	if source == nil {
		return ReplayResult{Coverage: Coverage{CaptureErrors: 1}}, errors.New("input source is nil")
	}

	result := ReplayResult{Source: source.Metadata()}
	if err := validateSource(result.Source); err != nil {
		if errors.Is(err, ErrUnsupportedInput) {
			result.Coverage.UnsupportedInput++
		}
		return result, err
	}
	if err := ctx.Err(); err != nil {
		result.Coverage.CaptureErrors++
		return result, err
	}

	reader, err := source.Open(ctx)
	if err != nil {
		result.Coverage.CaptureErrors++
		return result, fmt.Errorf("open input source: %w", err)
	}
	payload, bytesRead, err := readBounded(ctx, reader, config.MaxInputBytes)
	result.Coverage.InputBytesRead = bytesRead
	closeErr := reader.Close()
	if err != nil {
		if errors.Is(err, ErrInputTooLarge) {
			result.Coverage.InputByteLimitHit = true
		} else {
			result.Coverage.CaptureErrors++
		}
		return result, err
	}
	if closeErr != nil {
		result.Coverage.CaptureErrors++
		return result, errors.New("close input source")
	}
	if len(bytes.TrimSpace(payload)) == 0 {
		return result, nil
	}

	request := new(collectlogsv1.ExportLogsServiceRequest)
	if err := protojson.Unmarshal(payload, request); err != nil {
		result.Coverage.CaptureErrors++
		result.Coverage.MalformedInput++
		return result, ErrMalformedInput
	}

	for _, resourceLogs := range request.ResourceLogs {
		if resourceLogs == nil {
			result.Coverage.UnsupportedInput++
			continue
		}
		for _, scopeLogs := range resourceLogs.ScopeLogs {
			if scopeLogs == nil {
				result.Coverage.UnsupportedInput++
				continue
			}
			for _, record := range scopeLogs.LogRecords {
				if err := ctx.Err(); err != nil {
					result.Coverage.CaptureErrors++
					return result, err
				}
				result.Coverage.RecordsObserved++
				classification := validateLog(resourceLogs.Resource, scopeLogs.Scope, record)
				switch classification {
				case recordAccepted:
					result.Records = append(result.Records, RawLog{
						Resource:          resourceLogs.Resource,
						ResourceSchemaURL: resourceLogs.SchemaUrl,
						Scope:             scopeLogs.Scope,
						ScopeSchemaURL:    scopeLogs.SchemaUrl,
						Record:            record,
					})
					result.Coverage.RecordsAccepted++
					result.Coverage.Interval.observe(record.TimeUnixNano)
				case recordMalformed:
					result.Coverage.RecordsRejected++
					result.Coverage.MalformedRecords++
				case recordUnsupported:
					result.Coverage.RecordsRejected++
					result.Coverage.UnsupportedRecords++
				}
			}
		}
	}
	return result, nil
}

func validateSource(source SourceMetadata) error {
	if source.Identity == "" {
		return fmt.Errorf("%w: source identity is required", ErrUnsupportedInput)
	}
	if source.Format != FormatOTLPJSON {
		return fmt.Errorf("%w: format %q", ErrUnsupportedInput, source.Format)
	}
	switch source.Kind {
	case SourceOfflineReplay:
		return nil
	case SourceBestEffortReplica:
		if source.NonAuthoritative {
			return nil
		}
		return fmt.Errorf("%w: replicated input must be explicitly non-authoritative", ErrUnsupportedInput)
	default:
		return fmt.Errorf("%w: source kind %q", ErrUnsupportedInput, source.Kind)
	}
}

func readBounded(ctx context.Context, reader io.Reader, maximum int64) ([]byte, uint64, error) {
	var output bytes.Buffer
	buffer := make([]byte, 32*1024)
	for {
		if err := ctx.Err(); err != nil {
			return nil, uint64(output.Len()), err
		}
		count, err := reader.Read(buffer)
		if count > 0 {
			if int64(output.Len()+count) > maximum {
				remaining := maximum - int64(output.Len())
				if remaining > 0 {
					_, _ = output.Write(buffer[:remaining])
				}
				return nil, uint64(output.Len()) + 1, ErrInputTooLarge
			}
			_, _ = output.Write(buffer[:count])
		}
		if errors.Is(err, io.EOF) {
			return output.Bytes(), uint64(output.Len()), nil
		}
		if err != nil {
			return nil, uint64(output.Len()), fmt.Errorf("read input source: %w", err)
		}
	}
}

func (i *Interval) observe(timestamp uint64) {
	if timestamp == 0 {
		return
	}
	observed := time.Unix(0, int64(timestamp)).UTC()
	if !i.Known || observed.Before(i.Start) {
		i.Start = observed
	}
	if !i.Known || observed.After(i.End) {
		i.End = observed
	}
	i.Known = true
}

type recordClassification uint8

const (
	recordAccepted recordClassification = iota
	recordMalformed
	recordUnsupported
)

func validateLog(resource *resourcev1.Resource, scope *commonv1.InstrumentationScope, record *logsv1.LogRecord) recordClassification {
	if record == nil {
		return recordMalformed
	}
	if (resource != nil && !validAttributes(resource.Attributes, 0)) ||
		(scope != nil && !validAttributes(scope.Attributes, 0)) ||
		!validAttributes(record.Attributes, 0) {
		return recordUnsupported
	}
	if record.Body != nil && !validValue(record.Body, 0) {
		return recordUnsupported
	}
	return recordAccepted
}

const maxValueDepth = 32

func validAttributes(attributes []*commonv1.KeyValue, depth int) bool {
	for _, attribute := range attributes {
		if attribute == nil || attribute.Key == "" || !validValue(attribute.Value, depth) {
			return false
		}
	}
	return true
}

func validValue(value *commonv1.AnyValue, depth int) bool {
	if value == nil || value.Value == nil || depth > maxValueDepth {
		return false
	}
	switch value := value.Value.(type) {
	case *commonv1.AnyValue_ArrayValue:
		if value.ArrayValue == nil {
			return false
		}
		for _, entry := range value.ArrayValue.Values {
			if !validValue(entry, depth+1) {
				return false
			}
		}
	case *commonv1.AnyValue_KvlistValue:
		if value.KvlistValue == nil {
			return false
		}
		return validAttributes(value.KvlistValue.Values, depth+1)
	}
	return true
}
