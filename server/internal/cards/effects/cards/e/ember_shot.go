package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ember Shot — Instant {6}{R} (Judgment, #5611):
//
//	"Ember Shot deals 3 damage to any target.
//	 Draw a card."
//
// Playful Shove at 3: the damage to the one target, then the caster's
// draw as the next step (CR 608.2c). A target that has gone counters the
// whole spell, so no card is drawn then (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f300320-ad98-4812-b34b-54885a86aee9",
		Name:         "Ember Shot",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(effects.DamageToTarget(0, 3))},
		OnResolve:    effects.ResolveSteps(effects.DamageTheClauseTarget(0, 3), effects.DrawCards{N: 1}),
	})
}
