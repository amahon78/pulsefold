// Package fixture creates deterministic, synthetic OTLP log scenarios.
package fixture

import (
	"fmt"
	"math/rand"

	collectlogsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcev1 "go.opentelemetry.io/proto/otlp/resource/v1"
)

// Config controls the deterministic synthetic scenario. Seed is part of the
// fixture identity and must be recorded by callers that persist a fixture.
type Config struct {
	Seed                 int64
	HighCardinalityCount int
	LargeBodyBytes       int
}

func (c Config) normalised() Config {
	if c.HighCardinalityCount <= 0 {
		c.HighCardinalityCount = 32
	}
	if c.LargeBodyBytes <= 0 {
		c.LargeBodyBytes = 4096
	}
	return c
}

// Truth is independently created by the generator, rather than inferred from
// a PulseFold analyser. ExpectedGroup is the redacted exact-duplicate group.
type Truth struct {
	Seed                 int64
	Records              []RecordTruth
	ExpectedGroupRecords map[string]uint64
}

type RecordTruth struct {
	ID              string
	ExpectedGroup   string
	SecretShaped    bool
	HighCardinality bool
}

type Scenario struct {
	Request *collectlogsv1.ExportLogsServiceRequest
	Truth   Truth
}

// Generate returns a fixed multi-service scenario containing exact duplicates,
// unique records and adversarial near-duplicates. All apparent secrets are
// deliberately fake test strings.
func Generate(config Config) Scenario {
	config = config.normalised()
	random := rand.New(rand.NewSource(config.Seed))
	truth := Truth{Seed: config.Seed, ExpectedGroupRecords: map[string]uint64{}}
	appendTruth := func(id, group string, secret, highCardinality bool) {
		truth.Records = append(truth.Records, RecordTruth{ID: id, ExpectedGroup: group, SecretShaped: secret, HighCardinality: highCardinality})
		truth.ExpectedGroupRecords[group]++
	}

	paymentRecords := make([]*logsv1.LogRecord, 0, 10+config.HighCardinalityCount)
	for index := 0; index < 3; index++ {
		paymentRecords = append(paymentRecords, record(uint64(1+index), logsv1.SeverityNumber_SEVERITY_NUMBER_INFO, stringValue("payment accepted"), attrs("event.kind", stringValue("payment"))))
		appendTruth(fmt.Sprintf("duplicate-%d", index), "payment-accepted", false, false)
	}
	paymentRecords = append(paymentRecords,
		record(10, logsv1.SeverityNumber_SEVERITY_NUMBER_ERROR, stringValue("payment accepted"), attrs("event.kind", stringValue("payment"))),
		record(11, logsv1.SeverityNumber_SEVERITY_NUMBER_INFO, stringValue("payment acceptd"), attrs("event.kind", stringValue("payment"))),
		record(12, logsv1.SeverityNumber_SEVERITY_NUMBER_ERROR, stringValue("gateway error code=500"), nil),
		record(13, logsv1.SeverityNumber_SEVERITY_NUMBER_ERROR, stringValue("gateway error code=501"), nil),
		record(14, logsv1.SeverityNumber_SEVERITY_NUMBER_INFO, mapValue(attrs("amount", intValue(7), "currency", stringValue("AUD"))), nil),
		record(15, logsv1.SeverityNumber_SEVERITY_NUMBER_INFO, mapValue(attrs("currency", stringValue("AUD"), "amount", intValue(7))), nil),
		record(16, logsv1.SeverityNumber_SEVERITY_NUMBER_INFO, stringValue("value=1"), attrs("typed", stringValue("1"))),
		record(17, logsv1.SeverityNumber_SEVERITY_NUMBER_INFO, stringValue("value=1"), attrs("typed", intValue(1))),
		record(18, logsv1.SeverityNumber_SEVERITY_NUMBER_ERROR, stringValue("authorisation failed: Bearer pf_fake_secret_payment_123"), nil),
		record(19, logsv1.SeverityNumber_SEVERITY_NUMBER_ERROR, stringValue("authorisation failed: Bearer pf_fake_secret_payment_456"), nil),
		record(20, logsv1.SeverityNumber_SEVERITY_NUMBER_INFO, stringValue(repeat("x", config.LargeBodyBytes)), nil),
	)
	for _, id := range []string{"same-body-different-severity", "one-character-difference", "error-code-500", "error-code-501", "structured-reordered", "structured-reordered", "string-one", "numeric-one", "secret-one", "secret-two", "large-value"} {
		group := id
		if id == "structured-reordered" {
			group = "structured-reordered"
		}
		if id == "secret-one" || id == "secret-two" {
			group = "redacted-secret"
		}
		appendTruth(id, group, id == "secret-one" || id == "secret-two", false)
	}
	for index := 0; index < config.HighCardinalityCount; index++ {
		identifier := fmt.Sprintf("request-%08x", random.Uint64())
		paymentRecords = append(paymentRecords, record(uint64(100+index), logsv1.SeverityNumber_SEVERITY_NUMBER_INFO, stringValue("high-cardinality event"), attrs("request.id", stringValue(identifier))))
		appendTruth(fmt.Sprintf("high-cardinality-%d", index), fmt.Sprintf("high-cardinality-%d", index), false, true)
	}

	ordersRecords := []*logsv1.LogRecord{
		record(30, logsv1.SeverityNumber_SEVERITY_NUMBER_INFO, stringValue("payment accepted"), attrs("event.kind", stringValue("payment"))),
		record(31, logsv1.SeverityNumber_SEVERITY_NUMBER_INFO, arrayValue(stringValue("queued"), boolValue(true)), nil),
	}
	appendTruth("same-body-different-service", "same-body-different-service", false, false)
	appendTruth("array-body", "array-body", false, false)

	return Scenario{
		Request: &collectlogsv1.ExportLogsServiceRequest{ResourceLogs: []*logsv1.ResourceLogs{
			resourceLogs("payments", paymentRecords),
			resourceLogs("orders", ordersRecords),
		}},
		Truth: truth,
	}
}

