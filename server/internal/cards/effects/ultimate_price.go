package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ultimate Price — Instant {1}{B} (Return to Ravnica, #5566):
//
//	"Destroy target monocolored creature."
//
// Sultai Charm's first bullet as a whole spell: a creature of exactly one
// color (colorless and multicolored creatures are not legal), re-checked
// at resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d740453b-d8f8-48d0-8c50-72a7cdb791b5",
		Name:         "Ultimate Price",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target monocolored creature", b41Monocolored()),
		OnResolve:    destroyTheTargetPermanent,
	})
}
