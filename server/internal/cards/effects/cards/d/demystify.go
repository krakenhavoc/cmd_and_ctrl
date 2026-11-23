package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Demystify — Instant {W} (Rise of the Eldrazi, #5581):
//
//	"Destroy target enchantment."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fd591199-9f7a-4147-a150-13279dbb4498",
		Name:         "Demystify",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target enchantment", effects.Enchantment()),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
