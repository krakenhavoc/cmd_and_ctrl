package decisionlog_test

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/decisionlog"
)

// plan_test.go pins ADR 0136 §7's additive trace field: a window's turn
// plan is written into the record and read back as it was, and a
// window with no plan writes no `plan` key, so records from before the
// field and after it read the same to an older reader.
func TestTracePlanRoundTrips(t *testing.T) {
	planned := event(1, "B")
	planned.Trace.Plan = []aiseat.PlanMember{
		{Index: 1, Label: "Cast Lightning Bolt"},
		{Index: 2, Label: "Cast Harrow", Held: true},
	}
	plain := event(2, "B")

	_, g, dir := newLog(t, decisionlog.ModeAll, 0)
	g.Observe(planned)
	g.Observe(plain)
	if err := g.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	var got []decisionlog.Record
	if err := decisionlog.Scan(decisionlog.Path(dir, gameID), func(r decisionlog.Record) error {
		got = append(got, r)
		return nil
	}); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("%d records, want 2", len(got))
	}
	if !reflect.DeepEqual(got[0].Trace.Plan, planned.Trace.Plan) {
		t.Errorf("plan read back as %+v, want %+v", got[0].Trace.Plan, planned.Trace.Plan)
	}
	if got[1].Trace.Plan != nil {
		t.Errorf("a window with no plan read back a plan: %+v", got[1].Trace.Plan)
	}

	raw, err := os.ReadFile(decisionlog.Path(dir, gameID))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var traces []map[string]json.RawMessage
	for _, l := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		var rec struct {
			Trace map[string]json.RawMessage `json:"trace"`
		}
		if err := json.Unmarshal([]byte(l), &rec); err != nil || rec.Trace == nil {
			continue
		}
		traces = append(traces, rec.Trace)
	}
	if len(traces) != 2 {
		t.Fatalf("found %d trace objects in the file, want 2", len(traces))
	}
	if _, ok := traces[0]["plan"]; !ok {
		t.Error("the planned window's trace has no `plan` key")
	}
	if _, ok := traces[1]["plan"]; ok {
		t.Error("a window with no plan wrote a `plan` key")
	}
}
