package c1

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/amahon78/pulsefold/internal/fixture"
	"github.com/amahon78/pulsefold/internal/input"
	"github.com/amahon78/pulsefold/internal/redaction"
	collectlogsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcev1 "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestAnalyseFindsOnlySafeExactDuplicateOpportunities(t *testing.T) {
	batch := fixtureBatch(t, fixture.Config{Seed: 7, HighCardinalityCount: 4, LargeBodyBytes: 128})
	result, err := Analyse(context.Background(), batch, Config{})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := result.RecordsPresented, uint64(20); got != want {
		t.Errorf("records presented = %d, want %d", got, want)
	}
	if got, want := result.DuplicateGroups, uint64(3); got != want {
		t.Errorf("duplicate groups = %d, want %d", got, want)
	}
	if got, want := result.DuplicateRecords, uint64(4); got != want {
		t.Errorf("duplicate records = %d, want %d", got, want)
	}
	if result.CounterfactualRecords != result.UniqueIdentities || result.RawByteOpportunity == 0 {
		t.Errorf("invalid counterfactual result: %#v", result)
	}
}

func TestAnalysePreservesSemanticDifferencesAndCanonicalisesMaps(t *testing.T) {
	batch := batchForRequest(t, &collectlogsv1.ExportLogsServiceRequest{ResourceLogs: []*logsv1.ResourceLogs{
		resourceLogs("payments", []*logsv1.LogRecord{
			record(1, logsv1.SeverityNumber_SEVERITY_NUMBER_INFO, stringValue("same"), attrs("type", stringValue("1"))),
			record(2, logsv1.SeverityNumber_SEVERITY_NUMBER_ERROR, stringValue("same"), attrs("type", stringValue("1"))),
			record(3, logsv1.SeverityNumber_SEVERITY_NUMBER_INFO, mapValue(attrs("a", intValue(1), "b", stringValue("two"))), nil),
			record(4, logsv1.SeverityNumber_SEVERITY_NUMBER_INFO, mapValue(attrs("b", stringValue("two"), "a", intValue(1))), nil),
			record(5, logsv1.SeverityNumber_SEVERITY_NUMBER_INFO, stringValue("typed"), attrs("type", stringValue("1"))),
			record(6, logsv1.SeverityNumber_SEVERITY_NUMBER_INFO, stringValue("typed"), attrs("type", intValue(1))),
		}),
		resourceLogs("orders", []*logsv1.LogRecord{
			record(7, logsv1.SeverityNumber_SEVERITY_NUMBER_INFO, stringValue("same"), attrs("type", stringValue("1"))),
		}),
	}})
	result, err := Analyse(context.Background(), batch, Config{})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := result.DuplicateGroups, uint64(1); got != want {
		t.Errorf("duplicate groups = %d, want reordered maps only", got)
	}
	if got, want := result.DuplicateRecords, uint64(1); got != want {
		t.Errorf("duplicate records = %d, want 1", got)
	}
}

func TestAnalyseHandlesMissingBodyAsAnExplicitValue(t *testing.T) {
	batch := batchForRequest(t, &collectlogsv1.ExportLogsServiceRequest{ResourceLogs: []*logsv1.ResourceLogs{
		resourceLogs("payments", []*logsv1.LogRecord{
			record(1, logsv1.SeverityNumber_SEVERITY_NUMBER_INFO, nil, nil),
			record(2, logsv1.SeverityNumber_SEVERITY_NUMBER_INFO, nil, nil),
		}),
	}})
	result, err := Analyse(context.Background(), batch, Config{})
	if err != nil {
		t.Fatal(err)
	}
	if result.RecordsAnalysed != 2 || result.DuplicateRecords != 1 || result.ExcludedRecords != 0 {
		t.Errorf("missing bodies were not safely canonicalised: %#v", result)
	}
}

