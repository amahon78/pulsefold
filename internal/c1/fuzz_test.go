package c1

import (
	"bytes"
	"context"
	"testing"

	"github.com/amahon78/pulsefold/internal/input"
	"github.com/amahon78/pulsefold/internal/redaction"
	collectlogsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

func FuzzCanonicalValueTypes(f *testing.F) {
	f.Add("body", "attribute")
	f.Add("", "")
	f.Fuzz(func(t *testing.T, body, attribute string) {
		request := &collectlogsv1.ExportLogsServiceRequest{ResourceLogs: []*logsv1.ResourceLogs{{ScopeLogs: []*logsv1.ScopeLogs{{LogRecords: []*logsv1.LogRecord{{
			Body:       &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: body}},
			Attributes: []*commonv1.KeyValue{{Key: "attribute", Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: attribute}}}},
		}}}}}}}
		payload, err := protojson.Marshal(request)
		if err != nil {
			// Fuzz-generated strings can be invalid UTF-8, which OTLP JSON
			// correctly refuses before they reach canonicalisation.
			return
		}
		replay, err := input.Replay(context.Background(), input.ReaderSource{Reader: bytes.NewReader(payload), Info: input.SourceMetadata{Identity: "fuzz", Kind: input.SourceOfflineReplay, Format: input.FormatOTLPJSON}}, input.Config{})
		if err != nil {
			t.Fatal(err)
		}
		batch := redaction.Apply(replay)
		if batch.Len() != 1 {
			t.Fatalf("redacted records = %d, want 1", batch.Len())
		}
		if _, err := Canonical(batch.Record(0)); err != nil {
			t.Fatal(err)
		}
	})
}
