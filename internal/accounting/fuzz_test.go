package accounting

import (
	"bytes"
	"testing"
)

func FuzzLoadModelJSON(f *testing.F) {
	f.Add([]byte(`{"schema_version":"pulsefold/billing-model/v1","currency":"TST","metric":"accepted_records","indexed_fraction":{"numerator":1,"denominator":1},"pricing":{"fixed":{"unit_block":1,"minor_units_per_block":1}}}`))
	f.Add([]byte("not-json"))
	f.Fuzz(func(t *testing.T, payload []byte) {
		if len(payload) > 4096 {
			t.Skip()
		}
		_, _ = LoadModelJSON(bytes.NewReader(payload))
	})
}
