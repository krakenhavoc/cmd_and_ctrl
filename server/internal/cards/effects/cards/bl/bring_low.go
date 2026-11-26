package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bring Low — Khans of Tarkir (S82 OLD b346, #5554):
//
//	Instant {3}{R}: Bring Low deals 3 damage to target creature. If that creature has
//	a +1/+1 counter on it, Bring Low deals 5 damage to it instead.
//
// The counter is read as the spell resolves.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cfe4d707-41b3-4897-b13a-dcae8f1f3350",
		Name:         "Bring Low",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			n := 3
			if c, found := ctx.Game.LookupCardForEffect(id); found && c.Counters[game.CounterPlusOne] > 0 {
				n = 5
			}
			return effects.DealDamage{Source: item.SourceCardID, Target: id, Amount: n}.Apply(ctx)
		},
	})
}
