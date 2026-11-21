package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "14945b80-10ec-4205-98df-cc3e119ca57c",
		Name:         "Springsage Ritual",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve:    effects.DestroyTargetThenGainLife(4),
	})
}
