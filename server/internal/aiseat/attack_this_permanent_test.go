package aiseat_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestBotsObeyAttackThisPermanentInFourSeats — #2567: a creature on a
// bot's side that must attack one opponent's planeswalker (Gideon Jura's
// requirement) attacks it and nothing else, and the combat ends under
// every policy. driveOneCombatWatching fails on a refused enumerated
// move and on a combat that never ends.
func TestBotsObeyAttackThisPermanentInFourSeats(t *testing.T) {
	for name, pol := range map[string]aiseat.Policy{"heuristic": heuristic.New(), "aggressive": aggressive()} {
		t.Run(name, func(t *testing.T) {
			g := limitTable(t, 4)
			active := g.Seats[g.Turn.ActiveSeat]
			owner := g.Seats[(g.Turn.ActiveSeat+2)%4]
			// A 1/1 the heuristic would rather keep home.
			bound := limitPush(g, active, "Bound Ogre", "Creature — Ogre", "", 1, 1)
			walker := limitPush(g, owner, "Gideon Jura", "Legendary Planeswalker — Gideon", "", 0, 0)
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == walker {
					g.Battlefield.Cards[i].Counters = map[string]int{game.CounterLoyalty: 6}
				}
			}
			g.WithWriteLock(func() {
				if !g.RegisterScopedEffectForEffect(walker, g.PinnedObjectsLocked(bound),
					[]game.Mod{game.AttackPermanentIfAbleMod(game.ObjectRef{ID: walker}, 0)},
					game.IndefiniteDuration(), "test — attacks Gideon Jura if able") {
					t.Fatal("registered nothing")
				}
			})
			attackedAt := map[uuid.UUID]bool{}
			driveOneCombatWatching(t, g, pol, func() {
				for i := range g.Battlefield.Cards {
					c := &g.Battlefield.Cards[i]
					if c.InstanceID == bound && c.AttackingTarget != uuid.Nil {
						attackedAt[c.AttackingTarget] = true
					}
				}
			})
			if !attackedAt[walker] || len(attackedAt) != 1 {
				t.Fatalf("the bound creature attacked %v, want only the planeswalker", attackedAt)
			}
		})
	}
}
