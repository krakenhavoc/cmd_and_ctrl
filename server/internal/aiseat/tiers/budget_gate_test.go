package tiers

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
)

// budget_gate_test.go: #2980's picker half. A spend cap or the kill
// switch makes the model tiers unavailable in the picker with the
// cap's own note, but never makes an already-added seat unbuildable.

type fakeGate struct{ note string }

func (g *fakeGate) Blocked() string { return g.note }

func TestBlockedGateGreysTheModelTiersWithItsNote(t *testing.T) {
	gate := &fakeGate{}
	f := NewFactory(FactoryOptions{Client: model.AlwaysIndex(0), Gate: gate})

	for _, tier := range []aiseat.Tier{aiseat.TierAssisted, aiseat.TierStrong} {
		if st := f.TierStatus(tier); !st.Available {
			t.Fatalf("%s with an open gate: %+v, want available", tier, st)
		}
	}

	gate.note = "daily budget used up; resets 00:00 UTC"
	for _, tier := range []aiseat.Tier{aiseat.TierAssisted, aiseat.TierStrong} {
		st := f.TierStatus(tier)
		if st.Available || st.Reason != gate.note {
			t.Fatalf("%s with a blocked gate: %+v, want unavailable carrying the note", tier, st)
		}
	}
	// No tier is ever sold: the free tiers are untouched.
	for _, tier := range []aiseat.Tier{aiseat.TierRandom, aiseat.TierHeuristic} {
		if st := f.TierStatus(tier); !st.Available {
			t.Fatalf("%s: %+v, want available whatever the gate says", tier, st)
		}
	}
}

func TestBlockedGateStillBuildsASeatThatWasAlreadyAdded(t *testing.T) {
	gate := &fakeGate{note: "switched off"}
	f := NewFactory(FactoryOptions{Client: model.AlwaysIndex(0), Gate: gate})
	p, err := f.NewPolicy(aiseat.SeatSpec{Tier: string(aiseat.TierAssisted)})
	if err != nil || p == nil {
		t.Fatalf("NewPolicy with a blocked gate: %v, want a policy: a seat added before the cap must still play, on the heuristic", err)
	}
}

func TestNoClientStillOutranksTheGate(t *testing.T) {
	f := NewFactory(FactoryOptions{Gate: &fakeGate{note: "x"}})
	st := f.TierStatus(aiseat.TierAssisted)
	if st.Available || st.Reason == "x" {
		t.Fatalf("%+v: with no transport the reason is the missing-model one", st)
	}
}
