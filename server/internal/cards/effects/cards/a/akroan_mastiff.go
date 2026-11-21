package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Akroan Mastiff — Creature — Dog {3}{W}, 2/2 (Journey into Nyx, #5558):
//
//	"{W}, {T}: Tap target creature."
//
// Nebelgast Beguiler's tapper.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6039b9b7-e49a-4ea3-887c-a17dd502d63f",
		Name:         "Akroan Mastiff",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Tap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
