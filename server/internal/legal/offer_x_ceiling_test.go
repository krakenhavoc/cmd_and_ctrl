package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Silver Scrutiny
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// offer_x_ceiling_test.go — the enumerator half of #3225: a per-offer X
// ceiling ("as though it had flash if X is 3 or less", CR 601.3c) caps
// the X announced under THAT claim and no other. At instant speed only
// the flash offer is open, so no cast move announces X > 3; in a main
// phase the printed cost still announces what the mana pays for.

const oracleSilverScrutiny = "2b0b4fc1-9ad6-4259-a0a2-9a14484c246e"

func altKeyOf(t *testing.T, m legal.Move) string {
	t.Helper()
	var p struct {
		AlternativeCost string `json:"alternative_cost"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params %s: %v", string(m.Params), err)
	}
	return p.AlternativeCost
}

func silverScrutinyTable(t *testing.T, step game.Step) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	card := handCard(active, game.Card{Name: "Silver Scrutiny", TypeLine: "Sorcery", ManaCost: "{X}{U}{U}", OracleID: oracleSilverScrutiny})
	// Nine Islands pay for X = 7.
	lands(g, active, "Island", "Island", 9)
	advanceTo(t, g, step)
	return g, active, card
}

func TestFlashOfferIsNeverEnumeratedAboveItsCeilingAtInstantSpeed(t *testing.T) {
	g, active, card := silverScrutinyTable(t, game.StepBeginCombat)

	moves := legal.EnumerateFor(g, active.ID)
	casts := castMovesFor(moves, card)
	if len(casts) == 0 {
		t.Fatal("no cast offered at instant speed under the flash offer")
	}
	for _, m := range casts {
		if k := altKeyOf(t, m); k != "flash-x-at-most" {
			t.Errorf("move %q claims %q at instant speed, want only the flash offer", m.Label, k)
		}
		if x := xValueOf(t, m); x > 3 {
			t.Errorf("move %q announces X=%d over the offer's ceiling 3", m.Label, x)
		}
		if m.Value != nil && m.Value.Max != nil && *m.Value.Max > 3 {
			t.Errorf("move %q opens X up to %d over the offer's ceiling 3", m.Label, *m.Value.Max)
		}
	}
	if x := xValueOf(t, casts[0]); x != 3 {
		t.Errorf("offered X=%d, want the ceiling 3 (the largest legal X)", x)
	}
	dispatchAll(t, g, active.ID, moves)
}

func TestPrintedCostIsNotBoundByTheOfferInAMainPhase(t *testing.T) {
	g, active, card := silverScrutinyTable(t, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	var printedMax, offerMax int
	for _, m := range castMovesFor(moves, card) {
		x := xValueOf(t, m)
		if m.Value != nil && m.Value.Max != nil && *m.Value.Max > x {
			x = *m.Value.Max
		}
		if altKeyOf(t, m) == "flash-x-at-most" {
			offerMax = max(offerMax, x)
		} else {
			printedMax = max(printedMax, x)
		}
	}
	if printedMax != 7 {
		t.Errorf("the printed cost announces up to X=%d, want 7 (what nine lands pay for)", printedMax)
	}
	if offerMax != 3 {
		t.Errorf("the flash offer announces up to X=%d, want 3", offerMax)
	}
	dispatchAll(t, g, active.ID, moves)
}
