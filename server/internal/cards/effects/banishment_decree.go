package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "567b676b-54f3-44c7-8654-4d2b0262ea12",
		Name:         "Banishment Decree",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // removes a permanent
		Targets:      TargetPermanent("target artifact, creature, or enchantment", Or(Artifact(), Creature(), Enchantment())),
		OnResolve:    putTargetOnTopOfOwnersLibrary,
	})
}
