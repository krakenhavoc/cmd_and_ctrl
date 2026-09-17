package model

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/rules"
)

// trace_test.go pins DecideTraced against every failure the funnel
// can produce, and pins the one property that makes it safe to ship:
// it is the SAME decision as Decide, with the working attached. A
// DecideTraced that quietly played differently would make the
// decision log a record of a game nobody played.

// Compile-time assertions, so a refactor that drops a method is a
// build failure rather than a runner that silently stops tracing.
var (
	_ aiseat.Tracer = (*Policy)(nil)
	_ aiseat.Tracer = (*rules.Filter)(nil)
	_ aiseat.Tracer = (*heuristic.Policy)(nil)
)

func tracedPolicy(t *testing.T, client Client) (*Policy, *stubB) {
	t.Helper()
	b := &stubB{index: 1, reason: "layer b", rank: []heuristic.Candidate{
		{Index: 1, Value: 9, Reason: "best"},
		{Index: 0, Value: 1, Reason: "pass"},
	}}
	cfg := DefaultConfig()
	cfg.Fallback = b
	cfg.Client = client
	cfg.Routine.ID, cfg.Frontier.ID = "routine-model", "frontier-model"
	// Small budgets: these tests are about classification, not about
	// how long a call may take, and the timeout case has to actually
	// expire inside a unit test.
	cfg.Reserve, cfg.MinBudget, cfg.MaxCall = 10*time.Millisecond, 5*time.Millisecond, 150*time.Millisecond
	return New(cfg), b
}

func TestDecideTracedRecordsTheModelAnswer(t *testing.T) {
	p, _ := tracedPolicy(t, AlwaysIndex(2))
	in := castWindow()
	d, tr, err := p.DecideTraced(context.Background(), in)
	if err != nil {
		t.Fatalf("DecideTraced: %v", err)
	}
	if d.Index != 2 {
		t.Fatalf("decision index %d, want the model's 2", d.Index)
	}
	if tr.Layer != LayerC {
		t.Errorf("layer %q, want %q", tr.Layer, LayerC)
	}
	if tr.Fallback != "" {
		t.Errorf("fallback %q on a window the model answered", tr.Fallback)
	}
	if tr.ParsedIndex == nil || *tr.ParsedIndex != 2 {
		t.Errorf("parsed index %v, want 2", tr.ParsedIndex)
	}
	if tr.HeuristicIndex != 1 {
		t.Errorf("heuristic index %d, want Layer B's 1", tr.HeuristicIndex)
	}
	if len(tr.Candidates) != 2 || tr.Candidates[0].Index != 1 {
		t.Errorf("ranking not carried: %+v", tr.Candidates)
	}
	if tr.Model != "routine-model" {
		t.Errorf("model %q", tr.Model)
	}
	if !strings.Contains(tr.Reply, `"index"`) {
		t.Errorf("reply not recorded: %q", tr.Reply)
	}
	// The prompt is the whole point: without it a decision log is a
	// list of numbers nobody can review.
	if tr.Prompt == nil || tr.Prompt.User == "" || len(tr.Prompt.System) == 0 {
		t.Fatalf("prompt not recorded: %+v", tr.Prompt)
	}
	if !strings.Contains(tr.Prompt.System[0], "You are playing one seat") {
		t.Errorf("system block is not the primer: %q", truncate(tr.Prompt.System[0], 60))
	}
}

func TestDecideTracedClassifiesEveryModelFailure(t *testing.T) {
	cases := []struct {
		name     string
		client   Client
		fallback string
		wantCall bool
	}{
		{"malformed", AlwaysText("nope"), FallbackMalformed, true},
		{"out of range", AlwaysIndex(99), FallbackOutOfRange, true},
		{"timeout", NeverAnswers(), FallbackError, true},
		{"no client", nil, FallbackNoClient, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, b := tracedPolicy(t, tc.client)
			d, tr, err := p.DecideTraced(context.Background(), castWindow())
			if err != nil {
				t.Fatalf("DecideTraced: %v", err)
			}
			if d.Index != b.index {
				t.Errorf("index %d; every model failure plays Layer B's %d", d.Index, b.index)
			}
			if tr.Layer != LayerB {
				t.Errorf("layer %q, want %q", tr.Layer, LayerB)
			}
			if tr.Fallback != tc.fallback {
				t.Errorf("fallback %q, want %q", tr.Fallback, tc.fallback)
			}
			if tc.name == "timeout" && !tr.TimedOut {
				t.Error("a call that ran out of budget is not marked TimedOut")
			}
			if got := tr.Prompt != nil; got != tc.wantCall {
				t.Errorf("prompt recorded = %v, want %v", got, tc.wantCall)
			}
			if tr.ParsedIndex != nil && tc.name != "out of range" {
				t.Errorf("parsed index %d on a reply that was not an index", *tr.ParsedIndex)
			}
		})
	}
}

