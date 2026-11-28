package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Disempower — Instant {1}{W} (Mirage, S82 OLD b799, #5633):
//
//	"Put target artifact or enchantment on top of its owner's library."
//
// Aethertow's tuck over an artifact or enchantment, to the top of its owner's
// library whoever controls it (CR 401.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff9f4bfe-1585-4eaa-943c-1e85ed79a820",
		Name:         "Disempower",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve:    effects.PutTargetOnTopOfOwnersLibrary,
	})
}
