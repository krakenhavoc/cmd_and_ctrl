package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ironwright's Cleansing — Sorcery {2}{W} (Eldritch Moon, #5541):
//
//	"Exile target artifact or enchantment."
//
// Revoke Existence's text and body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "599faf5b-88fd-4373-b6ce-6047a67b30ea",
		Name:         "Ironwright's Cleansing",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileFirstTarget(ctx.Game, item)
		},
	})
}
