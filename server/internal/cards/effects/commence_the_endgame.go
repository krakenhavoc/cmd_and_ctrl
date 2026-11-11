package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Commence the Endgame — Instant {4}{U}{U}:
//
//	"This spell can't be countered.
//	 Draw two cards, then amass Zombies X, where X is the number of cards
//	 in your hand."
//
// The two draws are the instruction before the amass, and X is read after
// them, from the hand as it stands then (CR 608.2c), so a draw that paused
// for a replacement is finished first (the amass is the draw's keyed
// continuation).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fedb1947-9fa5-4655-8d02-b370fa63dd16",
		Name:            "Commence the Endgame",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Draws: 2},
		CantBeCountered: true,
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return DrawCards{N: 2}.Apply(ctx)
			}),
			// X is read after the draws, from the hand as it stands then.
			Instruction(func(ctx *Context) error {
				p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
				if p == nil || p.Hand == nil {
					return nil
				}
				return Amass{Subtype: "Zombie", N: len(p.Hand.Cards)}.Apply(ctx)
			}),
		),
	})
}
