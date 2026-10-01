package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// target_sameness_test.go — #1807 (ADR 0106 §5), the enumerator's half
// of "from a single graveyard": it builds every set inside one
// graveyard, the engine accepts every move it offers, and the view
// ships the owner keys the picker greys by.

const oracleDecompose = "12c2b6bb-05f8-4291-b9a5-c4ddb7a69b9e"

// Decompose off two Swamps, with cards in two graveyards: every offered
// set comes from one graveyard, a set from each graveyard is offered,
// and the engine accepts them all.
func TestDecomposeEnumeratesOnlySingleGraveyardSets(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	card := handCard(active, game.Card{
		Name: "Decompose", TypeLine: "Sorcery", ManaCost: "{1}{B}", OracleID: oracleDecompose,
	})
	for i := 0; i < 2; i++ {
		battlefieldCard(g, active, basic("Swamp", "Swamp"))
	}
	owner := map[uuid.UUID]uuid.UUID{
		graveyardCreature(active, "Mine A", "{B}"):   active.ID,
		graveyardCreature(active, "Mine B", "{B}"):   active.ID,
		graveyardCreature(opp, "Theirs A", "{B}"):    opp.ID,
		graveyardCreature(opp, "Theirs B", "{1}{B}"): opp.ID,
	}
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	casts := castMovesFor(moves, card)
	if len(casts) == 0 {
		t.Fatalf("Decompose is castable off two Swamps: %v", labels(moves))
	}
	sawPair := map[uuid.UUID]bool{}
	for _, m := range casts {
		_, ids := castTargetIDs(t, m)
		graveyards := map[uuid.UUID]bool{}
		for _, id := range ids {
			graveyards[owner[id]] = true
		}
		if len(graveyards) > 1 {
			t.Errorf("%q reaches %d graveyards", m.Label, len(graveyards))
		}
		if len(ids) == 2 {
			for gy := range graveyards {
				sawPair[gy] = true
			}
		}
	}
	if !sawPair[active.ID] || !sawPair[opp.ID] {
		t.Errorf("a pair from each graveyard is a legal set and should be offered, saw %v", sawPair)
	}
	dispatchAll(t, g, active.ID, moves)
}

// The view ships the sameness rule with the owner of each legal card as
// its key, the keys the enumerator groups by.
func TestSamenessRuleViewShipsTheOwnerKeys(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	card := handCard(active, game.Card{
		Name: "Decompose", TypeLine: "Sorcery", ManaCost: "{1}{B}", OracleID: oracleDecompose,
	})
	mine := graveyardCreature(active, "Mine", "{B}")
	theirs := graveyardCreature(opp, "Theirs", "{B}")
	advanceTo(t, g, game.StepPrecombatMain)

	v := protocol.ViewOfGameFor(g, active.ID.String())
	var lt *protocol.LegalTargetsView
	for _, s := range v.Seats {
		for _, c := range s.Hand.Cards {
			if c.InstanceID == card.String() {
				lt = c.LegalTargets
			}
		}
	}
	if lt == nil || lt.Same == nil {
		t.Fatalf("the hand card ships its sameness rule: %+v", lt)
	}
	if lt.Same.Label != "come from a single graveyard" {
		t.Errorf("label = %q", lt.Same.Label)
	}
	if lt.Different != nil {
		t.Error("a sameness clause ships no difference rule")
	}
	if lt.Same.Keys[mine.String()] != active.ID.String() || lt.Same.Keys[theirs.String()] != opp.ID.String() {
		t.Errorf("keys = %v, want each card's owner", lt.Same.Keys)
	}
}
