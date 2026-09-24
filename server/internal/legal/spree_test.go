package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// spree_test.go — ADR 0065's 2026-09-23 amendment, CR 702.172a: the
// enumerator half. A Spree spell is several casts, one per legal
// mode selection, and — new since this amendment — each selection
// can have its OWN price. The #544 invariant this file pins: a
// selection the seat cannot afford is never offered, and every
// selection that IS offered is one the engine accepts.

const oracleExplosiveDerailment = "b23dc81d-01bb-4bf0-9932-5d32a6b22cf7"

func explosiveDerailmentModes(moves []legal.Move, source uuid.UUID) map[string]bool {
	out := make(map[string]bool)
	for _, m := range moves {
		if m.Kind != legal.KindCast || m.Source != source {
			continue
		}
		var p struct {
			Modes []int `json:"modes"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			continue
		}
		key := ""
		for _, mode := range p.Modes {
			if key != "" {
				key += ","
			}
			key += string(rune('0' + mode))
		}
		out[key] = true
	}
	return out
}

// TestSpreeOffersOnlyAffordableModeCombinations is the "never offer
// what the seat cannot pay for" half. Explosive Derailment's two
// bullets are each {2} beside a printed {R}: three Mountains cover
// the printed cost plus exactly ONE bullet ({R}+{2} = 3), never both
// ({R}+{2}+{2} = 5).
func TestSpreeOffersOnlyAffordableModeCombinations(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[1]
	clearHand(active)
	spell := handCard(active, game.Card{
		Name: "Explosive Derailment", TypeLine: "Instant",
		OracleID: oracleExplosiveDerailment, ManaCost: "{R}",
	})
	battlefieldCard(g, opp, creature("Their Bear", "{1}{G}", 2, 2))
	battlefieldCard(g, opp, game.Card{Name: "Their Sphere", TypeLine: "Artifact"})
	lands(g, active, "Mountain", "Mountain", 3)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	combos := explosiveDerailmentModes(castMovesFor(moves, spell), spell)
	if !combos["0"] {
		t.Errorf("the {2} damage bullet alone should be affordable and offered; combos: %v", combos)
	}
	if !combos["1"] {
		t.Errorf("the {2} artifact bullet alone should be affordable and offered; combos: %v", combos)
	}
	if combos["0,1"] {
		t.Errorf("both bullets together cost {R}+{2}+{2}=5, more than the 3 Mountains fund; combos: %v", combos)
	}

	// #544: every offer this seat sees, the engine accepts.
	dispatchAll(t, g, active.ID, castMovesFor(moves, spell))
}

// TestSpreeOffersTheFullCombinationWhenAffordable is the flip side:
// funded to 5 Mountains, the two-bullet cast is offered too, priced
// at the printed cost plus BOTH bullets — not just one of them.
func TestSpreeOffersTheFullCombinationWhenAffordable(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[1]
	clearHand(active)
	spell := handCard(active, game.Card{
		Name: "Explosive Derailment", TypeLine: "Instant",
		OracleID: oracleExplosiveDerailment, ManaCost: "{R}",
	})
	battlefieldCard(g, opp, creature("Their Bear", "{1}{G}", 2, 2))
	battlefieldCard(g, opp, game.Card{Name: "Their Sphere", TypeLine: "Artifact"})
	lands(g, active, "Mountain", "Mountain", 5)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	casts := castMovesFor(moves, spell)
	combos := explosiveDerailmentModes(casts, spell)
	if !combos["0,1"] {
		t.Fatalf("5 Mountains cover {R}+{2}+{2}=5, so both bullets together should be offered; combos: %v", combos)
	}

	// #544: including the double-bullet cast itself.
	dispatchAll(t, g, active.ID, casts)
}
