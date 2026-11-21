package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Venom Blast — {2}{G}{G} Sorcery (Marvel's Spider-Man, #4460):
//
//	"Put two +1/+1 counters on target creature you control. It deals
//	 damage equal to its power to up to one other target creature."
//
// Two target clauses read positionally (slot 0 grows and bites, slot 1
// takes), the second "up to one" and Distinct so it cannot be the first.
// The layers are recomputed after the counters so the bite reads the
// grown power; oneSidedBite is the shared one-sided damage body. With no
// victim chosen only the counters go on.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c4b3dc5e-651f-4307-9d76-d728e5054a36",
		Name:         "Venom Blast",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.Distinct(effects.TargetCreature("up to one other target creature").WithCount(0, 1)),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				ctx.Game.RecomputeLayersIfStaleLocked()
				return effects.OneSidedBite(ctx.Item, ctx)
			}),
		),
	})
}
