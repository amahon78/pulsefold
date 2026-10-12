package fixture

import (
	"reflect"
	"testing"

	"google.golang.org/protobuf/proto"
)

func TestGenerateIsDeterministicAndTruthLabelled(t *testing.T) {
	first := Generate(Config{Seed: 42, HighCardinalityCount: 4, LargeBodyBytes: 128})
	second := Generate(Config{Seed: 42, HighCardinalityCount: 4, LargeBodyBytes: 128})
	if !proto.Equal(first.Request, second.Request) {
		t.Fatal("same seed produced different OTLP fixtures")
	}
	if !reflect.DeepEqual(first.Truth, second.Truth) {
		t.Fatal("same seed produced different fixture truth")
	}
	if got, want := first.Truth.ExpectedGroupRecords["payment-accepted"], uint64(3); got != want {
		t.Errorf("payment duplicate truth = %d, want %d", got, want)
	}
	if got, want := first.Truth.ExpectedGroupRecords["structured-reordered"], uint64(2); got != want {
		t.Errorf("structured duplicate truth = %d, want %d", got, want)
	}
	if got, want := first.Truth.ExpectedGroupRecords["redacted-secret"], uint64(2); got != want {
		t.Errorf("redacted secret truth = %d, want %d", got, want)
	}

	different := Generate(Config{Seed: 43, HighCardinalityCount: 4, LargeBodyBytes: 128})
	if proto.Equal(first.Request, different.Request) {
		t.Error("different seed did not change synthetic high-cardinality input")
	}
}
