package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prophetic Bolt — Instant {3}{U}{R} (Apocalypse, #5614):
//
//	"Prophetic Bolt deals 4 damage to any target. Look at the top four
//	 cards of your library. Put one of those cards into your hand and
//	 the rest on the bottom of your library in any order."
//
// Lightning Bolt's body at 4 (damageToFirstTarget), then Impulse
// (LookAtTopThenTakeOneRestOnBottom). The two sentences are two steps
// (ResolveSteps, CR 608.2c), so the look waits for anything the damage
// asks first. A Bolt whose only target is gone does not resolve at all
// (CR 608.2b), so it looks at nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d5e4bf5e-1a66-4315-8c82-640c27977b88",
		Name:         "Prophetic Bolt",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 4)),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DamageToFirstTargetStep(4)(ctx.Item, ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.LookAtTopThenTakeOneRestOnBottom(4, "Prophetic Bolt — put one of them into your hand")(ctx.Game, ctx.Item)
			}),
		),
	})
}
