package model

import (
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// policystats_test.go is #505 part 2's model-side half: the funnel
// projects Stats into aiseat.PolicyStats, and the projection must
// never drift from the counters it is projecting — the same
// discipline spend_test.go holds Spend() to.

// The funnel is a PolicyStatser, and PolicyStats agrees with Stats.
func TestPolicyStatsProjectsStats(t *testing.T) {
	b := &stubB{index: 0, reason: "stub"}
	fake := AlwaysIndex(1)
	p := testPolicy(t, fake, b, func(c *Config) {
		// One routine call, no escalation noise.
		c.ThreatPower, c.ThreatCreature, c.DangerLife = 999, 999, 0
		c.Epsilon = 0
	})

	if ps := p.PolicyStats(); ps.Windows != 0 {
		t.Fatalf("a seat that has decided nothing reports windows: %+v", ps)
	}

	// A routine window (Layer C, cheap model) …
	decide(t, p, castWindow(), 2*time.Second)
	// … and an escalating one (combat).
	in := castWindow()
	in.Moves = append(in.Moves, mv(legal.KindAttack, "Attack with Bear", ""))
	decide(t, p, in, 2*time.Second)

	st := p.Stats()
	ps := p.PolicyStats()

	if ps.Windows != st.Windows {
		t.Errorf("Windows = %d, want %d", ps.Windows, st.Windows)
	}
	if ps.ModelCalls != st.ModelCalls || ps.ModelCalls == 0 {
		t.Errorf("ModelCalls = %d, want %d (and nonzero)", ps.ModelCalls, st.ModelCalls)
	}
	if ps.Escalated != st.Escalated {
		t.Errorf("Escalated = %d, want %d", ps.Escalated, st.Escalated)
	}
	if ps.ByLayer[LayerC] != st.ByLayer[LayerC] {
		t.Errorf("ByLayer[C] = %d, want %d", ps.ByLayer[LayerC], st.ByLayer[LayerC])
	}
	if ps.Usage.InputTokens != st.Usage.InputTokens {
		t.Errorf("Usage.InputTokens = %d, want %d — the projection disagrees with Stats",
			ps.Usage.InputTokens, st.Usage.InputTokens)
	}
	if ps.ModelLatency != st.ModelLatency {
		t.Errorf("ModelLatency = %v, want %v", ps.ModelLatency, st.ModelLatency)
	}

	// And it is reachable the same way every other optional Policy
	// extension is: aiseat.Capability, not a type assertion the
	// caller has to know about.
	got, ok := aiseat.Capability[aiseat.PolicyStatser](p)
	if !ok {
		t.Fatal("the funnel was not found as a PolicyStatser through Capability")
	}
	if w := got.PolicyStats().Windows; w != st.Windows {
		t.Errorf("Capability-found Windows = %d, want %d", w, st.Windows)
	}
}

// A policy with no model attached is still a PolicyStatser — the
// zero value is the true answer for a deployment with no transport,
// exactly as it is for Spend.
func TestPolicyStatsOnAClientlessSeatIsZero(t *testing.T) {
	p := testPolicy(t, nil, &stubB{index: 0, reason: "stub"}, nil)
	decide(t, p, castWindow(), 2*time.Second)
	ps := p.PolicyStats()
	if ps.ModelCalls != 0 {
		t.Errorf("a clientless seat reported model calls: %+v", ps)
	}
	if ps.ByFallback[FallbackNoClient] != 1 {
		t.Errorf("ByFallback = %v, want one %q", ps.ByFallback, FallbackNoClient)
	}
}
