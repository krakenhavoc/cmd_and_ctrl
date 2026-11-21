package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ertai, Wizard Adept — Legendary Creature — Human Wizard {2}{U}, 1/1
// (Exodus, #5626):
//
//	"{2}{U}{U}, {T}: Counter target spell."
//
// Kozilek, the Great Distortion's activated counter with a mana-and-tap
// cost: counterTheChosenSpell counters the first target still legal as
// the ability resolves (CR 608.2b, CR 701.6a). The tap cost means it
// can't be activated the turn Ertai comes under your control
// (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f2902dfd-888f-4b02-b8ad-1056f0059fc3",
		Name:         "Ertai, Wizard Adept",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}{U}, {T}: Counter target spell.",
			Cost:    effects.Plus(effects.ManaCost("{2}{U}{U}"), effects.TapCost()),
			Targets: effects.TargetSpell("target spell"),
			Effect:  effects.CounterTheChosenSpell,
		}},
	})
}
