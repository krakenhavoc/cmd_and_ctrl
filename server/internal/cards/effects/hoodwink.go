package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hoodwink — Instant {1}{U}:
//
//	"Return target artifact, enchantment, or land to its owner's hand."
//
// The shared bounceTheTarget body: the permanent goes to its OWNER's hand,
// and a target gone by resolution does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a2fe10b8-857e-4881-85ef-24a0f96a3d79",
		Name:         "Hoodwink",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact, enchantment, or land", Or(Artifact(), Enchantment(), Land())),
		OnResolve:    bounceTheTarget,
	})
}
