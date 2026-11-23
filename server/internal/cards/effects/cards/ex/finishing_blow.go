package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Finishing Blow — Instant {4}{B} (Core Set 2021, #5511):
//
//	"Destroy target creature or planeswalker."
//
// Hero's Downfall's clause over both types; the destroy honours
// indestructible and skips a target that left in response.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c8a6c58-f4af-4809-af86-a931313fa71e",
		Name:         "Finishing Blow",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyChosenPermanent(ctx.Game, item)
		},
	})
}