func TestAnalyseVerifiesCanonicalIdentityAfterForcedHashCollision(t *testing.T) {
	batch := fixtureBatch(t, fixture.Config{Seed: 1, HighCardinalityCount: 1, LargeBodyBytes: 64})
	result, err := Analyse(context.Background(), batch, Config{Hash: func([]byte) string { return "one-bucket" }})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := result.DuplicateGroups, uint64(3); got != want {
		t.Errorf("collision-safe duplicate groups = %d, want %d", got, want)
	}
	if got, want := result.DuplicateRecords, uint64(4); got != want {
		t.Errorf("collision-safe duplicate records = %d, want %d", got, want)
	}
}

func TestAnalyseAppliesGroupCapWithinAForcedHashCollisionBucket(t *testing.T) {
	batch := fixtureBatch(t, fixture.Config{Seed: 1, HighCardinalityCount: 1, LargeBodyBytes: 64})
	result, err := Analyse(context.Background(), batch, Config{
		MaxGroups: 1, MaxStateBytes: 1 << 20, MaxIdentityBytes: 1 << 20,
		Hash: func([]byte) string { return "one-bucket" },
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.StateCapReached || result.UniqueIdentities != 1 || result.ExcludedStateCap == 0 {
		t.Errorf("group cap did not apply within a collision bucket: %#v", result)
	}
}

func TestAnalyseReportsStateAndIdentityExclusions(t *testing.T) {
	batch := fixtureBatch(t, fixture.Config{Seed: 1, HighCardinalityCount: 10, LargeBodyBytes: 512})
	stateCapped, err := Analyse(context.Background(), batch, Config{MaxGroups: 1, MaxStateBytes: 1 << 20, MaxIdentityBytes: 1 << 20})
	if err != nil {
		t.Fatal(err)
	}
	if !stateCapped.StateCapReached || stateCapped.ExcludedStateCap == 0 || stateCapped.ExcludedRecords == 0 {
		t.Errorf("state cap not reported: %#v", stateCapped)
	}
	if stateCapped.DuplicateRecords > 2 {
		t.Errorf("state-excluded records were treated as reducible: %#v", stateCapped)
	}

	oversized, err := Analyse(context.Background(), batch, Config{MaxGroups: 100, MaxStateBytes: 1 << 20, MaxIdentityBytes: 8})
	if err != nil {
		t.Fatal(err)
	}
	if oversized.ExcludedOversized != oversized.RecordsPresented || oversized.RecordsAnalysed != 0 || oversized.DuplicateRecords != 0 {
		t.Errorf("oversized identities were not conservatively excluded: %#v", oversized)
	}
}

func TestAnalyseIsDeterministicCancellableAndDoesNotMutateInput(t *testing.T) {
	scenario := fixture.Generate(fixture.Config{Seed: 2, HighCardinalityCount: 2, LargeBodyBytes: 64})
	payload, err := protojson.Marshal(scenario.Request)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := input.Replay(context.Background(), input.ReaderSource{Reader: bytes.NewReader(payload), Info: metadata()}, input.Config{})
	if err != nil {
		t.Fatal(err)
	}
	before, err := protojson.Marshal(replay.Records[0].Record)
	if err != nil {
		t.Fatal(err)
	}
	batch := redaction.Apply(replay)
	first, err := Analyse(context.Background(), batch, Config{})
	if err != nil {
		t.Fatal(err)
	}
	second, err := Analyse(context.Background(), batch, Config{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Error("identical input produced a different C1 result")
	}
	after, err := protojson.Marshal(replay.Records[0].Record)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Error("C1 or redaction mutated the raw replay input")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Analyse(ctx, batch, Config{}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancelled analysis error = %v, want ErrCancelled", err)
	}
}

func TestCanonicalNeverContainsFixtureSecretAfterRedaction(t *testing.T) {
	batch := fixtureBatch(t, fixture.Config{Seed: 3, HighCardinalityCount: 1, LargeBodyBytes: 64})
	for index := 0; index < batch.Len(); index++ {
		identity, err := Canonical(batch.Record(index))
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(identity, []byte("pf_fake_secret_")) || bytes.Contains(identity, []byte("Bearer ")) {
			t.Fatalf("canonical identity %d exposed a fake secret", index)
		}
	}
}

func BenchmarkAnalyseRepeatedSmallLogs(b *testing.B) {
	benchmarkAnalyse(b, fixture.Config{Seed: 1, HighCardinalityCount: 1, LargeBodyBytes: 64})
}
func BenchmarkAnalyseUniqueSmallLogs(b *testing.B) {
	benchmarkAnalyse(b, fixture.Config{Seed: 1, HighCardinalityCount: 2_000, LargeBodyBytes: 64})
}
func BenchmarkAnalyseHighCardinality(b *testing.B) {
	benchmarkAnalyse(b, fixture.Config{Seed: 1, HighCardinalityCount: 5_000, LargeBodyBytes: 64})
}
func BenchmarkAnalyseLargeBodies(b *testing.B) {
	benchmarkAnalyse(b, fixture.Config{Seed: 1, HighCardinalityCount: 1, LargeBodyBytes: 64 << 10})
}

func benchmarkAnalyse(b *testing.B, config fixture.Config) {
	batch := fixtureBatchForBenchmark(b, config)
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if _, err := Analyse(context.Background(), batch, Config{}); err != nil {
			b.Fatal(err)
		}
	}
}

func fixtureBatch(t testing.TB, config fixture.Config) redaction.Batch {
	t.Helper()
	return fixtureBatchForBenchmark(t, config)
}

func fixtureBatchForBenchmark(t testing.TB, config fixture.Config) redaction.Batch {
	t.Helper()
	scenario := fixture.Generate(config)
	payload, err := protojson.Marshal(scenario.Request)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := input.Replay(context.Background(), input.ReaderSource{Reader: bytes.NewReader(payload), Info: metadata()}, input.Config{MaxInputBytes: 128 << 20})
	if err != nil {
		t.Fatal(err)
	}
	return redaction.Apply(replay)
}

func metadata() input.SourceMetadata {
	return input.SourceMetadata{Identity: "c1-test", Kind: input.SourceOfflineReplay, Format: input.FormatOTLPJSON}
}

func batchForRequest(t *testing.T, request *collectlogsv1.ExportLogsServiceRequest) redaction.Batch {
	t.Helper()
	payload, err := protojson.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := input.Replay(context.Background(), input.ReaderSource{Reader: bytes.NewReader(payload), Info: metadata()}, input.Config{})
	if err != nil {
		t.Fatal(err)
	}
	return redaction.Apply(replay)
}

func resourceLogs(service string, records []*logsv1.LogRecord) *logsv1.ResourceLogs {
	return &logsv1.ResourceLogs{Resource: &resourcev1.Resource{Attributes: attrs("service.name", stringValue(service))}, ScopeLogs: []*logsv1.ScopeLogs{{Scope: &commonv1.InstrumentationScope{Name: "test"}, LogRecords: records}}}
}

func record(timestamp uint64, severity logsv1.SeverityNumber, body *commonv1.AnyValue, attributes []*commonv1.KeyValue) *logsv1.LogRecord {
	return &logsv1.LogRecord{TimeUnixNano: timestamp, SeverityNumber: severity, Body: body, Attributes: attributes}
}

func attrs(values ...interface{}) []*commonv1.KeyValue {
	output := make([]*commonv1.KeyValue, 0, len(values)/2)
	for index := 0; index < len(values); index += 2 {
		output = append(output, &commonv1.KeyValue{Key: values[index].(string), Value: values[index+1].(*commonv1.AnyValue)})
	}
	return output
}

func stringValue(value string) *commonv1.AnyValue {
	return &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: value}}
}
func intValue(value int64) *commonv1.AnyValue {
	return &commonv1.AnyValue{Value: &commonv1.AnyValue_IntValue{IntValue: value}}
}
func mapValue(value []*commonv1.KeyValue) *commonv1.AnyValue {
	return &commonv1.AnyValue{Value: &commonv1.AnyValue_KvlistValue{KvlistValue: &commonv1.KeyValueList{Values: value}}}
}

var _ = strings.Builder{}
