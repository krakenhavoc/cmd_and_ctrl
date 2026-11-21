package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aggressive Instinct — Sorcery {1}{G} (Foundations Jumpstart, #5929):
//
//	"Target creature you control deals damage equal to its power to target
//	 creature you don't control."
//
// Bite Down's one-sided fight (oneSidedBite) with the second slot narrowed to
// creatures: slot 0 deals, its power read as the spell resolves; slot 1 takes.
// Each slot's predicate is enforced at announce (CR 601.2c) and re-checked at
// resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d1c8f14b-7977-488a-8bd9-cefd6687f017",
		Name:         "Aggressive Instinct",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		OnResolve: effects.OneSidedBite,
	})
}
