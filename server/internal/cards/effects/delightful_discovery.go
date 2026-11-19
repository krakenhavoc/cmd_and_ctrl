package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Delightful Discovery — Instant {4}{U} (Foundations Jumpstart, #5929):
//
//	"This spell costs {1} less to cast for each spell your opponents have
//	 cast this turn.
//	 Scry 2, then draw two cards."
//
// The discount reads the per-turn cast tally of every opponent when the
// spell is priced and reduces only the generic part (CR 601.2f). The
// draw waits for the scry's answer (Champions from Beyond), so the cards
// left on top are the cards drawn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "705cc928-d6d4-4294-979b-a385d1b07dc6",
		Name:         "Delightful Discovery",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2, Scry: 2},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(opponentSpellsCastThisTurn, "This spell costs {1} less to cast for each spell your opponents have cast this turn."),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return Scry{N: 2, ThenTail: ctx.TailN(thenDrawCards, 2)}.Apply(ctx)
		},
	})
}
