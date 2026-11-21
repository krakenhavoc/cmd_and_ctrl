package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ostiary Thrull — Creature — Thrull {3}{B}, 2/2 (Guildpact, #5599):
//
//	"{W}, {T}: Tap target creature."
//
// Nebelgast Beguiler's tapper: the shared tapChosenPermanent body taps
// the creature named as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c99c9d9-3ea8-43bc-96d7-4dea12770b5c",
		Name:         "Ostiary Thrull",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Tap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
