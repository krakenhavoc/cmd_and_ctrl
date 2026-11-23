package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nightfire Giant — Creature — Zombie Giant {4}{B}, 4/3 (Magic 2015,
// #5556):
//
//	"This creature gets +1/+1 as long as you control a Mountain.
//	 {4}{R}: This creature deals 2 damage to any target."
//
// Kird Ape's land-gated pump with a Mountain, and Ogre Shaman's 2 damage
// to any target behind a mana-only cost (sourceDealsDamageToEachLegalTarget).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f351af6-0c94-4881-b91a-0211645c43c8",
		Name:         "Nightfire Giant",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 as long as you control a Mountain.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfWhileYouControlA(effects.Subtype("Mountain")),
			Apply:     effects.PumpBy(1, 1),
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{R}: This creature deals 2 damage to any target.",
			Cost:    effects.ManaCost("{4}{R}"),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(2),
		}},
	})
}
