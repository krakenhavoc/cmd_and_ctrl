package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Banishment Decree — Instant {3}{W}{W} (Mirrodin Besieged,
// #5576):
//
//	"Put target artifact, creature, or enchantment on top of its
//	 owner's library."
//
// Temporal Spring's tuck with the clause narrowed to an artifact, creature
// or enchantment: the permanent goes on top of its owner's library, not its
// controller's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "567b676b-54f3-44c7-8654-4d2b0262ea12",
		Name:         "Banishment Decree",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // removes a permanent
		Targets:      effects.TargetPermanent("target artifact, creature, or enchantment", effects.Or(effects.Artifact(), effects.Creature(), effects.Enchantment())),
		OnResolve:    effects.PutTargetOnTopOfOwnersLibrary,
	})
}
