package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Akroan Jailer — Creature — Human Soldier {W}, 1/1 (Magic Origins, #5549):
//
//	"{2}{W}, {T}: Tap target creature."
//
// Frostbridge Guard's activated tap.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e3ad2c77-2a41-4a4a-a6bf-e542d0162b88",
		Name:         "Akroan Jailer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}, {T}: Tap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{2}{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
