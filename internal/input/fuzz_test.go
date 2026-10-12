package input

import (
	"bytes"
	"context"
	"testing"
)

func FuzzReplayOTLPJSON(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"resourceLogs":[{}]}`))
	f.Add([]byte(`{"resourceLogs":`))
	f.Fuzz(func(t *testing.T, payload []byte) {
		if len(payload) > 64<<10 {
			t.Skip()
		}
		_, _ = Replay(context.Background(), ReaderSource{
			Reader: bytes.NewReader(payload),
			Info:   SourceMetadata{Identity: "fuzz", Kind: SourceOfflineReplay, Format: FormatOTLPJSON},
		}, Config{MaxInputBytes: 64 << 10})
	})
}
