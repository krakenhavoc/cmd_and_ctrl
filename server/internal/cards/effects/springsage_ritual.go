package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Springsage Ritual — Instant {3}{G} (Eldritch Moon, #5541):
//
//	"Destroy target artifact or enchantment. You gain 4 life."
//
// Sanctify's shape with four life (destroyTargetThenGainLife): a target
// that is no longer legal makes the spell do nothing, the life included
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "14945b80-10ec-4205-98df-cc3e119ca57c",
		Name:         "Springsage Ritual",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve:    destroyTargetThenGainLife(4),
	})
}
