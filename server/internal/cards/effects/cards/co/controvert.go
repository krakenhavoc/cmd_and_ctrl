package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Controvert — Instant {2}{U}{U} (Coldsnap):
//
//	"Counter target spell.
//	 Recover {2}{U}{U}"
//
// Recover (CR 702.59a) is S82OldB564Recover.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f7ad79e3-d128-4aee-8542-99fd3527f9d3",
		Name:         "Controvert",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve:    effects.CounterTheTargetSpell,
		Triggered:    []game.TriggeredAbility{effects.S82OldB564Recover("Controvert", "{2}{U}{U}")},
	})
}
