package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eye of Nowhere — Sorcery — Arcane {U}{U} (Champions of Kamigawa,
// #5603):
//
//	"Return target permanent to its owner's hand."
//
// Boomerang at sorcery speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "27ffdf11-bb7f-40bf-94c4-6bfbc41668e5",
		Name:         "Eye of Nowhere",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target permanent"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve:    bounceTheTarget,
	})
}
