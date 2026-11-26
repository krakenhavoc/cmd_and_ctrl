package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ignite Disorder — Instant {1}{R} (Magic 2010, #5585):
//
//	"Ignite Disorder deals 3 damage divided as you choose among one, two,
//	 or three target white and/or blue creatures."
//
// Arc Lightning's division over creatures only, with Combust's colour
// filter: the division is announced with the targets (CR 601.2d) and each
// target gets at least 1.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f66317c-06fe-4320-9d98-eaedaef67c85",
		Name:         "Ignite Disorder",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCreature("one, two, or three target white and/or blue creatures", effects.Or(effects.OfColor("W"), effects.OfColor("U"))).
			WithCount(1, 3).Dividing(effects.Divide(3)),
		Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DealDividedDamage(ctx)
		},
	})
}
