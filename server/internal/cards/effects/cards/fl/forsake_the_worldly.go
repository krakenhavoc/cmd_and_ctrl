package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forsake the Worldly — Instant {2}{W}
// (Amonkhet, #5535):
//
//	"Exile target artifact or enchantment.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Revoke Existence, plus cycling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f60fb81a-969a-476c-a227-5231bbed4ad4",
		Name:         "Forsake the Worldly",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileFirstTarget(ctx.Game, item)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
