package testsm

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Every face-down cast declared in the catalog has to be reachable by
// the key the client will send, priced at the keyword's {3}, and
// carry the CARD's own face-up cost — which is the only place in the
// catalog the engine can read it from once the permanent is face down
// and has no text at all (ADR 0082 decision 4). A typo in any of the
// three is a morph that cannot be cast, cannot be turned up, or is
// turned up for the wrong price, and nothing else would catch it.
func TestMorphDeclarationsAreWired(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, key, faceUp string
		kind                      game.FaceDownKind
		counter                   bool
	}{
		{"Willbender", effectstest.WillbenderOracle, "morph", "{1}{U}", game.FaceDownMorphed, false},
		{"Ainok Tracker", effectstest.AinokTrackerOracle, "morph", "{4}{R}", game.FaceDownMorphed, false},
		{"Aerie Bowmasters", effectstest.AerieBowmastersOracle, "megamorph", "{5}{G}", game.FaceDownMorphed, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ac := game.AlternativeCostByKey(tc.oracle, tc.key)
			if ac == nil {
				t.Fatalf("%s offers no %q cost", tc.name, tc.key)
			}
			if ac.ManaCost != effects.MorphCastCost {
				t.Errorf("%s casts face down for %q, want %q — the price is the keyword's (CR 702.37b)",
					tc.name, ac.ManaCost, effects.MorphCastCost)
			}
			if ac.Label == "" {
				t.Errorf("%s has no label for the cast picker", tc.name)
			}
			if ac.FaceDown == nil {
				t.Fatalf("%s %q is not a face-down cast", tc.name, tc.key)
			}
			if ac.FaceDown.Kind != tc.kind {
				t.Errorf("%s face-down kind = %q, want %q", tc.name, ac.FaceDown.Kind, tc.kind)
			}
			if ac.FaceDown.FaceUpCost != tc.faceUp {
				t.Errorf("%s face-up cost = %q, want the printed one %q", tc.name, ac.FaceDown.FaceUpCost, tc.faceUp)
			}
			if ac.FaceDown.FaceUpCounter != tc.counter {
				t.Errorf("%s megamorph counter = %v, want %v (CR 702.37b)", tc.name, ac.FaceDown.FaceUpCounter, tc.counter)
			}
		})
	}
}

// The whole round trip through the real catalog: Ainok Tracker casts
// face down for {3} as a nameless 2/2 with no first strike, resolves
// into a permanent nobody but its controller can read, and comes back
// as the 3/3 that strikes first for {4}{R} — the same object it was
// the whole time (CR 708.8).
func TestAinokTrackerMorphsAndTurnsBackUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	id := effectstest.CastWithAltCost(t, g, "Ainok Tracker", "Creature — Dog Scout", effectstest.AinokTrackerOracle, "morph")

	down := effectstest.FindCardAnywhere(t, g, id)
	if !down.FaceDown || down.FaceDownKind != game.FaceDownMorphed {
		t.Fatalf("the spell is not face down: (%v, %q)", down.FaceDown, down.FaceDownKind)
	}
	if game.HasKeyword(&down, "first strike") {
		t.Error("a face-down spell kept its printed keyword (CR 708.2a)")
	}
	if down.IsKnownTo(opp.ID) {
		t.Error("an opponent may read a face-down spell on the stack (CR 708.5)")
	}

	effectstest.PassPriorityAroundTable(t, g)
	perm := effectstest.FindCardAnywhere(t, g, id)
	if !perm.FaceDownIsPermanent() {
		t.Fatal("the permanent did not enter face down")
	}

	g.WithWriteLock(func() { active.ManaPool.AddMana(effectstest.ManaTokens("R", "C", "C", "C", "C")...) })
	if err := g.PerformSpecialAction(active.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up: %v", err)
	}
	up := effectstest.FindCardAnywhere(t, g, id)
	if up.FaceDown {
		t.Fatal("still face down after the special action")
	}
	if up.Name != "Ainok Tracker" {
		t.Errorf("name = %q, want the card back", up.Name)
	}
	if !game.HasKeyword(&up, "first strike") {
		t.Error("first strike did not come back with the card")
	}
	if up.InstanceID != id {
		t.Error("turning face up made a new object (CR 708.8 says it does not)")
	}
}

// Megamorph's one extra clause, through the real card.
func TestAerieBowmastersTurnsUpWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]

	id := effectstest.CastWithAltCost(t, g, "Aerie Bowmasters", "Creature — Dog Archer", effectstest.AerieBowmastersOracle, "megamorph")
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { active.ManaPool.AddMana(effectstest.ManaTokens("G", "C", "C", "C", "C", "C")...) })
	if err := g.PerformSpecialAction(active.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up: %v", err)
	}
	up := effectstest.FindCardAnywhere(t, g, id)
	if n := up.Counters["+1/+1"]; n != 1 {
		t.Errorf("+1/+1 counters = %d, want 1 (CR 702.37b)", n)
	}
	if !game.HasKeyword(&up, "reach") {
		t.Error("reach did not come back with the card")
	}
}
