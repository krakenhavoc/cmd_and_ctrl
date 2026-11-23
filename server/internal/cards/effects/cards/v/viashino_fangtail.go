package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viashino Fangtail — Creature — Lizard Warrior {2}{R}{R}, 3/3 (Ravnica:
// City of Guilds, #5600):
//
//	"{T}: This creature deals 1 damage to any target."
//
// Thornwind Faeries' tap ping: the creature is the source of the damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5140cd2f-af30-4405-b5e4-996ef6349f59",
		Name:         "Viashino Fangtail",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to any target.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
