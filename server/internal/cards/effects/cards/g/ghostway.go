package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ghostway — Instant {2}{W} (Ravnica Remastered):
//
//	"Exile each creature you control. Return those cards to the battlefield
//	 under their owner's control at the beginning of the next end step."
//
// The set is every creature the caster controls as the spell resolves (CR
// 611.2c), exiled together with one delayed return (the batch exile and
// delayed trigger Eerie Interlude uses); the cards come back under their
// owners' control at the next end step, as new objects (CR 400.7). A token
// ceases to exist in exile and does not return (CR 111.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8981af1-b0aa-4c19-8520-a7fb59018bc8",
		Name:         "Ghostway",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerSurvive},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.YouControl())) {
				ids = append(ids, c.InstanceID)
			}
			return effects.ExileCardsThenScheduleReturn(ctx, ids, "Ghostway — return the exiled creatures")
		},
	})
}
