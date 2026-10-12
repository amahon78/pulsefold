package redaction

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/amahon78/pulsefold/internal/fixture"
	"github.com/amahon78/pulsefold/internal/input"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcev1 "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestApplyRedactsFixtureSecretsBeforeExposure(t *testing.T) {
	scenario := fixture.Generate(fixture.Config{Seed: 1, HighCardinalityCount: 2, LargeBodyBytes: 64})
	payload, err := protojson.Marshal(scenario.Request)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(payload, []byte("pf_fake_secret_")) {
		t.Fatal("fixture unexpectedly contains no fake secret")
	}

	replay, err := input.Replay(context.Background(), input.ReaderSource{
		Reader: bytes.NewReader(payload),
		Info:   input.SourceMetadata{Identity: "fixture", Kind: input.SourceOfflineReplay, Format: input.FormatOTLPJSON},
	}, input.Config{})
	if err != nil {
		t.Fatal(err)
	}
	batch := Apply(replay)
	if batch.PolicyVersion != PolicyVersion {
		t.Errorf("policy version = %q, want %q", batch.PolicyVersion, PolicyVersion)
	}
	for index := 0; index < batch.Len(); index++ {
		resource, resourceSchema, scope, scopeSchema, log := batch.Record(index).Clone()
		encoded := marshalRecord(t, resource, resourceSchema, scope, scopeSchema, log)
		if bytes.Contains(encoded, []byte("pf_fake_secret_")) || bytes.Contains(encoded, []byte("Bearer ")) {
			t.Fatalf("redacted record %d exposed a fake secret", index)
		}
	}
	if !bytes.Contains(payload, []byte("pf_fake_secret_")) {
		t.Error("redaction mutated the original replay payload")
	}
}

func TestApplyRedactsNestedAndSecretNamedAttributes(t *testing.T) {
	raw := input.RawLog{Record: &logsv1.LogRecord{
		Body: &commonv1.AnyValue{Value: &commonv1.AnyValue_KvlistValue{KvlistValue: &commonv1.KeyValueList{Values: []*commonv1.KeyValue{{
			Key: "nested", Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "token=pf_fake_secret_nested"}},
		}}}}},
		Attributes: []*commonv1.KeyValue{{Key: "authorization", Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "Bearer pf_fake_secret_attribute"}}}},
	}}
	batch := Apply(input.ReplayResult{Records: []input.RawLog{raw}})
	_, _, _, _, log := batch.Record(0).Clone()
	encoded, err := protojson.Marshal(log)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "pf_fake_secret_") || strings.Contains(string(encoded), "Bearer ") {
		t.Error("redacted structured log leaked a fake secret")
	}
}

func marshalRecord(t *testing.T, resource *resourcev1.Resource, resourceSchema string, scope *commonv1.InstrumentationScope, scopeSchema string, log *logsv1.LogRecord) []byte {
	t.Helper()
	encoded, err := protojson.Marshal(&logsv1.ResourceLogs{
		Resource: resource, SchemaUrl: resourceSchema,
		ScopeLogs: []*logsv1.ScopeLogs{{Scope: scope, SchemaUrl: scopeSchema, LogRecords: []*logsv1.LogRecord{log}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}
