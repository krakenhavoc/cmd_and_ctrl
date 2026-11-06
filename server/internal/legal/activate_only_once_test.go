package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// activate_only_once_test.go — the enumerator half of #5221. An ability
// that says "Activate only once" (CR 602.5b), without the exhaust
// keyword, is not a move once this object has activated it: the engine
// refuses it with ErrAbilityActivatedOnce, so a bot seat must never be
// offered it, and both read game.Game.OnlyOnceSpent.

const oracleHauntedScreen = "49f92c0b-16ac-4cf0-8ec3-e3638990ee75"

func TestASpentOnlyOnceAbilityIsNotEnumerated(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	screen := battlefieldCard(g, active, game.Card{
		Name: "Haunted Screen", TypeLine: "Artifact", ManaCost: "{3}", OracleID: oracleHauntedScreen,
	})
	basics(g, active, "Plains", 7)
	advanceTo(t, g, game.StepPrecombatMain)

	acts := activationsOf(legal.EnumerateFor(g, active.ID), screen)
	if len(acts) != 1 {
		t.Fatalf("an unspent only-once ability: want 1 activation offered, got %v", labels(acts))
	}
	dispatchAll(t, g, active.ID, acts)

	if err := g.ActivateCatalogAbility(active.ID, screen, 0, game.ActivateAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("activate the {7} ability: %v", err)
	}
	basics(g, active, "Plains", 7)
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), screen); len(acts) != 0 {
		t.Errorf("a spent only-once ability: want no activation offered, got %v — the engine "+
			"refuses it, so nothing may propose it", labels(acts))
	}
}
