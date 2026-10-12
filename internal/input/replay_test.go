package input

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	collectlogsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcev1 "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestReplayAcceptsStructuredLogsAcrossResources(t *testing.T) {
	request := &collectlogsv1.ExportLogsServiceRequest{ResourceLogs: []*logsv1.ResourceLogs{
		resourceLogs("payments", []*logsv1.LogRecord{
			logRecord(10, stringValue("payment accepted")),
			logRecord(20, mapValue("id", intValue(7))),
		}),
		resourceLogs("orders", []*logsv1.LogRecord{
			logRecord(30, arrayValue(stringValue("queued"), boolValue(true))),
		}),
	}}

	result, err := Replay(context.Background(), testSource(request), Config{})
	if err != nil {
		t.Fatalf("Replay() error = %v", err)
	}
	if got, want := result.Coverage.RecordsObserved, uint64(3); got != want {
		t.Errorf("observed = %d, want %d", got, want)
	}
	if got, want := result.Coverage.RecordsAccepted, uint64(3); got != want {
		t.Errorf("accepted = %d, want %d", got, want)
	}
	if !result.Coverage.Interval.Known || !result.Coverage.Interval.Start.Equal(time.Unix(0, 10).UTC()) || !result.Coverage.Interval.End.Equal(time.Unix(0, 30).UTC()) {
		t.Errorf("interval = %#v, want [10,30]ns", result.Coverage.Interval)
	}
	if got, want := result.Records[0].Resource.Attributes[0].Value.GetStringValue(), "payments"; got != want {
		t.Errorf("first service = %q, want %q", got, want)
	}
	if got := result.Records[1].Record.Body.GetKvlistValue().Values[0].Value.GetIntValue(); got != 7 {
		t.Errorf("structured body integer = %d, want 7", got)
	}
}

func TestReplayEmptyInputIsSuccessfulAndCovered(t *testing.T) {
	result, err := Replay(context.Background(), ReaderSource{
		Reader: strings.NewReader(" \n\t"),
		Info:   testMetadata(),
	}, Config{})
	if err != nil {
		t.Fatalf("Replay() error = %v", err)
	}
	if result.Coverage.RecordsObserved != 0 || result.Coverage.RecordsAccepted != 0 || result.Coverage.InputBytesRead != 3 {
		t.Errorf("unexpected empty coverage: %#v", result.Coverage)
	}
}

func TestReplayMalformedInputDoesNotExposeTelemetry(t *testing.T) {
	secret := "Bearer not-a-real-secret-value"
	result, err := Replay(context.Background(), ReaderSource{
		Reader: strings.NewReader(`{"resourceLogs":` + secret),
		Info:   testMetadata(),
	}, Config{})
	if !errors.Is(err, ErrMalformedInput) {
		t.Fatalf("Replay() error = %v, want malformed input", err)
	}
	if result.Coverage.CaptureErrors != 1 || result.Coverage.MalformedInput != 1 || result.Coverage.RecordsAccepted != 0 {
		t.Errorf("coverage = %#v", result.Coverage)
	}
	if strings.Contains(err.Error(), secret) {
		t.Error("malformed-input error leaked raw telemetry")
	}
}

func TestReplayRejectsUnsupportedRecordValues(t *testing.T) {
	request := &collectlogsv1.ExportLogsServiceRequest{ResourceLogs: []*logsv1.ResourceLogs{
		resourceLogs("payments", []*logsv1.LogRecord{
			logRecord(10, stringValue("accepted")),
			{TimeUnixNano: 11, Attributes: []*commonv1.KeyValue{{Key: "broken", Value: &commonv1.AnyValue{}}}},
			{TimeUnixNano: 12, Attributes: []*commonv1.KeyValue{{Key: "", Value: stringValue("invalid")}}},
		}),
	}}
	result, err := Replay(context.Background(), testSource(request), Config{})
	if err != nil {
		t.Fatalf("Replay() error = %v", err)
	}
	if got, want := result.Coverage.RecordsObserved, uint64(3); got != want {
		t.Errorf("observed = %d, want %d", got, want)
	}
	if got, want := result.Coverage.RecordsAccepted, uint64(1); got != want {
		t.Errorf("accepted = %d, want %d", got, want)
	}
	if got, want := result.Coverage.UnsupportedRecords, uint64(2); got != want {
		t.Errorf("unsupported = %d, want %d", got, want)
	}
}

