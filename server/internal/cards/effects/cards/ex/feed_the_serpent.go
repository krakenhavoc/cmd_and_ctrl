package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feed the Serpent — Instant {2}{B}{B} (Kaldheim, #5505):
//
//	"Exile target creature or planeswalker."
//
// Wander Off's exile (exileTheFirstTargetIfLegal) on Never's "creature
// or planeswalker" target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ad404b46-dee7-41bb-969d-35b4cbaadece",
		Name:         "Feed the Serpent",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ExileTheFirstTargetIfLegal(ctx)
		},
	})
}
