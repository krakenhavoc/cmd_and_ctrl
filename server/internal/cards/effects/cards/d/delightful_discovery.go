package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "705cc928-d6d4-4294-979b-a385d1b07dc6",
		Name:         "Delightful Discovery",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2, Scry: 2},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.OpponentSpellsCastThisTurn, "This spell costs {1} less to cast for each spell your opponents have cast this turn."),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.Scry{N: 2, ThenTail: ctx.TailN(effects.ThenDrawCards, 2)}.Apply(ctx)
		},
	})
}
