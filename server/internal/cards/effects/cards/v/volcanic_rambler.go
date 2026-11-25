package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Volcanic Rambler — Creature — Elemental {5}{R}, 6/4 (Magic Origins):
//
//	"{2}{R}: This creature deals 1 damage to target player or
//	 planeswalker."
//
// Chandra's Magmutt's ping with a mana cost in place of the tap.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97f929ce-3eb3-4040-972f-0b21d5f2f61f",
		Name:         "Volcanic Rambler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}: This creature deals 1 damage to target player or planeswalker.",
			Cost:    effects.ManaCost("{2}{R}"),
			Targets: effects.TargetPlayerOrPlaneswalker(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsOneToFirstTarget,
		}},
	})
}
