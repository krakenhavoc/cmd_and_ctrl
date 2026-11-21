package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "b3c12848-528d-4005-878f-4f950a07240a",
		Name:         "Prey Upon",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		OnResolve: effects.ResolveSteps(effects.Instruction(effects.FightTargetClauses)),
	})
}
