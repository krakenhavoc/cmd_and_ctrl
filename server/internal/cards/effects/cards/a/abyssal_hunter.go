package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abyssal Hunter — Creature — Human Assassin {3}{B}, 1/1 (Mirage, #5633):
//
//	"{B}, {T}: Tap target creature. This creature deals damage equal to its power
//	 to that creature."
//
// Its power is read as the ability resolves, or as it last existed if the Hunter
// has left the battlefield (CR 608.2h); the damage comes from the Hunter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b96ccc2f-88e5-4522-924e-76fc76dc3b39",
		Name:         "Abyssal Hunter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, {T}: Tap target creature. This creature deals damage equal to its power to that creature.",
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.OldB795TapAndDealItsPower,
		}},
	})
}