func TestReplayRejectsUnsupportedSourceContracts(t *testing.T) {
	request := &collectlogsv1.ExportLogsServiceRequest{}
	source := testSource(request)
	source.Info.Format = "otlp-protobuf"
	result, err := Replay(context.Background(), source, Config{})
	if !errors.Is(err, ErrUnsupportedInput) {
		t.Fatalf("Replay() error = %v, want unsupported input", err)
	}
	if result.Coverage.UnsupportedInput != 1 {
		t.Errorf("unsupported input = %d, want 1", result.Coverage.UnsupportedInput)
	}

	source = testSource(request)
	source.Info.Kind = SourceBestEffortReplica
	source.Info.NonAuthoritative = false
	if _, err := Replay(context.Background(), source, Config{}); !errors.Is(err, ErrUnsupportedInput) {
		t.Fatalf("authoritative replica error = %v, want unsupported input", err)
	}
}

func TestReplayCoverageIsDeterministic(t *testing.T) {
	request := &collectlogsv1.ExportLogsServiceRequest{ResourceLogs: []*logsv1.ResourceLogs{
		resourceLogs("payments", []*logsv1.LogRecord{logRecord(20, stringValue("same")), logRecord(10, stringValue("same"))}),
	}}
	first, err := Replay(context.Background(), testSource(request), Config{})
	if err != nil {
		t.Fatal(err)
	}
	second, err := Replay(context.Background(), testSource(request), Config{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first.Coverage, second.Coverage) {
		t.Errorf("coverage differs:\nfirst:  %#v\nsecond: %#v", first.Coverage, second.Coverage)
	}
}

func TestReplayHonoursInputLimitAndCancellation(t *testing.T) {
	request := &collectlogsv1.ExportLogsServiceRequest{ResourceLogs: []*logsv1.ResourceLogs{
		resourceLogs("payments", []*logsv1.LogRecord{logRecord(10, stringValue(strings.Repeat("x", 2048)))}),
	}}
	result, err := Replay(context.Background(), testSource(request), Config{MaxInputBytes: 32})
	if !errors.Is(err, ErrInputTooLarge) {
		t.Fatalf("Replay() error = %v, want input too large", err)
	}
	if !result.Coverage.InputByteLimitHit || result.Coverage.RecordsAccepted != 0 {
		t.Errorf("limited replay coverage = %#v", result.Coverage)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, err = Replay(ctx, testSource(request), Config{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Replay() cancellation error = %v, want context cancelled", err)
	}
	if result.Coverage.CaptureErrors != 1 || result.Coverage.RecordsAccepted != 0 {
		t.Errorf("cancelled replay coverage = %#v", result.Coverage)
	}
}

func testSource(request *collectlogsv1.ExportLogsServiceRequest) ReaderSource {
	payload, err := protojson.Marshal(request)
	if err != nil {
		panic(err)
	}
	return ReaderSource{Reader: bytes.NewReader(payload), Info: testMetadata()}
}

func testMetadata() SourceMetadata {
	return SourceMetadata{
		Identity:  "synthetic-fixture",
		Partition: "test",
		Kind:      SourceOfflineReplay,
		Format:    FormatOTLPJSON,
	}
}

func resourceLogs(service string, records []*logsv1.LogRecord) *logsv1.ResourceLogs {
	return &logsv1.ResourceLogs{
		Resource: &resourcev1.Resource{Attributes: []*commonv1.KeyValue{{Key: "service.name", Value: stringValue(service)}}},
		ScopeLogs: []*logsv1.ScopeLogs{{
			Scope:      &commonv1.InstrumentationScope{Name: "pulsefold.test"},
			LogRecords: records,
		}},
	}
}

func logRecord(timestamp uint64, body *commonv1.AnyValue) *logsv1.LogRecord {
	return &logsv1.LogRecord{TimeUnixNano: timestamp, Body: body}
}

func stringValue(value string) *commonv1.AnyValue {
	return &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: value}}
}

func intValue(value int64) *commonv1.AnyValue {
	return &commonv1.AnyValue{Value: &commonv1.AnyValue_IntValue{IntValue: value}}
}

func boolValue(value bool) *commonv1.AnyValue {
	return &commonv1.AnyValue{Value: &commonv1.AnyValue_BoolValue{BoolValue: value}}
}

func mapValue(key string, value *commonv1.AnyValue) *commonv1.AnyValue {
	return &commonv1.AnyValue{Value: &commonv1.AnyValue_KvlistValue{KvlistValue: &commonv1.KeyValueList{Values: []*commonv1.KeyValue{{Key: key, Value: value}}}}}
}

func arrayValue(values ...*commonv1.AnyValue) *commonv1.AnyValue {
	return &commonv1.AnyValue{Value: &commonv1.AnyValue_ArrayValue{ArrayValue: &commonv1.ArrayValue{Values: values}}}
}
