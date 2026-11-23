package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pacification Array — Artifact {1} (Aether Revolt,
// #5536):
//
//	"{2}, {T}: Tap target artifact or creature."
//
// Sterling Keykeeper's tap row with an artifact-or-creature target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b2c14195-2747-490e-99ef-5b71048ff51a",
		Name:         "Pacification Array",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Tap target artifact or creature.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