func resourceLogs(service string, records []*logsv1.LogRecord) *logsv1.ResourceLogs {
	return &logsv1.ResourceLogs{
		Resource:  &resourcev1.Resource{Attributes: attrs("service.name", stringValue(service), "service.version", stringValue("v1"))},
		ScopeLogs: []*logsv1.ScopeLogs{{Scope: &commonv1.InstrumentationScope{Name: "pulsefold.fixture"}, LogRecords: records}},
	}
}

func record(timestamp uint64, severity logsv1.SeverityNumber, body *commonv1.AnyValue, attributes []*commonv1.KeyValue) *logsv1.LogRecord {
	return &logsv1.LogRecord{TimeUnixNano: timestamp, SeverityNumber: severity, Body: body, Attributes: attributes}
}

func attrs(values ...interface{}) []*commonv1.KeyValue {
	attributes := make([]*commonv1.KeyValue, 0, len(values)/2)
	for index := 0; index < len(values); index += 2 {
		attributes = append(attributes, &commonv1.KeyValue{Key: values[index].(string), Value: values[index+1].(*commonv1.AnyValue)})
	}
	return attributes
}

func stringValue(value string) *commonv1.AnyValue {
	return &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: value}}
}

func intValue(value int64) *commonv1.AnyValue {
	return &commonv1.AnyValue{Value: &commonv1.AnyValue_IntValue{IntValue: value}}
}

func boolValue(value bool) *commonv1.AnyValue {
	return &commonv1.AnyValue{Value: &commonv1.AnyValue_BoolValue{BoolValue: value}}
}

func mapValue(values []*commonv1.KeyValue) *commonv1.AnyValue {
	return &commonv1.AnyValue{Value: &commonv1.AnyValue_KvlistValue{KvlistValue: &commonv1.KeyValueList{Values: values}}}
}

func arrayValue(values ...*commonv1.AnyValue) *commonv1.AnyValue {
	return &commonv1.AnyValue{Value: &commonv1.AnyValue_ArrayValue{ArrayValue: &commonv1.ArrayValue{Values: values}}}
}

func repeat(value string, count int) string {
	output := make([]byte, count)
	for index := range output {
		output[index] = value[0]
	}
	return string(output)
}
