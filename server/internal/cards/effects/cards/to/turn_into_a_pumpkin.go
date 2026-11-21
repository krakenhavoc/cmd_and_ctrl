package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Turn into a Pumpkin — Instant {3}{U} (Throne of Eldraine, #5516):
//
//	"Return target nonland permanent to its owner's hand. Draw a card.
//	 Adamant — If at least three blue mana was spent to cast this spell, create a Food token."
//
// The bounce, the draw and the adamant Food, each a step that waits for any
// prompt the one before it asked (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "cfda6f65-3755-40e5-8061-a21941c4b1bd",
		Name:           "Turn into a Pumpkin",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"U"},
		Purpose:        game.Purpose{Draws: 1},
		Targets:        effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.BounceTheTarget(ctx.Item, ctx) }),
			effects.DrawCards{N: 1},
			effects.EldB1AdamantSpentFood("U"),
		),
	})
}
