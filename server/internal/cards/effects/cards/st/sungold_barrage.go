package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sungold Barrage — Instant {2}{W} (Innistrad: Midnight Hunt, #5497):
//
//	"Destroy target creature with toughness 4 or greater."
//
// Murdock's Crusade's toughness predicate over Doom Blade's body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a33d0ea-9b19-4fde-8e7a-2a9d17f716ea",
		Name:         "Sungold Barrage",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature with toughness 4 or greater", effects.ToughnessGE(4)),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve:    effects.S82MidDestroyTheTarget,
	})
}
