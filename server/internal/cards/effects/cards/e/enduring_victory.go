package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enduring Victory — Instant {4}{W} (Dragons of Tarkir, S82 old batch 330):
//
//	"Destroy target attacking or blocking creature. Bolster 1. (Choose a
//	 creature with the least toughness among creatures you control and put a
//	 +1/+1 counter on it.)"
//
// Righteous Blow's attacking-or-blocking target, destroyed, then bolster 1 in
// printed order (CR 608.2c). A target that is no longer attacking or blocking
// is illegal, and the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "955ff3bf-4323-452e-a8b1-b7662034473d",
		Name:         "Enduring Victory",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		OnResolve:    effects.DestroyTargetThenBolster(1),
	})
}
