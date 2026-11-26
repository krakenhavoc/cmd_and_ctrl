package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Final Revels — Sorcery {4}{B} (Lorwyn, S82 OLD b545, #5593):
//
//	"Choose one —
//	 • All creatures get +2/+0 until end of turn.
//	 • All creatures get -0/-2 until end of turn."
//
// Two modes, each an until-end-of-turn boost over every creature on the
// battlefield as the spell resolves (CR 611.2c): the creatures that are
// there then, and no others.
//
// Purpose: the second mode is a -0/-2 sweep.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "978aef8e-6af8-4838-97ce-afb84baa9edc",
		Name:         "Final Revels",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2}},
		Modes: effects.ChooseOne(
			effects.ModeDoing("All creatures get +2/+0 until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.BoostUntilEOT{Match: effects.Creature(), Power: 2, Label: "Final Revels — +2/+0"}.Apply(ctx)
				}),
			effects.ModeDoing("All creatures get -0/-2 until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.BoostUntilEOT{Match: effects.Creature(), Toughness: -2, Label: "Final Revels — -0/-2"}.Apply(ctx)
				}),
		),
	})
}
