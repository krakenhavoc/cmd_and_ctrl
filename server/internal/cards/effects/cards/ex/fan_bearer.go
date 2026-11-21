package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fan Bearer — Creature — Zombie {W}, 1/2
// (Amonkhet, #5535):
//
//	"{2}, {T}: Tap target creature."
//
// Gavony Trapper's ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2e7198a-0eb5-4d30-973f-4b2d9617436f",
		Name:         "Fan Bearer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Tap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
