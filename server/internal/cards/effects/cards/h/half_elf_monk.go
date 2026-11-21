package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Half-Elf Monk — Creature — Human Elf Monk {3}{W}, 1/4 (Adventures in
// the Forgotten Realms):
//
//	"Vigilance
//	 Stunning Strike — {1}{W}, {T}: Tap target creature."
//
// Arashin Sunshield's tapper. "Stunning Strike" has no rules meaning
// (CR 207.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "00d11662-9d10-4915-b80c-c1ad54b4e71b",
		Name:            "Half-Elf Monk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Stunning Strike — {1}{W}, {T}: Tap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
