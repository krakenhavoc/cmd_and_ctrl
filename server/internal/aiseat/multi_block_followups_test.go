package aiseat_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// multi_block_followups_test.go — the bot half of #1715. Three 4/4s
// that must attack, against a defender whose 2/2 carries Blaze of
// Glory's record ("can block any number of creatures this turn; it
// blocks each attacking creature this turn if able"). Whatever the
// policy wants — including one that never blocks of its own accord —
// the 2/2 ends up blocking all three, the CR 510.1d division it owes
// is answered, and the combat ends.
func TestBotsObeyBlazeOfGlory(t *testing.T) {
	policies := map[string]aiseat.Policy{
		"heuristic":    heuristic.New(),
		"aggressive":   aggressive(),
		"never-blocks": &scripted{prefer: []string{"Attack"}},
	}
	for name, pol := range policies {
		t.Run(name, func(t *testing.T) {
			g := limitTable(t, 2)
			active := g.Seats[g.Turn.ActiveSeat]
			def := g.Seats[(g.Turn.ActiveSeat+1)%2]
			var ogres []uuid.UUID
			blazed := uuid.Nil
			for i := range g.Battlefield.Cards {
				c := &g.Battlefield.Cards[i]
				switch {
				case c.Controller == active.ID:
					ogres = append(ogres, c.InstanceID)
				case c.Controller == def.ID && blazed == uuid.Nil:
					blazed = c.InstanceID
				}
			}
			g.WithWriteLock(func() {
				g.RegisterScopedEffectForEffect(uuid.New(), g.PinnedObjectsLocked(ogres...),
					[]game.Mod{game.AddAttackRequirementMod(uuid.Nil)}, game.IndefiniteDuration(), "test — attacks")
				g.RegisterScopedEffectForEffect(uuid.New(), g.PinnedObjectsLocked(blazed),
					[]game.Mod{game.BlockAnyNumberMod(), game.AddBlockRequirementMod(game.BlockRequirementBlocksEach)},
					game.IndefiniteDuration(), "test — Blaze of Glory")
			})
			most, divided := driveMultiBlockCombat(t, g, pol)
			if most != 3 || !divided {
				t.Fatalf("the blazed creature blocked at most %d attackers (divided %v), want all three and a division", most, divided)
			}
		})
	}
}
