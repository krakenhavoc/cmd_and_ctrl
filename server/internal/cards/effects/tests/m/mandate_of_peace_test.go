package testsm

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mandate_of_peace_test.go — #1316's second proof card: the plainer
// of the two cast-ban shapes, an outright ban with no exception zone,
// only for the turn it was cast. The combat-ending half (#1317, CR
// 724.2) and the "cast only during combat" condition are pinned in
// combat_control_cards_test.go, which also owns the `mandateOfPeaceOracle`
// constant this file reuses.

// TestMandateOfPeaceLocksEveryOpponentForTheRestOfTheTurn is #1316's
// own proof: an outright ban, no exception zone, granted to each
// opponent, and gone at the very next cleanup step rather than
// surviving to the caster's own next turn.
func TestMandateOfPeaceLocksEveryOpponentForTheRestOfTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.CastInPlace(t, g, me.ID, "Mandate of Peace", effectstest.MandateOfPeaceOracle)
	effectstest.PassPriorityAroundTable(t, g)

	spell := effectstest.InstantCardFor(t, g)
	for _, opp := range []*game.Player{g.Seats[1], g.Seats[2], g.Seats[3]} {
		var banErr *game.CantCastError
		if err := effectstest.CastGateFor(g, opp.ID, spell, game.ZoneHand); !errors.As(err, &banErr) {
			t.Errorf("seat %s cast from hand: got %v, want a *CantCastError", opp.ID, err)
		}
	}
	if err := effectstest.CastGateFor(g, me.ID, spell, game.ZoneHand); err != nil {
		t.Errorf("the caster's own cast from hand was refused: %v", err)
	}

	// "This turn" — not "until your next turn": the ban is gone the
	// moment the FIRST opponent's turn begins, not after a whole
	// rotation. The combat-ending half already moved the cursor to
	// the postcombat main phase, so one advanceOneTurnForTest reaches
	// the first opponent's turn from there.
	effectstest.AdvanceOneTurnForTest(t, g)
	if err := effectstest.CastGateFor(g, g.Seats[1].ID, spell, game.ZoneHand); err != nil {
		t.Errorf("the ban survived the end of the turn it was cast on: %v", err)
	}
}

// TestMandateOfPeaceHasNoCaveats — #1317 (end the combat phase, PR #1343) and
// #1316 (the granted cast ban) together finish the card: three
// printed clauses, three shipped.
func TestMandateOfPeaceHasNoCaveats(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.MandateOfPeaceOracle)
	if !ok {
		t.Fatal("Mandate of Peace is not registered")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Mandate of Peace is %v with %d caveats, want full with none", spec.Completeness, len(spec.Caveats))
	}
}
