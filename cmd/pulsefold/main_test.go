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
	if !strings.Contains(stdout.String(), `"wire_bytes_not_measured"`) || !strings.Contains(stdout.String(), `"billing_model_not_supplied"`) {
		t.Error("CLI report lacks explicit unknown wire and billing states")
	}
}

func TestRunAppliesAnOperatorSuppliedGenericBillingModel(t *testing.T) {
	scenario := fixture.Generate(fixture.Config{Seed: 2, HighCardinalityCount: 0, LargeBodyBytes: 0})
	payload, err := protojson.Marshal(scenario.Request)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	inputPath := filepath.Join(directory, "fixture.otlp.json")
	billingPath := filepath.Join(directory, "billing.json")
	if err := os.WriteFile(inputPath, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	billing := []byte(`{"schema_version":"pulsefold/billing-model/v1","currency":"TST","metric":"accepted_records","indexed_fraction":{"numerator":1,"denominator":1},"pricing":{"fixed":{"unit_block":1,"minor_units_per_block":2}}}`)
	if err := os.WriteFile(billingPath, billing, 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if exitCode := run([]string{"analyse", inputPath, "--billing", billingPath}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"model_supplied": true`) || !strings.Contains(stdout.String(), `"gross_observability_cost_opportunity"`) {
		t.Errorf("billing report missing modelled opportunity: %s", stdout.String())
	}
}

func TestRunRejectsInvalidBillingConfigurationWithoutLeakingItsContents(t *testing.T) {
	directory := t.TempDir()
	inputPath := filepath.Join(directory, "fixture.otlp.json")
	billingPath := filepath.Join(directory, "billing.json")
	if err := os.WriteFile(inputPath, []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}
	secret := "pf_fake_secret_billing_configuration"
	if err := os.WriteFile(billingPath, []byte(`{"unknown":"`+secret+`"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if exitCode := run([]string{"analyse", inputPath, "--billing", billingPath}, &stdout, &stderr); exitCode != 2 {
		t.Fatalf("exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	if strings.Contains(stdout.String(), secret) || strings.Contains(stderr.String(), secret) {
		t.Error("invalid billing configuration leaked its contents")
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

func TestRunNeverEchoesInputPathsInReport(t *testing.T) {
	scenario := fixture.Generate(fixture.Config{Seed: 3, HighCardinalityCount: 0, LargeBodyBytes: 0})
	payload, err := protojson.Marshal(scenario.Request)
	if err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(t.TempDir(), "Users", "alex", "example")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"file.json", "my logs.json", "ghp_0123456789abcdef.json", "file.json?token=abc", `C:\Users\example\file.json`} {
		path := filepath.Join(directory, name)
		if err := os.WriteFile(path, payload, 0o600); err != nil {
			t.Fatal(err)
		}
		var stdout, stderr bytes.Buffer
		if exitCode := run([]string{"analyse", path}, &stdout, &stderr); exitCode != 0 {
			t.Fatalf("%q: exit code = %d", name, exitCode)
		}
		output := stdout.String() + stderr.String()
		for _, forbidden := range []string{directory, "/Users/", "alex", "ghp_", "token=abc", `C:\`, "my logs"} {
			if strings.Contains(output, forbidden) {
				t.Errorf("%q: report leaked %q", name, forbidden)
			}
		}
		wantLabel := `"identity": "file.json"`
		if name != "file.json" {
			wantLabel = `"identity": "` + fallbackSourceLabel + `"`
		}
		if !strings.Contains(stdout.String(), wantLabel) || !strings.Contains(stdout.String(), `"manifest_id": "pfm_v0_`) {
			t.Errorf("%q: unexpected source label or missing manifest", name)
		}
	}
}
