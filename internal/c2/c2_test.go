package c2

import (
	"bytes"
	"context"
	"errors"
	"math/rand"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/amahon78/pulsefold/internal/c1"
	"github.com/amahon78/pulsefold/internal/fixture"
	"github.com/amahon78/pulsefold/internal/input"
	"github.com/amahon78/pulsefold/internal/redaction"
	collectlogsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcev1 "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestAnalyseFindsSafeStructuralCandidatesAndPreservesBoundaries(t *testing.T) {
	scenario := fixture.GenerateTemplateScenario(fixture.Config{Seed: 1})
	batch := templateBatch(t, scenario.Request)
	c1Result, err := c1.Analyse(context.Background(), batch, c1.Config{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := Analyse(context.Background(), batch, c1Result, Config{})
	if err != nil {
		t.Fatal(err)
	}
	if result.SourceTemplateRecords != 2 || result.SourceTemplateGroups != 1 {
		t.Errorf("source template accounting = %#v", result)
	}
	if result.ExcludedUnsupportedBody != 1 || result.ExcludedRecords == 0 {
		t.Errorf("structured body was not explicitly excluded: %#v", result)
	}
	groups := assignmentGroups(result, scenario.Truth)
	mustSameGroup(t, groups, "login-1", "login-2")
	mustSameGroup(t, groups, "uuid-ip-time-1", "uuid-ip-time-2")
	mustSameGroup(t, groups, "stack-101", "stack-102")
	mustSameGroup(t, groups, "json-1", "json-2")
	mustSameGroup(t, groups, "secret-1", "secret-2")
	mustDifferentGroup(t, groups, "error-500", "error-501")
	mustDifferentGroup(t, groups, "login-1", "orders-service-boundary")
	mustDifferentGroup(t, groups, "orders-service-boundary", "orders-severity-boundary")
	if result.AdditionalOpportunity == 0 || result.InferredTemplateGroups >= result.CandidateRecords {
		t.Errorf("C2 did not report structural opportunity: %#v", result)
	}

	metrics := Evaluate(result, templateTruth(scenario.Truth))
	if metrics.UnsafeMergeGroups != 0 || metrics.UnsafeMergeRecordRate != 0 {
		t.Errorf("unsafe merge detected: %#v", metrics)
	}
	if metrics.ErrorPreservationRate != 1 || metrics.NovelPreservationRate != 1 {
		t.Errorf("novel or error records were not isolated: %#v", metrics)
	}
}

func TestAnalyseHasVisibleBoundsAndCancellation(t *testing.T) {
	scenario := fixture.Generate(fixture.Config{Seed: 4, HighCardinalityCount: 20, LargeBodyBytes: 512})
	batch := templateBatch(t, scenario.Request)
	c1Result, err := c1.Analyse(context.Background(), batch, c1.Config{})
	if err != nil {
		t.Fatal(err)
	}
	capped, err := Analyse(context.Background(), batch, c1Result, Config{MaxClusters: 1, MaxPartitions: 1, MaxStateBytes: 1024, MaxBodyBytes: 32, MaxTokens: 8, MaxSourceTemplates: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !capped.StateCapReached || capped.ExcludedStateCap == 0 || capped.ExcludedOversizedBody == 0 {
		t.Errorf("bounds were not visible and conservative: %#v", capped)
	}
	if capped.AdditionalOpportunity > capped.CandidateRecords {
		t.Errorf("capped analysis made an unsafe opportunity claim: %#v", capped)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Analyse(ctx, batch, c1Result, Config{}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancelled analysis error = %v, want ErrCancelled", err)
	}
}

func TestOversizedSourceTemplateIsVisibleWithoutBlockingInference(t *testing.T) {
	scenario := fixture.GenerateTemplateScenario(fixture.Config{Seed: 3})
	request := proto.Clone(scenario.Request).(*collectlogsv1.ExportLogsServiceRequest)
	request.ResourceLogs[0].ScopeLogs[0].LogRecords[0].Attributes[0].Value = &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: strings.Repeat("x", 128)}}
	batch := templateBatch(t, request)
	c1Result, err := c1.Analyse(context.Background(), batch, c1.Config{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := Analyse(context.Background(), batch, c1Result, Config{MaxTemplateBytes: 8})
	if err != nil {
		t.Fatal(err)
	}
	if result.SourceTemplateExcluded == 0 || result.RecordsAnalysed == 0 {
		t.Errorf("oversized source template was not safely visible: %#v", result)
	}
}

func TestIndependentLearnersAgreeAcrossDeterministicOrderings(t *testing.T) {
	scenario := fixture.GenerateTemplateScenario(fixture.Config{Seed: 6})
	firstBatch := templateBatch(t, scenario.Request)
	shuffled := proto.Clone(scenario.Request).(*collectlogsv1.ExportLogsServiceRequest)
	random := rand.New(rand.NewSource(73))
	for _, resourceLogs := range shuffled.ResourceLogs {
		for _, scopeLogs := range resourceLogs.ScopeLogs {
			random.Shuffle(len(scopeLogs.LogRecords), func(first, second int) {
				scopeLogs.LogRecords[first], scopeLogs.LogRecords[second] = scopeLogs.LogRecords[second], scopeLogs.LogRecords[first]
			})
		}
	}
	secondBatch := templateBatch(t, shuffled)
	firstC1, err := c1.Analyse(context.Background(), firstBatch, c1.Config{})
	if err != nil {
		t.Fatal(err)
	}
	secondC1, err := c1.Analyse(context.Background(), secondBatch, c1.Config{})
	if err != nil {
		t.Fatal(err)
	}
	first, err := Analyse(context.Background(), firstBatch, firstC1, Config{})
	if err != nil {
		t.Fatal(err)
	}
	second, err := Analyse(context.Background(), secondBatch, secondC1, Config{})
	if err != nil {
		t.Fatal(err)
	}
	if first.InferredTemplateGroups != second.InferredTemplateGroups || first.AdditionalOpportunity != second.AdditionalOpportunity || first.ExcludedRecords != second.ExcludedRecords {
		t.Errorf("deterministic learner diverged across orderings:\nfirst=%#v\nsecond=%#v", first, second)
	}
}

func TestIndependentPartitionedLearnersDoNotIncreaseClaimedOpportunity(t *testing.T) {
	scenario := fixture.GenerateTemplateScenario(fixture.Config{Seed: 6})
	fullBatch := templateBatch(t, scenario.Request)
	fullC1, err := c1.Analyse(context.Background(), fullBatch, c1.Config{})
	if err != nil {
		t.Fatal(err)
	}
	full, err := Analyse(context.Background(), fullBatch, fullC1, Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, count := range []int{2, 4} {
		var candidateRecords, templateGroups, opportunity uint64
		for _, request := range partitionRequests(scenario.Request, count) {
			batch := templateBatch(t, request)
			c1Result, err := c1.Analyse(context.Background(), batch, c1.Config{})
			if err != nil {
				t.Fatal(err)
			}
			result, err := Analyse(context.Background(), batch, c1Result, Config{})
			if err != nil {
				t.Fatal(err)
			}
			candidateRecords += result.CandidateRecords
			templateGroups += result.InferredTemplateGroups
			opportunity += result.AdditionalOpportunity
		}
		if candidateRecords != full.CandidateRecords || templateGroups < full.InferredTemplateGroups || opportunity > full.AdditionalOpportunity {
			t.Fatalf("%d independent partitions changed the conservative C2 claim: full=%#v candidate=%d groups=%d opportunity=%d", count, full, candidateRecords, templateGroups, opportunity)
		}
	}
}

func TestC2ReceivesOnlyRedactedContent(t *testing.T) {
	scenario := fixture.GenerateTemplateScenario(fixture.Config{Seed: 2})
	batch := templateBatch(t, scenario.Request)
	c1Result, err := c1.Analyse(context.Background(), batch, c1.Config{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := Analyse(context.Background(), batch, c1Result, Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, assignment := range result.Assignments {
		_, _, _, _, log := batch.Record(assignment.InputIndex).Clone()
		encoded, err := protojson.Marshal(log)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(encoded, []byte("pf_fake_secret_")) || bytes.Contains(encoded, []byte("Bearer ")) {
			t.Fatal("C2 was supplied unredacted telemetry")
		}
	}
}

func BenchmarkAnalyseC1AndC2(b *testing.B) {
	benchmarkC2(b, fixture.Config{Seed: 1, HighCardinalityCount: 20, LargeBodyBytes: 64})
}

func BenchmarkAnalyseC1AndC2HighCardinality(b *testing.B) {
	benchmarkC2(b, fixture.Config{Seed: 1, HighCardinalityCount: 5_000, LargeBodyBytes: 64})
}

func BenchmarkAnalyseC1Only(b *testing.B) {
	batch := templateBatchForBenchmark(b, fixture.Generate(fixture.Config{Seed: 1, HighCardinalityCount: 20, LargeBodyBytes: 64}).Request)
	benchmarkLatency(b, uint64(batch.Len()), func() error {
		_, err := c1.Analyse(context.Background(), batch, c1.Config{})
		return err
	})
}

func benchmarkC2(b *testing.B, config fixture.Config) {
	batch := templateBatchForBenchmark(b, fixture.Generate(config).Request)
	c1Result, err := c1.Analyse(context.Background(), batch, c1.Config{})
	if err != nil {
		b.Fatal(err)
	}
	benchmarkLatency(b, uint64(batch.Len()), func() error {
		_, err := Analyse(context.Background(), batch, c1Result, Config{})
		return err
	})
}

func benchmarkLatency(b *testing.B, records uint64, run func() error) {
	b.Helper()
	samples := make([]time.Duration, 0, min(b.N, 100_000))
	var elapsed time.Duration
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		started := time.Now()
		if err := run(); err != nil {
			b.Fatal(err)
		}
		duration := time.Since(started)
		elapsed += duration
		if len(samples) < cap(samples) {
			samples = append(samples, duration)
		}
	}
	b.StopTimer()
	if len(samples) == 0 {
		return
	}
	sort.Slice(samples, func(first, second int) bool { return samples[first] < samples[second] })
	b.ReportMetric(float64(percentile(samples, 50).Nanoseconds()), "p50-ns/op")
	b.ReportMetric(float64(percentile(samples, 95).Nanoseconds()), "p95-ns/op")
	b.ReportMetric(float64(percentile(samples, 99).Nanoseconds()), "p99-ns/op")
	if elapsed > 0 {
		b.ReportMetric(float64(records*uint64(b.N))/elapsed.Seconds(), "sample-throughput-records/s")
	}
}

func percentile(values []time.Duration, value int) time.Duration {
	index := (len(values)*value + 99) / 100
	if index == 0 {
		return values[0]
	}
	return values[index-1]
}

func templateBatch(t testing.TB, request *collectlogsv1.ExportLogsServiceRequest) redaction.Batch {
	t.Helper()
	return templateBatchForBenchmark(t, request)
}

func templateBatchForBenchmark(t testing.TB, request *collectlogsv1.ExportLogsServiceRequest) redaction.Batch {
	t.Helper()
	payload, err := protojson.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := input.Replay(context.Background(), input.ReaderSource{Reader: bytes.NewReader(payload), Info: input.SourceMetadata{Identity: "c2-fixture", Kind: input.SourceOfflineReplay, Format: input.FormatOTLPJSON}}, input.Config{MaxInputBytes: 128 << 20})
	if err != nil {
		t.Fatal(err)
	}
	return redaction.Apply(replay)
}

func assignmentGroups(result Result, truth fixture.Truth) map[string]uint64 {
	byIndex := make(map[int]uint64, len(result.Assignments))
	for _, assignment := range result.Assignments {
		byIndex[assignment.InputIndex] = assignment.Group
	}
	groups := make(map[string]uint64, len(truth.Records))
	for index, record := range truth.Records {
		groups[record.ID] = byIndex[index]
	}
	// C1 supplies one representative for exact duplicates. Propagate that
	// representative's opaque C2 group to its truth-equivalent duplicate.
	for _, record := range truth.Records {
		if groups[record.ID] != 0 {
			continue
		}
		for _, representative := range truth.Records {
			if record.TemplateGroup == representative.TemplateGroup && groups[representative.ID] != 0 {
				groups[record.ID] = groups[representative.ID]
				break
			}
		}
	}
	return groups
}

func templateTruth(truth fixture.Truth) map[int]GroundTruth {
	output := make(map[int]GroundTruth, len(truth.Records))
	for index, record := range truth.Records {
		output[index] = GroundTruth{Group: record.TemplateGroup, Novel: record.Novel, Error: record.Error}
	}
	return output
}

func partitionRequests(request *collectlogsv1.ExportLogsServiceRequest, count int) []*collectlogsv1.ExportLogsServiceRequest {
	partitions := make([]*collectlogsv1.ExportLogsServiceRequest, count)
	for index := range partitions {
		partitions[index] = &collectlogsv1.ExportLogsServiceRequest{}
	}
	resources := make([][]*logsv1.ResourceLogs, count)
	for partition := range partitions {
		resources[partition] = make([]*logsv1.ResourceLogs, len(request.ResourceLogs))
		for resourceIndex, resourceLogs := range request.ResourceLogs {
			copy := &logsv1.ResourceLogs{Resource: proto.Clone(resourceLogs.Resource).(*resourcev1.Resource), SchemaUrl: resourceLogs.SchemaUrl}
			for _, scopeLogs := range resourceLogs.ScopeLogs {
				copy.ScopeLogs = append(copy.ScopeLogs, &logsv1.ScopeLogs{Scope: proto.Clone(scopeLogs.Scope).(*commonv1.InstrumentationScope), SchemaUrl: scopeLogs.SchemaUrl})
			}
			resources[partition][resourceIndex] = copy
			partitions[partition].ResourceLogs = append(partitions[partition].ResourceLogs, copy)
		}
	}
	position := 0
	for resourceIndex, resourceLogs := range request.ResourceLogs {
		for scopeIndex, scopeLogs := range resourceLogs.ScopeLogs {
			for _, record := range scopeLogs.LogRecords {
				target := resources[position%count][resourceIndex].ScopeLogs[scopeIndex]
				position++
				target.LogRecords = append(target.LogRecords, proto.Clone(record).(*logsv1.LogRecord))
			}
		}
	}
	return partitions
}

func mustSameGroup(t *testing.T, groups map[string]uint64, first, second string) {
	t.Helper()
	if groups[first] == 0 || groups[first] != groups[second] {
		t.Errorf("%s and %s were not grouped together: %#v", first, second, groups)
	}
}

func mustDifferentGroup(t *testing.T, groups map[string]uint64, first, second string) {
	t.Helper()
	if groups[first] == 0 || groups[first] == groups[second] {
		t.Errorf("%s and %s were unsafely grouped: %#v", first, second, groups)
	}
}
