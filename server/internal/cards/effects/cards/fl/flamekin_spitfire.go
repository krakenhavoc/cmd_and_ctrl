package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flamekin Spitfire — Creature — Elemental Shaman {1}{R}, 1/1 (Lorwyn):
//
//	"{3}{R}: This creature deals 1 damage to any target."
//
// Ballista Wielder's ping without the rider: a mana-only ability, so it can be
// activated again and again, summoning sick or not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b6f2bec3-cbff-40a5-a2ed-ec078f9fabf5",
		Name:         "Flamekin Spitfire",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{R}: This creature deals 1 damage to any target.",
			Cost:    effects.ManaCost("{3}{R}"),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
