package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mkm_faceup_counters_cards_test.go — Murders at Karlov Manor's two
// "As this creature is turned face up, put N +1/+1 counters on it"
// cards (#4469, #5237): Bubble Smuggler's fixed four and Crowd-Control
// Warden's X. Each test plays the real catalog entry.

const (
	mkmBubbleSmugglerOracle = "ca22fa33-5460-4d31-a9c8-4115d35b994c"
	mkmCrowdControlOracle   = "439e59ab-0350-4729-9352-60e697eaa78d"
)

func TestMkmFaceUpCountersCardsAreRegisteredFullWithTheirClause(t *testing.T) {
	for oracle, name := range map[string]string{
		mkmBubbleSmugglerOracle: "Bubble Smuggler",
		mkmCrowdControlOracle:   "Crowd-Control Warden",
	} {
		def, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Completeness != CompletenessFull {
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := mkmB1CastFaceDown(t, g, "Bubble Smuggler", "Creature — Octopus Fish", mkmBubbleSmugglerOracle, 2, 1)
	if n := tmtB1Card(t, g, id).Counters["+1/+1"]; n != 0 {
		t.Fatalf("a face-down Smuggler already has %d +1/+1 counters", n)
	}
	mkmB1TurnFaceUp(t, g, me, id, "U", "C", "C", "C", "C", "C")
	if n := tmtB1Card(t, g, id).Counters["+1/+1"]; n != 4 {
		t.Errorf("+1/+1 counters = %d, want 4", n)
	}
	if p, tough := mkmB1PT(t, g, id); p != 6 || tough != 5 {
		t.Errorf("turned up = %d/%d, want 6/5", p, tough)
	}
}

func TestMkmBubbleSmugglerCastFaceUpGetsNoCounters(t *testing.T) {
	g := newCatalogGame(t)
	id := castHoldCreature(t, g, "Bubble Smuggler", "Creature — Octopus Fish", mkmBubbleSmugglerOracle, 2, 1)
	if n := tmtB1Card(t, g, id).Counters["+1/+1"]; n != 0 {
		t.Errorf("a Smuggler cast face up got %d +1/+1 counters", n)
	}
	if p, tough := mkmB1PT(t, g, id); p != 2 || tough != 1 {
		t.Errorf("cast face up = %d/%d, want the printed 2/1", p, tough)
	}
}

// A Smuggler restored face down still gets its four as it turns up, and
// the four survive a second restore.
func TestMkmBubbleSmugglerCountersSurviveARestore(t *testing.T) {
	g := newCatalogGame(t)
	id := mkmB1CastFaceDown(t, g, "Bubble Smuggler", "Creature — Octopus Fish", mkmBubbleSmugglerOracle, 2, 1)

	restored := restoreThroughJSON(t, g)
	me := restored.Seats[restored.Turn.ActiveSeat]
	if !tmtB1Card(t, restored, id).FaceDownIsPermanent() {
		t.Fatal("the restored Smuggler is not face down")
	}
	mkmB1TurnFaceUp(t, restored, me, id, "U", "C", "C", "C", "C", "C")
	if n := tmtB1Card(t, restored, id).Counters["+1/+1"]; n != 4 {
		t.Fatalf("restored Smuggler turned up with %d +1/+1 counters, want 4", n)
	}

	again := restoreThroughJSON(t, restored)
	if n := tmtB1Card(t, again, id).Counters["+1/+1"]; n != 4 {
		t.Errorf("after a second restore the Smuggler has %d +1/+1 counters, want 4", n)
	}
	if p, tough := mkmB1PT(t, again, id); p != 6 || tough != 5 {
		t.Errorf("restored = %d/%d, want 6/5", p, tough)
	}
}

// --- Crowd-Control Warden ------------------------------------------------------

func TestMkmCrowdControlWardenEntersWithACounterPerOtherCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := castHoldCreature(t, g, "Crowd-Control Warden", "Creature — Centaur Soldier", mkmCrowdControlOracle, 4, 4)
	if n := tmtB1Card(t, g, id).Counters["+1/+1"]; n != 2 {
		t.Errorf("+1/+1 counters on entering = %d, want 2 (the other creatures you control)", n)
	}
	if p, tough := mkmB1PT(t, g, id); p != 6 || tough != 6 {
		t.Errorf("entered = %d/%d, want 6/6", p, tough)
	}
}

// Face down it enters with nothing (CR 708.2a). Turned up later it
// counts the board as it is then, once (CR 708.8: it is not entering).
func TestMkmCrowdControlWardenCountsTheBoardAsItTurnsFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	id := mkmB1CastFaceDown(t, g, "Crowd-Control Warden", "Creature — Centaur Soldier", mkmCrowdControlOracle, 4, 4)
	if n := tmtB1Card(t, g, id).Counters["+1/+1"]; n != 0 {
		t.Fatalf("a face-down Warden entered with %d +1/+1 counters; it has no abilities face down", n)
	}
	b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)

	mkmB1TurnFaceUp(t, g, me, id, "G", "W", "C", "C", "C")

	if n := tmtB1Card(t, g, id).Counters["+1/+1"]; n != 3 {
		t.Errorf("+1/+1 counters after turning face up = %d, want 3 (the other creatures you control now)", n)
	}
	if p, tough := mkmB1PT(t, g, id); p != 7 || tough != 7 {
		t.Errorf("turned up = %d/%d, want 7/7", p, tough)
	}
}
