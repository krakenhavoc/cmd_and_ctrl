package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2094fb68-3859-401f-97f2-c8bae64a4822",
		Name:         "Playful Shove",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(DamageToTarget(0, 1))},
		OnResolve:    ResolveSteps(damageTheClauseTarget(0, 1), DrawCards{N: 1}),
	})
}
