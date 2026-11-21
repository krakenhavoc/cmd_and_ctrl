package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Totally Lost — Instant {4}{U} (Ravnica Remastered):
//
//	"Put target nonland permanent on top of its owner's library."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e0462a2c-fb88-495f-813b-6476ee3e62bb",
		Name:         "Totally Lost",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		OnResolve:    gapB79PutTargetOnTopOfItsOwnersLibrary,
	})
}
