package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunder from Within — Sorcery — Arcane {2}{R}{R} (Saviors of Kamigawa,
// #5601):
//
//	"Destroy target artifact or land."
//
// Pillage without "it can't be regenerated": a regeneration shield saves
// the target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "09f75f28-09b9-4a6e-a1b0-671208d274d3",
		Name:         "Sunder from Within",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact or land", effects.Or(effects.Artifact(), effects.Land())),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
