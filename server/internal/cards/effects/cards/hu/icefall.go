package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Icefall — Sorcery {2}{R}{R} (Coldsnap):
//
//	"Destroy target artifact or land.
//	 Recover {R}{R}"
//
// Recover (CR 702.59a) is S82OldB564Recover.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "30baca42-50e2-45f8-8f99-8e2f9795b4da",
		Name:         "Icefall",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact or land", effects.Or(effects.Artifact(), effects.Land())),
		OnResolve:    effects.DestroyTheTargetPermanent,
		Triggered:    []game.TriggeredAbility{effects.S82OldB564Recover("Icefall", "{R}{R}")},
	})
}
