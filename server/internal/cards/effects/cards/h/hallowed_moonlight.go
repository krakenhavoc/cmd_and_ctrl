package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hallowed Moonlight — Instant {1}{W} (Magic Origins):
//
//	"Until end of turn, if a creature would enter and it wasn't cast, exile
//	 it instead.
//	 Draw a card."
//
// A timed replacement (CR 614.1, 614.6) held as a scoped record, so it
// outlives the spell and ends with the turn: Mistcaller's shape, but
// without "nontoken" — a creature token, created from no zone and never
// cast, is exiled instead of entering too (it ceases to exist, CR 111.7).
// A creature spell that resolves, and anything that isn't a creature,
// enters as usual. The draw follows, so the record is in place first.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9541db2-2874-4307-b896-c403f6d82e68",
		Name:         "Hallowed Moonlight",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				ctx.Game.ExileUncastCreaturesAndTokensForEffect(ctx.Source(), ctx.Controller(),
					ctx.Game.UntilEndOfTurnDuration(), "Hallowed Moonlight: exile creatures that weren't cast instead of entering")
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DrawCards{N: 1}.Apply(ctx)
			}),
		),
	})
}
