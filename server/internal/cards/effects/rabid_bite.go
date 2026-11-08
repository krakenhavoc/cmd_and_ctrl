package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "6d5dc34b-3eea-4b77-8db7-94bc30b14c4c",
		Name:         "Rabid Bite",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // one-sided fight, as Bite Down (#4734)
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature you don't control", OpponentControls()),
		),
		OnResolve: oneSidedBite,
	})
}
