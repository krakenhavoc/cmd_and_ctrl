package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goldmeadow Harrier — Creature — Kithkin Soldier {W}, 1/1 (Duel Decks:
// Elspeth vs. Tezzeret, #5578):
//
//	"{W}, {T}: Tap target creature."
//
// The {T} in the cost needs the Harrier to have been under its controller's
// control since the start of the turn (CR 302.6), which the engine enforces
// for every tap-symbol ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "546f037b-aea8-4c7a-9b66-373f1be0bac6",
		Name:         "Goldmeadow Harrier",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Tap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapTheFirstLegalTarget,
		}},
	})
}
