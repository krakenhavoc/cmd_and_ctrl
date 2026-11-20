package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eat to Extinction — Instant {3}{B} (Theros Beyond Death, #5514):
//
//	"Exile target creature or planeswalker. Surveil 1. (Look at the
//	 top card of your library. You may put that card into your
//	 graveyard.)"
//
// Ray of Ruin's shape: the exile, then a surveil in place of the scry.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bf5f348e-d748-48e2-b483-740109e71176",
		Name:         "Eat to Extinction",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve:    ResolveSteps(ExileChosenTarget{}, Surveil{N: 1}),
	})
}
