package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rabid Bite — Sorcery {1}{G} (Bloomburrow, #4467):
//
//	"Target creature you control deals damage equal to its power to
//	 target creature you don't control."
//
// Bite Down at sorcery speed, with a creature (not a planeswalker) as
// the second target. Two target clauses read positionally (oneSidedBite):
// slot 0 deals damage equal to its power as the spell resolves, slot 1
// takes it. Either one gone means no damage (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6d5dc34b-3eea-4b77-8db7-94bc30b14c4c",
		Name:         "Rabid Bite",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // one-sided fight, as Bite Down (#4734)
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		OnResolve: effects.OneSidedBite,
	})
}
