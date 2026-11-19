package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Band Together — Instant {2}{G} (Ravnica: Clue Edition):
//
//	"Up to two target creatures you control each deal damage equal to their
//	 power to another target creature."
//
// Allies at Last's shape: slot 0 the damage sources, slot 1 a creature that is
// none of them (Distinct, CR 115.3 "another"). Each source still legal deals
// damage equal to its current power in one damage event; with no source
// chosen nothing happens, and a victim that has left takes nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a8631e87-bfc3-463b-bdd1-974ea1d79ae4",
		Name:         "Band Together",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("up to two target creatures you control", YouControl()).WithCount(0, 2),
			Distinct(TargetCreature("another target creature")),
		),
		Purpose: game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return s82Bns18SourcesDamageTheOther(ctx)
		},
	})
}
