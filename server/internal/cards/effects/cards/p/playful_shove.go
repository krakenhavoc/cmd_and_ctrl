package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Playful Shove — Sorcery {1}{R} (Bloomburrow, #4467):
//
//	"Playful Shove deals 1 damage to any target.
//	 Draw a card."
//
// Two steps (CR 608.2c), so the draw waits for any question the damage
// asks, as Flick a Coin's does. A target that left in response counters
// the spell (CR 608.2b), and then nothing is drawn either.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2094fb68-3859-401f-97f2-c8bae64a4822",
		Name:         "Playful Shove",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(effects.DamageToTarget(0, 1))},
		OnResolve:    effects.ResolveSteps(effects.DamageTheClauseTarget(0, 1), effects.DrawCards{N: 1}),
	})
}
