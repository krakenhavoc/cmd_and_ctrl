package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dazzling Ramparts — Creature — Wall {4}{W}, 0/7 (Khans of Tarkir):
//
//	"Defender
//	 {1}{W}, {T}: Tap target creature."
//
// Akroan Jailer's activated tap on a Wall with defender; the tap is part of
// the cost (CR 602.2b), so a summoning-sick Wall cannot use it (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0527fc47-2bfc-4d02-90fa-5a1284f89a0a",
		Name:            "Dazzling Ramparts",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, {T}: Tap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
