package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Essence Burn — Instant {1}{R}:
//
//	"Essence Burn deals 5 damage to target black or green creature or
//	 planeswalker. If that permanent would die this turn, exile it
//	 instead."
//
// Lava Coil's sentence on a colour-restricted creature-or-planeswalker
// target. The replacement is the spell's, so the permanent is marked
// whether or not the damage is dealt (ADR 0108 §1).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f1216f36-00fd-4d03-85c1-f4910a02d4d0",
		Name:         "Essence Burn",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetPermanent("target black or green creature or planeswalker",
			effects.And(effects.Or(effects.Creature(), effects.Planeswalker()), effects.Or(effects.OfColor("B"), effects.OfColor("G")))),
		Purpose:   effects.ForTargets(effects.DamageToTarget(0, 5)),
		OnResolve: effects.DamageFirstTargetExileIfItDies(effects.FixedAmount(5)),
	})
}
