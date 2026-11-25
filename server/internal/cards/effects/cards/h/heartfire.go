package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heartfire — Instant {1}{R} (Jumpstart, #5510):
//
//	"As an additional cost to cast this spell, sacrifice a creature or
//	 planeswalker.
//	 Heartfire deals 4 damage to any target."
//
// The sacrifice is the mandatory additional cost, paid as the spell is
// cast (CR 601.2h), so a caster with neither a creature nor a planeswalker
// cannot cast it; a creature or planeswalker the spell itself targets may
// be the one sacrificed (the spell then fizzles, CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "2f6f5054-4a48-458e-969e-3a0f0e507354",
		Name:           "Heartfire",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		Targets:        effects.TargetAny(),
		Purpose:        effects.ForTargets(effects.DamageToTarget(0, 4)),
		OnResolve:      effects.DamageToFirstTargetStep(4),
	})
}
