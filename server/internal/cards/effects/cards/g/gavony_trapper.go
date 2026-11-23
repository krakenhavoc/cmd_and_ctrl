package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gavony Trapper — Creature — Human Soldier {W}, 0/2 (Innistrad: Midnight Hunt, S82):
//
//	"{2}, {T}: Tap target creature."
//
// Arashin Sunshield's tapper with a generic cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d4da20c7-ae4a-44c6-8f03-9e719ccee8c9",
		Name:         "Gavony Trapper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Tap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
