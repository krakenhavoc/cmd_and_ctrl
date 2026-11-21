package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "d5862292-9ef9-4bf2-b468-45bfe0d70f7a",
		Name:         "Crisis of Conscience",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Destroy all tokens.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DestroyAllMatching{Match: effects.IsTokenPredicate()}.Apply(ctx)
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepDestroy, Partial: true}}),
			effects.ModeWithPurpose(effects.ModeDoing("Destroy all nonland, nontoken permanents.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DestroyAllMatching{Match: effects.And(effects.Nonland(), effects.Not(effects.IsTokenPredicate()))}.Apply(ctx)
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}}),
		),
	})
}
