package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prey Upon — Sorcery {G} (Innistrad, #5572):
//
//	"Target creature you control fights target creature you don't
//	 control. (Each deals damage equal to its power to the other.)"
//
// Chelonian Tackle's fight (fightTargetClauses) on its own: two clauses,
// each re-checked at resolution; if either creature is gone, no damage
// is dealt (CR 701.14b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b3c12848-528d-4005-878f-4f950a07240a",
		Name:         "Prey Upon",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature you don't control", OpponentControls()),
		),
		OnResolve: ResolveSteps(Instruction(fightTargetClauses)),
	})
}
