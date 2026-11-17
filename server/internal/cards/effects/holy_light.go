package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Holy Light — Instant {2}{W} (The Dark, #5638):
//
//	"Nonwhite creatures get -1/-1 until end of turn."
//
// Hideous Laughter's body over nonwhite creatures only. The set is
// fixed as it resolves, read through the creatures' current colours,
// so a creature that turned white in response is spared and one that
// arrives later this turn is not affected.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8f7dd328-b367-4546-ba79-1efc1e97f2ab",
		Name:         "Holy Light",
		Completeness: CompletenessFull,
		// ADR 0126 §6: a white creature is spared.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return BoostUntilEOT{Match: And(Creature(), Not(OfColor("W"))), Power: -1, Toughness: -1, Label: "Holy Light — -1/-1"}.Apply(ctx)
		},
	})
}
