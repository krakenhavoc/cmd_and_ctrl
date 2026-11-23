package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fry — Instant {1}{R} (Core Set 2020, #5518):
//
//	"This spell can't be countered.
//	 Fry deals 5 damage to target creature or planeswalker that's white or
//	 blue."
//
// Spec.CantBeCountered with a colour-gated damage target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3a747893-a514-4689-8486-ecccf654fecb",
		Name:            "Fry",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		Targets: effects.TargetPermanent("target creature or planeswalker that's white or blue",
			effects.And(effects.Or(effects.Creature(), effects.Planeswalker()), effects.Or(effects.OfColor("W"), effects.OfColor("U")))),
		Purpose:   effects.ForTargets(effects.DamageToTarget(0, 5)),
		OnResolve: effects.DamageToFirstTargetStep(5),
	})
}
