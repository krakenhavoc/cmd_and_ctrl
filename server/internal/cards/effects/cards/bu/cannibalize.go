package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cannibalize — Sorcery {1}{B} (Stronghold, S82 OLD batch 748, #5627):
//
//	"Choose two target creatures controlled by the same player. Exile one of
//	 those creatures and put two +1/+1 counters on the other."
//
// Barrin's Spite's pair of related targets (the second controlled by the
// first's controller and another creature), judged as the spell is cast
// (CR 601.2c) and again as it resolves (CR 608.2b). The spell's controller, not
// the creatures' controller, chooses which one is exiled
// (S82OldB748Cannibalize); the other gets two +1/+1 counters. If one target has
// become illegal the other is the one exiled and nothing gets counters; with
// both illegal the spell does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bf044267-92b9-4983-964f-2bc4ecf76bce",
		Name:         "Cannibalize",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.Distinct(effects.ControlledByTargetsController(effects.TargetCreature("another target creature controlled by the same player"), 0)),
		),
		OnResolve: effects.S82OldB748Cannibalize,
	})
}
