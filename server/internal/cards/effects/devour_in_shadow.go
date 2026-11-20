package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Devour in Shadow — Instant {B}{B} (Fifth Dawn, #5604):
//
//	"Destroy target creature. It can't be regenerated. You lose life
//	 equal to that creature's toughness."
//
// Vendetta's body (vendettaResolve) for any creature: the toughness is read
// as the spell resolves, before the destruction, and the life is lost even if
// the creature isn't destroyed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9086d8ff-e375-4149-996d-2a88a7596c1c",
		Name:         "Devour in Shadow",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		OnResolve:    vendettaResolve,
	})
}
