package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mystic Melting — Instant {3}{G} (Coldsnap):
//
//	"Destroy target artifact or enchantment.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// The draw is a delayed trigger set up as the spell resolves (CR 603.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "850fe8ad-f6b8-4d01-a94b-2b69adc3d8f0",
		Name:         "Mystic Melting",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
				if err := (effects.DestroyTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.DrawACardAtTheNextUpkeep("Mystic Melting — draw a card").Apply(ctx)
		},
	})
}
