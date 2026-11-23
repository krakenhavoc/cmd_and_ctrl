package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pounce — Instant {1}{G} (Ixalan, #5532):
//
//	"Target creature you control fights target creature you don't control."
//
// Khalni Ambush's fight: two clauses (CR 115.1d), each re-checked at
// resolution; if either has become illegal, no damage is dealt
// (CR 701.14b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f4126dac-add4-4d5f-bf52-6f74c7c6aa67",
		Name:         "Pounce",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		Purpose:   game.Purpose{Answers: game.AnswerRemove},
		OnResolve: effects.ResolveSteps(effects.Instruction(effects.FightTargetClauses)),
	})
}
