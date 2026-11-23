package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frostbridge Guard — Creature — Elemental Soldier {1}{W}, 2/2 (Wilds of
// Eldraine, #4471):
//
//	"{2}{W}, {T}: Tap target creature."
//
// Kytheon's Irregulars' tap ability at {2}{W} plus {T}; the {T} in the
// cost means a summoning-sick Guard can't activate it (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "515c1604-59f0-45b4-91be-d2f20fdd3e1e",
		Name:         "Frostbridge Guard",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}, {T}: Tap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{2}{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
