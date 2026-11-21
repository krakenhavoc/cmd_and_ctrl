package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elite Arrester — Creature — Human Soldier {W}, 0/3 (Ravnica Allegiance,
// #5521):
//
//	"{1}{U}, {T}: Tap target creature."
//
// Arashin Sunshield's tapper at a different cost: tapChosenPermanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f25af8de-3edb-400b-9b6c-ec446a8314fa",
		Name:         "Elite Arrester",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}, {T}: Tap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
