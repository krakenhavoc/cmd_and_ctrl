package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crisis of Conscience — Sorcery {4}{W}{W} (Doctor Who):
//
//	"Choose one —
//	 • Destroy all tokens.
//	 • Destroy all nonland, nontoken permanents."
//
// Two board wipes (Wrath of God's DestroyAllMatching), each declaring its
// Purpose: every token, or every nonland permanent that is not a token.
// Regeneration and indestructible still answer both.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d5862292-9ef9-4bf2-b468-45bfe0d70f7a",
		Name:         "Crisis of Conscience",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Destroy all tokens.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return DestroyAllMatching{Match: IsTokenPredicate()}.Apply(ctx)
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepDestroy, Partial: true}}),
			ModeWithPurpose(ModeDoing("Destroy all nonland, nontoken permanents.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return DestroyAllMatching{Match: And(Nonland(), Not(IsTokenPredicate()))}.Apply(ctx)
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}}),
		),
	})
}
