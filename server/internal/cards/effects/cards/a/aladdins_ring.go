package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aladdin's Ring — Artifact {8} (Arabian Nights, #5641):
//
//	"{8}, {T}: This artifact deals 4 damage to any target."
//
// The Ring is the source of the damage, as it last existed if it left
// in response.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7810915f-df4d-445a-8ba8-fe6ed3abf1ae",
		Name:         "Aladdin's Ring",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{8}, {T}: This artifact deals 4 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 4)),
			Cost:    effects.Plus(effects.ManaCost("{8}"), effects.TapCost()),
			Targets: effects.TargetAny(),
			Effect:  effects.FourDamageToFirstLegalTarget,
		}},
	})
}
