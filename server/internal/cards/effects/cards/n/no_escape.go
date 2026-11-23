package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// No Escape — Instant {2}{U} (Jumpstart 2022, #5948):
//
//	"Counter target creature or planeswalker spell. If that spell is
//	 countered this way, exile it instead of putting it into its owner's
//	 graveyard.
//	 Scry 1."
//
// Deny Existence's counter-and-exile body, then the scry.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7c6e0198-edd8-42b8-ba1b-549e1713d188",
		Name:         "No Escape",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 1},
		Targets:      effects.TargetSpell("target creature or planeswalker spell", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.CounterTargetSpellAndExileIt(item, ctx); err != nil {
				return err
			}
			return effects.Scry{N: 1}.Apply(ctx)
		},
	})
}
