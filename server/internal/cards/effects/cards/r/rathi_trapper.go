package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rathi Trapper — Creature — Human Rebel Rogue {1}{B}, 1/2 (Planar Chaos, #5595):
//
//	"{B}, {T}: Tap target creature."
//
// Aysen Bureaucrats' tapper with a mana cost and no power limit. The {T}
// keeps it off the turn it arrives (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0920d6aa-eab7-4116-b2f8-b13080d5d7e3",
		Name:         "Rathi Trapper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, {T}: Tap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapTheFirstLegalTarget,
		}},
	})
}
