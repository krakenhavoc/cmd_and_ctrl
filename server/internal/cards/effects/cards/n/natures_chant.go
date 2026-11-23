package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nature's Chant — Instant {1}{G/W} (Modern Horizons, #5519):
//
//	"Destroy target artifact or enchantment."
//
// Disenchant with a hybrid cost: destroyTheTargetPermanent on an artifact
// or enchantment.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "46e4ace2-e8bf-4b72-81de-286215886a10",
		Name:         "Nature's Chant",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
