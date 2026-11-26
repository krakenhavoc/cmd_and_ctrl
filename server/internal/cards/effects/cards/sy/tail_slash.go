package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tail Slash — Instant {2}{R} (Dragons of Tarkir):
//
//	"Target creature you control deals damage equal to its power to target
//	 creature you don't control."
//
// A one-sided bite (CR 120.3): the first creature deals the damage and takes none,
// and nothing is dealt unless both targets are still legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "69c51904-f758-4853-a86c-317c04462ea2",
		Name:         "Tail Slash",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		OnResolve: effects.OneSidedBite,
	})
}
