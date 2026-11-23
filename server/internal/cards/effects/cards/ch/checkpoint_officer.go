package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Checkpoint Officer — Creature — Human Soldier {1}{W}, 1/2 (Ikoria: Lair of Behemoths, #5512):
//
//	"{1}{W}, {T}: Tap target creature."
//
// Celestial Enforcer's tapper without its condition (tapChosenPermanent).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "03e25120-b4e4-409e-8dbc-7574ea4c78d6",
		Name:         "Checkpoint Officer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, {T}: Tap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
