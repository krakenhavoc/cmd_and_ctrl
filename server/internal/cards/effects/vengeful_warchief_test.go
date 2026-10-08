package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const vengefulWarchiefOracle = "be3dd4f6-794d-4bce-b71e-4b3477d5d388"

// Vengeful Warchief (#2540): one +1/+1 counter per turn, on the first
// loss, however many creatures dealt the damage.
func TestVengefulWarchiefGrowsOncePerTurnOnTheFirstLoss(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	warchief := pushCatalogPermanent(g, foe.ID, "Vengeful Warchief", "Creature — Orc Warrior", vengefulWarchiefOracle, false)
	a := pr7Creature(g, me.ID, "Attacker A", 2, "R")
	b := pr7Creature(g, me.ID, "Attacker B", 3, "R")

	declareAttack(t, g, foe.ID, a, b)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, warchief, "+1/+1"); got != 1 {
		t.Errorf("+1/+1 counters %d after two simultaneous hits, want 1", got)
	}
	pr7Hit(t, g, a, foe.ID, 1)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, warchief, "+1/+1"); got != 1 {
		t.Errorf("+1/+1 counters %d after a second loss, want 1", got)
	}
}
