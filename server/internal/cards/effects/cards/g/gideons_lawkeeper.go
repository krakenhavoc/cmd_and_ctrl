package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gideon's Lawkeeper — Creature — Human Soldier {W}, 1/1 (Magic 2012,
// #5573):
//
//	"{W}, {T}: Tap target creature."
//
// Nebelgast Beguiler's text: {W}, {T}: tap target creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bae86bb0-5a97-4f6a-8a3f-4e6848bfacc0",
		Name:         "Gideon's Lawkeeper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Tap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
