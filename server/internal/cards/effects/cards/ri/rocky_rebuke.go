package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rocky Rebuke — Instant {1}{G}:
//
//	"Target creature you control deals damage equal to its power to
//	 target creature an opponent controls."
//
// Tenderize's one-sided fight: two target clauses read by slot, the
// biter's power read as the spell resolves, and nothing dealt if either
// creature has become an illegal target (CR 608.2b).
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fd85db70-8616-42df-9dac-bad7923f88d7",
		Name:         "Rocky Rebuke",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		),
		OnResolve: effects.OneSidedBite,
	})
}
