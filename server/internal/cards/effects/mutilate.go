package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mutilate — Sorcery {2}{B}{B} (Foundations Commander):
//
//	"All creatures get -1/-1 until end of turn for each Swamp you
//	 control."
//
// Defile's count, over every creature. The Swamps are counted once, as
// Mutilate resolves, and only the creatures on the battlefield then are
// affected (CR 611.2c): neither changes later in the turn. "Swamp" is a word
// of Mutilate's text (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "698e79ec-c2bc-4973-a095-f3c57b50e39c",
		Name:     "Mutilate",
		// The amount is counted at resolution, like an X (AmountIsX).
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, AmountIsX: true}},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := swampsControlledBy(ctx.Game, ctx.Controller(), ctx.TextSubtype("Swamp"))
			if n <= 0 {
				return nil
			}
			return BoostUntilEOT{Match: Creature(), Power: -n, Toughness: -n,
				Label: "Mutilate — -1/-1 for each Swamp you control"}.Apply(ctx)
		},
	})
}
