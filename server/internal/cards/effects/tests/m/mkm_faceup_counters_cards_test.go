package testsm

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestMkmFaceUpCountersCardsAreRegisteredFullWithTheirClause(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.MkmBubbleSmugglerOracle: "Bubble Smuggler",
		effectstest.MkmCrowdControlOracle:   "Crowd-Control Warden",
	} {
		def, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Completeness != effects.CompletenessFull {
			t.Errorf("%s completeness = %v, want Full", name, def.Completeness)
		}
		alt := game.FaceDownCastFor(oracle)
		if alt == nil || alt.FaceDown == nil || alt.FaceDown.Kind != game.FaceDownDisguised {
			t.Errorf("%s has no disguise declaration", name)
			continue
		}
		if alt.FaceDown.AsTurnedFaceUp == nil {
			t.Errorf("%s declares no as-turned-face-up counters", name)
		}
		if alt.FaceDown.FaceUpCounter {
			t.Errorf("%s carries megamorph's counter; disguise has none", name)
		}
	}
}

// --- Bubble Smuggler -----------------------------------------------------------

func TestMkmBubbleSmugglerGetsFourCountersAsItTurnsFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.MkmB1CastFaceDown(t, g, "Bubble Smuggler", "Creature — Octopus Fish", effectstest.MkmBubbleSmugglerOracle, 2, 1)
	if n := effectstest.TmtB1Card(t, g, id).Counters["+1/+1"]; n != 0 {
		t.Fatalf("a face-down Smuggler already has %d +1/+1 counters", n)
	}
	effectstest.MkmB1TurnFaceUp(t, g, me, id, "U", "C", "C", "C", "C", "C")
	if n := effectstest.TmtB1Card(t, g, id).Counters["+1/+1"]; n != 4 {
		t.Errorf("+1/+1 counters = %d, want 4", n)
	}
	if p, tough := effectstest.MkmB1PT(t, g, id); p != 6 || tough != 5 {
		t.Errorf("turned up = %d/%d, want 6/5", p, tough)
	}
}

func TestMkmBubbleSmugglerCastFaceUpGetsNoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastHoldCreature(t, g, "Bubble Smuggler", "Creature — Octopus Fish", effectstest.MkmBubbleSmugglerOracle, 2, 1)
	if n := effectstest.TmtB1Card(t, g, id).Counters["+1/+1"]; n != 0 {
		t.Errorf("a Smuggler cast face up got %d +1/+1 counters", n)
	}
	if p, tough := effectstest.MkmB1PT(t, g, id); p != 2 || tough != 1 {
		t.Errorf("cast face up = %d/%d, want the printed 2/1", p, tough)
	}
}

// A Smuggler restored face down still gets its four as it turns up, and
// the four survive a second restore.
func TestMkmBubbleSmugglerCountersSurviveARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.MkmB1CastFaceDown(t, g, "Bubble Smuggler", "Creature — Octopus Fish", effectstest.MkmBubbleSmugglerOracle, 2, 1)

	restored := effectstest.RestoreThroughJSON(t, g)
	me := restored.Seats[restored.Turn.ActiveSeat]
	if !effectstest.TmtB1Card(t, restored, id).FaceDownIsPermanent() {
		t.Fatal("the restored Smuggler is not face down")
	}
	effectstest.MkmB1TurnFaceUp(t, restored, me, id, "U", "C", "C", "C", "C", "C")
	if n := effectstest.TmtB1Card(t, restored, id).Counters["+1/+1"]; n != 4 {
		t.Fatalf("restored Smuggler turned up with %d +1/+1 counters, want 4", n)
	}

	again := effectstest.RestoreThroughJSON(t, restored)
	if n := effectstest.TmtB1Card(t, again, id).Counters["+1/+1"]; n != 4 {
		t.Errorf("after a second restore the Smuggler has %d +1/+1 counters, want 4", n)
	}
	if p, tough := effectstest.MkmB1PT(t, again, id); p != 6 || tough != 5 {
		t.Errorf("restored = %d/%d, want 6/5", p, tough)
	}
}

// --- Crowd-Control Warden ------------------------------------------------------

func TestMkmCrowdControlWardenEntersWithACounterPerOtherCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := effectstest.CastHoldCreature(t, g, "Crowd-Control Warden", "Creature — Centaur Soldier", effectstest.MkmCrowdControlOracle, 4, 4)
	if n := effectstest.TmtB1Card(t, g, id).Counters["+1/+1"]; n != 2 {
		t.Errorf("+1/+1 counters on entering = %d, want 2 (the other creatures you control)", n)
	}
	if p, tough := effectstest.MkmB1PT(t, g, id); p != 6 || tough != 6 {
		t.Errorf("entered = %d/%d, want 6/6", p, tough)
	}
}

// Face down it enters with nothing (CR 708.2a). Turned up later it
// counts the board as it is then, once (CR 708.8: it is not entering).
func TestMkmCrowdControlWardenCountsTheBoardAsItTurnsFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	id := effectstest.MkmB1CastFaceDown(t, g, "Crowd-Control Warden", "Creature — Centaur Soldier", effectstest.MkmCrowdControlOracle, 4, 4)
	if n := effectstest.TmtB1Card(t, g, id).Counters["+1/+1"]; n != 0 {
		t.Fatalf("a face-down Warden entered with %d +1/+1 counters; it has no abilities face down", n)
	}
	effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)

	effectstest.MkmB1TurnFaceUp(t, g, me, id, "G", "W", "C", "C", "C")

	if n := effectstest.TmtB1Card(t, g, id).Counters["+1/+1"]; n != 3 {
		t.Errorf("+1/+1 counters after turning face up = %d, want 3 (the other creatures you control now)", n)
	}
	if p, tough := effectstest.MkmB1PT(t, g, id); p != 7 || tough != 7 {
		t.Errorf("turned up = %d/%d, want 7/7", p, tough)
	}
}
