package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Barbed Lightning — Instant {2}{R}:
//
//	"Choose one —
//	 • Barbed Lightning deals 3 damage to target creature.
//	 • Barbed Lightning deals 3 damage to target player or planeswalker.
//	 Entwine {2} (Choose both if you pay the entwine cost.)"
//
// Entwine (CR 702.42, #3866): {2} more takes both, each bullet with its
// own target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "f6e1eb52-9f13-425c-8309-1b08ed6557db",
		Name:          "Barbed Lightning",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Entwine("{2}")},
		Modes: Entwining(ChooseOne(
			ModeDoing("Barbed Lightning deals 3 damage to target creature.",
				TargetCreature("target creature"), barbedLightningDamage),
			ModeDoing("Barbed Lightning deals 3 damage to target player or planeswalker.",
				targetPlayerOrPlaneswalker(), barbedLightningDamage),
		)),
	})
}

// barbedLightningDamage is either bullet: 3 damage to the bullet's own
// target, if it is still legal (CR 608.2b).
var barbedLightningDamage = rfCreatureBModeTargetDoes(func(ctx *Context, id uuid.UUID) error {
	return DealDamage{Source: ctx.Source(), Target: id, Amount: 3}.Apply(ctx)
})
