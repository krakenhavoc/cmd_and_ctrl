package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mass_diminish_test.go pins the catalog half of the CR 611.2
// duration model (#756's siblings live in control_cards_test.go):
// "until your next turn" really does last a full round in a
// four-player game, and the affected set is locked in at resolution.

const massDiminishOracle = "3a3fd911-b71c-49eb-aadb-cfa80b9e3a3a"

// ctrlPushCreature seeds a creature on the battlefield with the
// zone-move event fired, so the layer listener stamps
// EnteredBattlefieldAt (the key every affected set is pinned on).
func ctrlPushCreature(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

// ctrlEffectivePower reads a card's post-layer power.
func ctrlEffectivePower(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return effectivePower(t, g, id)
}

// --- Mass Diminish (until your next turn) -----------------------

func TestMassDiminishLastsUntilYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := ctrlPushCreature(g, opp.ID, "Bear")
	mine := ctrlPushCreature(g, me.ID, "My Bear")

	castCatalogSpell(t, g, "Mass Diminish", "Sorcery", massDiminishOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	if got := ctrlEffectivePower(t, g, theirs); got != 1 {
		t.Errorf("the target player's creature: power %d, want 1", got)
	}
	if got := ctrlEffectivePower(t, g, mine); got != 2 {
		t.Errorf("the caster's own creature: power %d, want 2 (only the target player's are diminished)", got)
	}

	// Through every opponent's turn: a duration keyed on Turn.Round
	// would have ended here, because all four seats share one round.
	for seat := 1; seat <= 3; seat++ {
		advancePastCleanupForTest(t, g)
		if got := ctrlEffectivePower(t, g, theirs); got != 1 {
			t.Fatalf("after seat %d's turn: power %d, want 1 — the effect ended early", seat, got)
		}
	}
	// And now the caster's own turn begins.
	advancePastCleanupForTest(t, g)
	if got := ctrlEffectivePower(t, g, theirs); got != 2 {
		t.Errorf("after the caster's next turn began: power %d, want the printed 2", got)
	}
}

func TestMassDiminishLocksInItsAffectedSet(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := g.Seats[0], g.Seats[1]
	before := ctrlPushCreature(g, opp.ID, "Early Bear")

	castCatalogSpell(t, g, "Mass Diminish", "Sorcery", massDiminishOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	after := ctrlPushCreature(g, opp.ID, "Late Bear")
	if got := ctrlEffectivePower(t, g, before); got != 1 {
		t.Errorf("a creature present at resolution: power %d, want 1", got)
	}
	if got := ctrlEffectivePower(t, g, after); got != 2 {
		t.Errorf("a creature that arrived afterwards: power %d, want 2 (CR 611.2c)", got)
	}
}

// advancePastCleanupForTest walks the cursor until the active seat
// changes — the marker that this seat's cleanup step ran (Turn.Round
// counts rounds, not seat-turns, so the seat is the reliable signal).
func advancePastCleanupForTest(t *testing.T, g *game.Game) {
	t.Helper()
	start := g.Turn.ActiveSeat
	for i := 0; i < 40; i++ {
		if g.Turn.ActiveSeat != start {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatalf("cursor never left seat %d's turn", start)
}
