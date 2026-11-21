package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:      "f6e1eb52-9f13-425c-8309-1b08ed6557db",
		Name:          "Barbed Lightning",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Entwine("{2}")},
		Modes: effects.Entwining(effects.ChooseOne(
			effects.ModeDoing("Barbed Lightning deals 3 damage to target creature.",
				effects.TargetCreature("target creature"), barbedLightningDamage),
			effects.ModeDoing("Barbed Lightning deals 3 damage to target player or planeswalker.",
				effects.TargetPlayerOrPlaneswalker(), barbedLightningDamage),
		)),
	})
}

// barbedLightningDamage is either bullet: 3 damage to the bullet's own
// target, if it is still legal (CR 608.2b).
var barbedLightningDamage = effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
	return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 3}.Apply(ctx)
})
