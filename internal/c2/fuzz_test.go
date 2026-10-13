package c2

import (
	"bytes"
	"context"
	"testing"

	"github.com/amahon78/pulsefold/internal/c1"
	"github.com/amahon78/pulsefold/internal/input"
	"github.com/amahon78/pulsefold/internal/redaction"
	collectlogsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

func FuzzAnalyseRedactedStrings(f *testing.F) {
	f.Add("request.id=request-a", "second body")
	f.Add("", "")
	f.Fuzz(func(t *testing.T, first, second string) {
		request := &collectlogsv1.ExportLogsServiceRequest{ResourceLogs: []*logsv1.ResourceLogs{{ScopeLogs: []*logsv1.ScopeLogs{{LogRecords: []*logsv1.LogRecord{
			{Body: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: first}}},
			{Body: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: second}}},
		}}}}}}
		payload, err := protojson.Marshal(request)
		if err != nil {
			return
		}
		replay, err := input.Replay(context.Background(), input.ReaderSource{Reader: bytes.NewReader(payload), Info: input.SourceMetadata{Identity: "c2-fuzz", Kind: input.SourceOfflineReplay, Format: input.FormatOTLPJSON}}, input.Config{MaxInputBytes: 64 << 10})
		if err != nil {
			t.Fatal(err)
		}
		batch := redaction.Apply(replay)
		c1Result, err := c1.Analyse(context.Background(), batch, c1.Config{})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := Analyse(context.Background(), batch, c1Result, Config{}); err != nil {
			t.Fatal(err)
		}
	})
}