func TestDecideTracedRecordsLayerA(t *testing.T) {
	p, b := tracedPolicy(t, AlwaysIndex(0))
	// One move: Layer A's "only legal move" rule settles it and
	// nothing below is asked.
	in := castWindow()
	in.Moves = in.Moves[:1]
	d, tr, err := p.DecideTraced(context.Background(), in)
	if err != nil {
		t.Fatalf("DecideTraced: %v", err)
	}
	if d.Index != 0 || tr.Layer != LayerA {
		t.Fatalf("decision %+v layer %q, want index 0 at Layer A", d, tr.Layer)
	}
	if tr.Rule == "" {
		t.Error("Layer A answered without naming a rule")
	}
	if tr.HeuristicIndex != aiseat.Decline {
		t.Errorf("heuristic index %d on a window Layer A absorbed; nobody asked it", tr.HeuristicIndex)
	}
	if b.calls != 0 {
		t.Errorf("Layer B was consulted %d times on a window Layer A absorbed", b.calls)
	}
}

// The property that makes the whole thing trustworthy.
func TestDecideAndDecideTracedAgree(t *testing.T) {
	clients := map[string]Client{
		"answers":   AlwaysIndex(2),
		"malformed": AlwaysText("nope"),
		"oob":       AlwaysIndex(99),
		"none":      nil,
	}
	for name, c := range clients {
		t.Run(name, func(t *testing.T) {
			p1, _ := tracedPolicy(t, c)
			p2, _ := tracedPolicy(t, c)
			in := castWindow()
			plain, err1 := p1.Decide(context.Background(), in)
			traced, tr, err2 := p2.DecideTraced(context.Background(), in)
			if (err1 == nil) != (err2 == nil) {
				t.Fatalf("errors differ: %v vs %v", err1, err2)
			}
			if plain != traced {
				t.Errorf("Decide gave %+v, DecideTraced gave %+v — they must be the same decision", plain, traced)
			}
			if tr.Layer == "" {
				t.Error("a trace with no layer says nothing")
			}
		})
	}
}

// BuildRequest must produce the bytes a live call would, without
// making one.
func TestBuildRequestAssemblesTheCallWithoutMakingIt(t *testing.T) {
	fake := AlwaysIndex(0)
	p, _ := tracedPolicy(t, fake)
	in := castWindow()

	req, cands, v := p.BuildRequest(context.Background(), in)
	if fake.Calls() != 0 {
		t.Fatalf("BuildRequest made %d model calls", fake.Calls())
	}
	if v.Absorbed() {
		t.Fatal("the fixture window is supposed to survive Layer A")
	}
	if req.Model != "routine-model" || req.User == "" || len(req.System) == 0 {
		t.Fatalf("request not assembled: %+v", req)
	}
	if len(cands) == 0 {
		t.Error("no ranking returned")
	}

	// The same window through the real path must send the same bytes.
	_, tr, err := p.DecideTraced(context.Background(), in)
	if err != nil {
		t.Fatalf("DecideTraced: %v", err)
	}
	if tr.Prompt == nil || tr.Prompt.User != req.User {
		t.Errorf("BuildRequest's user delta is not what the call sent")
	}

	// A window Layer A settles has no request to build.
	one := castWindow()
	one.Moves = one.Moves[:1]
	req2, _, v2 := p.BuildRequest(context.Background(), one)
	if !v2.Absorbed() {
		t.Fatal("a one-move window is supposed to be absorbed")
	}
	if req2.User != "" || len(req2.System) != 0 {
		t.Errorf("an absorbed window assembled a request anyway: %+v", req2)
	}
}

// BuildRequest must not move the absorption meter: it is not a window
// anybody played.
func TestBuildRequestDoesNotMeterTheWindow(t *testing.T) {
	m := &rules.Meter{}
	cfg := DefaultConfig()
	cfg.Fallback = &stubB{index: 1}
	cfg.Meter = m
	p := New(cfg)
	one := castWindow()
	one.Moves = one.Moves[:1]
	p.BuildRequest(context.Background(), one)
	if got := m.Stats().Windows; got != 0 {
		t.Errorf("BuildRequest metered %d windows", got)
	}
}
