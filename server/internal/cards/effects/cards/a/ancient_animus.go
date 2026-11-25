package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ancient Animus — Instant {1}{G} (Commander Legends):
//
//	"Put a +1/+1 counter on target creature you control if it's legendary.
//	 Then it fights target creature an opponent controls."
//
// Two target clauses, as Ruthless Predation's. The counter is its own step
// gated on the first target being legendary as the spell resolves, read
// through the layered view; the fight follows, so it reads the new power
// (CR 608.2c) and happens only while both targets are still legal
// (CR 701.14b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4c9a38f8-8fcb-45f9-8b70-e54298c7b485",
		Name:         "Ancient Animus",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok || !effects.IsLegendary(&c) {
					return nil
				}
				return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
			effects.Instruction(effects.FightTargetClauses),
		),
	})
}
