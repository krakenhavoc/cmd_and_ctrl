package aiseat_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// block_requirements_test.go — the bot half of #1597: a Lure'd 4/4 on
// the attacking side, two 2/2s on a bot defender's side that would
// rather not chump it. Whatever the policy wants, both Bears block the
// Lure'd attacker and the combat ends. driveOneCombat
// (combat_limits_test.go) fails on a refused enumerated move, on a
// priority holder that declines, and on a combat that never ends —
// the #544 wedges a withheld pass could cause.
func TestBotsObeyLure(t *testing.T) {
	policies := map[string]aiseat.Policy{
		"heuristic":  heuristic.New(),
		"aggressive": aggressive(),
		// Attacks when it can and never chooses a block of its own
		// accord: the defender that has to be made to block.
		"never-blocks": &scripted{prefer: []string{"Attack"}},
	}
	for name, pol := range policies {
		t.Run(name, func(t *testing.T) {
			g := limitTable(t, 2)
			active := g.Seats[g.Turn.ActiveSeat]
			def := g.Seats[(g.Turn.ActiveSeat+1)%2]
			var lured uuid.UUID
			for i := range g.Battlefield.Cards {
				if c := &g.Battlefield.Cards[i]; c.Controller == active.ID {
					lured = c.InstanceID
					break
				}
			}
			ok := false
			g.WithWriteLock(func() {
				ok = g.RegisterScopedEffectForEffect(uuid.New(), g.PinnedObjectsLocked(lured),
					[]game.Mod{
						game.AddBlockRequirementMod(game.BlockRequirementLure),
						// It attacks, so the Lure has something to do.
						game.AddAttackRequirementMod(uuid.Nil),
					}, game.IndefiniteDuration(), "test — Lure")
			})
			if !ok {
				t.Fatal("registered nothing")
			}
			onLured := map[uuid.UUID]bool{}
			driveOneCombatWatching(t, g, pol, func() {
				for i := range g.Battlefield.Cards {
					c := &g.Battlefield.Cards[i]
					if c.Controller == def.ID && c.BlockingTarget == lured {
						onLured[c.InstanceID] = true
					}
				}
			})
			if len(onLured) != 2 {
				t.Fatalf("%d of the defender's two creatures blocked the Lure'd attacker, want both", len(onLured))
			}
		})
	}
}
