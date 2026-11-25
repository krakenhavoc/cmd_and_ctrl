package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Winds of Rebuke — {1}{U} Instant (Amonkhet, #5535):
//
//	"Return target nonland permanent to its owner's hand. Each player mills
//	 two cards."
//
// The mills are one library at a time in APNAP order (Breach the
// Multiverse's shape). With the target gone the whole spell is countered on
// resolution (CR 608.2b), so nobody mills.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6840a63d-d627-4a18-afa9-018406767642",
		Name:         "Winds of Rebuke",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.BounceTheTarget(ctx.Item, ctx) }),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.MillEachPlayerInTurn(ctx, effects.ApnapPlayers(ctx.Game), effects.ConstantMill(2), game.KeyedTail{})
			}),
		),
	})
}
