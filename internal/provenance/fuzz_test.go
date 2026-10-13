package provenance

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/amahon78/pulsefold/internal/accounting"
)

func FuzzParseManifestJSON(f *testing.F) {
	manifest := Build(testReplay(), "pulsefold/v0alpha1", "test", "v1", Completion{State: "complete"}, nil, nil, accounting.Result{})
	AttachIntegrity(&manifest, Commitment([]byte("safe")))
	manifest.CreatedAt = testCreatedAt
	payload, err := json.Marshal(manifest)
	if err != nil {
		f.Fatal(err)
	}
	f.Add(payload)
	f.Add([]byte("not-json"))
	f.Fuzz(func(t *testing.T, payload []byte) {
		if len(payload) > 4096 {
			t.Skip()
		}
		manifest, err := ParseManifestJSON(bytes.NewReader(payload))
		if err != nil {
			if strings.Contains(err.Error(), "{") || len(err.Error()) > 64 {
				t.Fatalf("parse error echoed input: %v", err)
			}
			return
		}
		if manifest.Validate() != nil || manifest.Integrity.Authenticated {
			t.Fatal("parser returned an invalid manifest")
		}
		encoded, err := json.Marshal(manifest)
		if err != nil {
			t.Fatal(err)
		}
		again, err := ParseManifestJSON(bytes.NewReader(encoded))
		if err != nil || !reflect.DeepEqual(again, manifest) {
			t.Fatalf("accepted manifest did not round-trip: %v", err)
		}
	})
}
