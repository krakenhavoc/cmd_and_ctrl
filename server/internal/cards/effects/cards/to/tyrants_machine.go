package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tyrant's Machine — Artifact {2} (Magic 2015, #5556):
//
//	"{4}, {T}: Tap target creature."
//
// Staff of Domination's {4}, {T} row on its own (tapChosenPermanent).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d9cb3d86-cb3a-4e2e-a24e-ed3cfe5233d0",
		Name:         "Tyrant's Machine",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Tap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
