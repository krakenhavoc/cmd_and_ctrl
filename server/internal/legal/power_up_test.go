package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// power_up_test.go — the enumerator half of #4527. A power-up ability
// this object has already activated is not a move (CR 702.193a,
// "Activate this ability only once"): the engine refuses it with
// ErrPowerUpUsed, so a bot seat must never be offered it, and both read
// game.Game.PowerUpUsed.

const oracleBraveBrawler = "a1a791be-71e6-4c1f-8717-08ca62d40de2"

func TestASpentPowerUpAbilityIsNotEnumerated(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	brawler := battlefieldCard(g, active, game.Card{
		Name: "Brave Brawler", TypeLine: "Creature — Human Warrior Hero",
		ManaCost: "{1}{W}", Power: 2, Toughness: 1, OracleID: oracleBraveBrawler,
	})
	basics(g, active, "Plains", 5)
	advanceTo(t, g, game.StepPrecombatMain)

	acts := activationsOf(legal.EnumerateFor(g, active.ID), brawler)
	if len(acts) != 1 {
		t.Fatalf("an unspent power-up ability: want 1 activation offered, got %v", labels(acts))
	}
	dispatchAll(t, g, active.ID, acts)

	if err := g.ActivateCatalogAbility(active.ID, brawler, 0, game.ActivateAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("activate the power-up ability: %v", err)
	}
	basics(g, active, "Plains", 5)
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), brawler); len(acts) != 0 {
		t.Errorf("a spent power-up ability: want no activation offered, got %v — the engine "+
			"refuses it, so nothing may propose it", labels(acts))
	}
}
