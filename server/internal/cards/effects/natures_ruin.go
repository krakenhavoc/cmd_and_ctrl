package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nature's Ruin — Sorcery {2}{B} (Portal, #5630):
//
//	"Destroy all green creatures."
//
// Mass Calcify's sweep, turned round: only green creatures die.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d119ca9a-ee82-42c2-81e6-684205ca5183",
		Name:         "Nature's Ruin",
		Completeness: CompletenessFull,
		// ADR 0126 §6: a creature that isn't green stays.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: And(Creature(), OfColor("G"))}.Apply(ctx)
		},
	})
}
