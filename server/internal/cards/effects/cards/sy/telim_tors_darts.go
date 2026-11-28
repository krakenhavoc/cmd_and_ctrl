package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Telim'Tor's Darts — Artifact {2} (Mirage, #5633):
//
//	"{2}, {T}: This artifact deals 1 damage to target player or planeswalker."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c978da4-9caa-4b4b-be52-889677962ec9",
		Name:         "Telim'Tor's Darts",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: This artifact deals 1 damage to target player or planeswalker.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetPlayerOrPlaneswalker(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
