package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Withdraw — Instant {U}{U} (Prophecy, #5617):
//
//	"Return target creature to its owner's hand. Then return another target
//	 creature to its owner's hand unless its controller pays {1}."
//
// The first creature is returned, then the second creature's controller is
// asked for {1} as the spell resolves, and the table waits for the answer
// (#5793, CR 118.12). A decline returns the second creature. Each target is
// re-checked on its own (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "626882c5-15a8-43ea-ab08-33a7e949c88c",
		Name:         "Withdraw",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.Distinct(effects.TargetCreature("another target creature")),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				first, ok := ctx.ClauseTarget(0)
				if !ok || first.Kind != game.TargetCard {
					return nil
				}
				return effects.BounceToHand{Target: first.ID}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				second, legal := ctx.ClauseTarget(1)
				if !legal || second.Kind != game.TargetCard {
					return nil
				}
				c, found := ctx.Game.LookupCardForEffect(second.ID)
				if !found {
					return nil
				}
				return effects.PayUnless{
					Chooser:  c.Controller,
					Cost:     "{1}",
					Question: "Withdraw — pay {1}, or " + c.Name + " returns to its owner's hand?",
					Else:     withdrawBounce,
				}.Apply(ctx)
			}),
		),
	})
}

// withdrawBounce is the decline: the second creature returns. Keyed, an
// on-disk identity.
var withdrawBounce = effects.PayUnlessElse("pay-unless/withdraw-return-to-hand",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		ts := ctx.Targets()
		if len(ts) < 2 || ts[1].Kind != game.TargetCard {
			return nil
		}
		return effects.BounceToHand{Target: ts[1].ID}.Apply(ctx)
	})
