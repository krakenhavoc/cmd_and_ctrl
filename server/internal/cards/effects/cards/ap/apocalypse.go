package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Apocalypse — Sorcery {2}{R}{R}{R} (Tempest, S82 OLD, #5628):
//
//	"Exile all permanents. You discard your hand."
//
// Two instructions in printed order (CR 608.2c): Worldfire's exile sweep over
// every permanent, tokens and lands included, then the caster discards the
// hand that is left (the spell itself has left the hand to be cast).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "82c8f5dd-563d-4fd0-bd43-7ee2001d3777",
		Name:         "Apocalypse",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepExile}},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.ExileAllMatching{Match: effects.Permanent()}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				_, err := effects.DiscardWholeHand(ctx.Game, ctx.Controller())
				return err
			}),
		),
	})
}
