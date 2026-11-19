package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Demystify — Instant {W} (Rise of the Eldrazi, #5581):
//
//	"Destroy target enchantment."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fd591199-9f7a-4147-a150-13279dbb4498",
		Name:         "Demystify",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target enchantment", Enchantment()),
		OnResolve:    destroyTheTargetPermanent,
	})
}
