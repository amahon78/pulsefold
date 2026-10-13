package c2

import (
	"strings"
	"testing"

	drain3 "github.com/jaeyo/go-drain3/pkg/drain3"
)

// TestUpstreamDrainBaselineShowsWhyC2KeepsErrorCodesLiteral exercises the
// exact go-drain3 version used by the upstream Collector Drain processor. It
// is a benchmark baseline only, not a PulseFold runtime dependency: with the
// upstream default similarity threshold it generalises two distinct error
// codes. C2 must retain that boundary in a report-only safety-first mode.
func TestUpstreamDrainBaselineShowsWhyC2KeepsErrorCodesLiteral(t *testing.T) {
	baseline, err := drain3.NewDrain(drain3.WithSimTh(0.4), drain3.WithMaxCluster(8))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := baseline.AddLogMessage("gateway failed error_code=500 request.id=request-a"); err != nil {
		t.Fatal(err)
	}
	cluster, _, err := baseline.AddLogMessage("gateway failed error_code=501 request.id=request-b")
	if err != nil {
		t.Fatal(err)
	}
	if cluster == nil || !strings.Contains(cluster.GetTemplate(), "<*>") {
		t.Fatalf("upstream Drain did not produce the expected generalisation: %#v", cluster)
	}
}
