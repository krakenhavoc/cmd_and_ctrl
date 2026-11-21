package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Perplexing Test — Instant {3}{U}{U} (EDHREC rank 2111):
//
//	"Choose one —
//	 • Return all creature tokens to their owners' hands.
//	 • Return all nontoken creatures to their owners' hands."
//
// A one-sided Evacuation for whichever half of the table you are
// not: the token deck bounces everyone's real creatures and keeps its
// army, the creature deck sends the tokens home (where they cease to
// exist). Each mode is one BounceAllMatching sweep — one simultaneous
// event, the S23 primitive — with the token / nontoken predicate on
// creatures.
//
// No simplification. (The file is perplexing_test_card.go because
// perplexing_test.go would be a Go test file.)
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8682266c-2b0f-496e-b1c6-1fb338feac24",
		Name:         "Perplexing Test",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.Mode("Return all creature tokens to their owners' hands."), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepBounce, Partial: true}}),
			effects.ModeWithPurpose(effects.Mode("Return all nontoken creatures to their owners' hands."), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepBounce, Partial: true}}),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if ctx.HasMode(0) {
				if err := (effects.BounceAllMatching{Match: effects.And(effects.Creature(), effects.IsTokenPredicate())}).Apply(ctx); err != nil {
					return err
				}
			}
			if ctx.HasMode(1) {
				return effects.BounceAllMatching{Match: effects.And(effects.Creature(), effects.Not(effects.IsTokenPredicate()))}.Apply(ctx)
			}
			return nil
		},
	})
}
