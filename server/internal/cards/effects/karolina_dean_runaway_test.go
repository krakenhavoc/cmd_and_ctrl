package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #2811: Karolina Dean adds {W}{U}{B}{R}{G} at the beginning of its
// controller's first main phase, and none of it can cast a spell from
// hand.
func TestKarolinaDeanAddsFiveRestrictedMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Karolina Dean, Runaway", "Legendary Creature — Alien Hero", "50b5faff-acaa-426b-9f3e-1005c0392e6c", 3, 3)
	advanceTo(t, g, game.StepPrecombatMain)
	passPriorityAroundTable(t, g)
	if got := len(me.ManaPool); got != 5 {
		t.Fatalf("pool has %d mana, want five", got)
	}
	for _, tok := range me.ManaPool {
		if len(tok.Restrictions) != 1 || tok.Restrictions[0] != game.ManaRestrictNotFromHand {
			t.Errorf("token %+v, want only the not-from-hand restriction", tok)
		}
	}
}
