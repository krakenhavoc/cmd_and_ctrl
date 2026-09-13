package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "8682266c-2b0f-496e-b1c6-1fb338feac24",
		Name:         "Perplexing Test",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Return all creature tokens to their owners' hands."),
			Mode("Return all nontoken creatures to their owners' hands."),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if ctx.HasMode(0) {
				if err := (BounceAllMatching{Match: And(Creature(), IsTokenPredicate())}).Apply(ctx); err != nil {
					return err
				}
			}
			if ctx.HasMode(1) {
				return BounceAllMatching{Match: And(Creature(), Not(IsTokenPredicate()))}.Apply(ctx)
			}
			return nil
		},
	})
}
