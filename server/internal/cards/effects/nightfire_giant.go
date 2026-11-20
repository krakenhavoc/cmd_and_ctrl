package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2f351af6-0c94-4881-b91a-0211645c43c8",
		Name:         "Nightfire Giant",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 as long as you control a Mountain.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfWhileYouControlA(Subtype("Mountain")),
			Apply:     pumpBy(1, 1),
		}},
		Activated: []ActivatedAbility{{
			Label:   "{4}{R}: This creature deals 2 damage to any target.",
			Cost:    ManaCost("{4}{R}"),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 2)),
			Effect:  sourceDealsDamageToEachLegalTarget(2),
		}},
	})
}
