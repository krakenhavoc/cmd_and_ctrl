package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Discovery // Dispersal — split card, oracle 8f1c1bd6-8420-4de2-81f4-5faa419da301:
//
//	Discovery {1}{U/B} Sorcery: "Surveil 2, then draw a card."
//	Dispersal {3}{U}{B} Instant: "Each opponent returns a nonland permanent
//	 they control with the greatest mana value among permanents they control
//	 to its owner's hand, then discards a card."
//
// Two specs, Dispersal under "<oracle_id>#1" as Dusk // Dawn is.
//
// Discovery is Consider's surveil-then-draw with a surveil of two.
//
// Dispersal's helpers are in cmdr2_s39_helpers.go. Each opponent picks for
// themself, on their own prompt (a player with one candidate is not asked),
// and the discard is that player's own pick, made after their permanent is
// back in hand, so the returned card may be the one they discard.
//
// No simplification.
func init() {
	const oracle = "8f1c1bd6-8420-4de2-81f4-5faa419da301"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Discovery",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue, Draws: 1},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return Surveil{Player: item.Controller, N: 2, ThenTail: ctx.Tail(thenDrawACard)}.Apply(ctx)
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Dispersal",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, opp := range ctx.Opponents() {
				if err := cmdr2S39OpponentReturnsGreatest(ctx, opp); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
