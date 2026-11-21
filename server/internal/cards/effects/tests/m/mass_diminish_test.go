package testsm

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Mass Diminish (until your next turn) -----------------------

func TestMassDiminishLastsUntilYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.CtrlPushCreature(g, opp.ID, "Bear")
	mine := effectstest.CtrlPushCreature(g, me.ID, "My Bear")

	effectstest.CastCatalogSpell(t, g, "Mass Diminish", "Sorcery", effectstest.MassDiminishOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CtrlEffectivePower(t, g, theirs); got != 1 {
		t.Errorf("the target player's creature: power %d, want 1", got)
	}
	if got := effectstest.CtrlEffectivePower(t, g, mine); got != 2 {
		t.Errorf("the caster's own creature: power %d, want 2 (only the target player's are diminished)", got)
	}

	// Through every opponent's turn: a duration keyed on Turn.Round
	// would have ended here, because all four seats share one round.
	for seat := 1; seat <= 3; seat++ {
		effectstest.AdvancePastCleanupForTest(t, g)
		if got := effectstest.CtrlEffectivePower(t, g, theirs); got != 1 {
			t.Fatalf("after seat %d's turn: power %d, want 1 — the effect ended early", seat, got)
		}
	}
	// And now the caster's own turn begins.
	effectstest.AdvancePastCleanupForTest(t, g)
	if got := effectstest.CtrlEffectivePower(t, g, theirs); got != 2 {
		t.Errorf("after the caster's next turn began: power %d, want the printed 2", got)
	}
}

func TestMassDiminishLocksInItsAffectedSet(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := g.Seats[0], g.Seats[1]
	before := effectstest.CtrlPushCreature(g, opp.ID, "Early Bear")

	effectstest.CastCatalogSpell(t, g, "Mass Diminish", "Sorcery", effectstest.MassDiminishOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	after := effectstest.CtrlPushCreature(g, opp.ID, "Late Bear")
	if got := effectstest.CtrlEffectivePower(t, g, before); got != 1 {
		t.Errorf("a creature present at resolution: power %d, want 1", got)
	}
	if got := effectstest.CtrlEffectivePower(t, g, after); got != 2 {
		t.Errorf("a creature that arrived afterwards: power %d, want 2 (CR 611.2c)", got)
	}
}
