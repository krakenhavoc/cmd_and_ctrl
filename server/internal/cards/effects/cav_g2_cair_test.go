package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fall of Cair Andros reads deathtouch: one point of damage from a
// deathtouch source is lethal, so the rest of the damage is excess. A
// 4-damage hit to a 2/5 is 3 excess from a deathtouch source and none from
// any other (#3172, cav-g2).
func TestFallOfCairAndrosCountsExcessPastDeathtouchLethal(t *testing.T) {
	for _, tc := range []struct {
		name       string
		deathtouch bool
		counters   int
	}{
		{"deathtouch source", true, 3},
		{"ordinary source", false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
			pushCatalogPermanent(g, me.ID, "Fall of Cair Andros", "Enchantment", fallOfCairOracle, false)
			victim := pushDiesCreatureForTest(g, opp.ID, "Big Bear", "", "Creature — Bear", 2, 5)
			var keywords []string
			if tc.deathtouch {
				keywords = []string{"deathtouch"}
			}
			src := pushKeywordCreature(g, me.ID, "Source", keywords...)
			g.WithWriteLock(func() {
				if err := g.DealDamageToCreatureForEffect(src, victim, 4); err != nil {
					t.Fatalf("damage: %v", err)
				}
			})
			passPriorityAroundTable(t, g)
			passPriorityAroundTable(t, g)
			got := 0
			for _, c := range g.Battlefield.Cards {
				if c.Controller == me.ID && game.IsArmy(c) {
					got = c.Counters[game.CounterPlusOne]
				}
			}
			if got != tc.counters {
				t.Errorf("Army counters = %d, want %d", got, tc.counters)
			}
		})
	}
}
