package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/amahon78/pulsefold/internal/fixture"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestRunWritesCounterfactualPayloadFreeReport(t *testing.T) {
	scenario := fixture.Generate(fixture.Config{Seed: 1, HighCardinalityCount: 1, LargeBodyBytes: 64})
	payload, err := protojson.Marshal(scenario.Request)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "fixture.otlp.json")
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if exitCode := run([]string{"analyse", path}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	if strings.Contains(stdout.String(), "pf_fake_secret_") || strings.Contains(stdout.String(), "Bearer ") {
		t.Error("CLI report leaked fake telemetry secret")
	}
	if !strings.Contains(stdout.String(), "counterfactual reduction opportunity") {
		t.Error("CLI report lacks counterfactual notice")
	}
}

func TestRunMalformedInputReturnsSafeJSONFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "malformed.otlp.json")
	if err := os.WriteFile(path, []byte(`{"resourceLogs":Bearer pf_fake_secret_cli`), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if exitCode := run([]string{"analyse", path}, &stdout, &stderr); exitCode != 1 {
		t.Fatalf("exit code = %d, want 1", exitCode)
	}
	if strings.Contains(stdout.String(), "pf_fake_secret_cli") || strings.Contains(stderr.String(), "pf_fake_secret_cli") {
		t.Error("CLI malformed-input path leaked fake telemetry secret")
	}
	if !strings.Contains(stdout.String(), `"code": "malformed_input"`) {
		t.Error("CLI did not emit a machine-readable malformed-input report")
	}
}
