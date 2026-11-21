package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nine-Ringed Bo — Artifact {3}:
//
//	"{T}: This artifact deals 1 damage to target Spirit creature. If that
//	 creature would die this turn, exile it instead."
//
// Lava Coil's sentence on an ability: the replacement is the ability's,
// not the damage's, so the Spirit is marked whether or not the damage is
// dealt (ADR 0108 §1). "Spirit" is read off its effective subtypes, so a
// changeling counts.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5dc3dbd-7eab-40f3-afe2-1e88a0c00538",
		Name:         "Nine-Ringed Bo",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This artifact deals 1 damage to target Spirit creature. If that creature would die this turn, exile it instead.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target Spirit creature", effects.Subtype("Spirit")),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.AbilityBody(effects.DamageFirstTargetExileIfItDies(effects.FixedAmount(1))),
		}},
	})
}
