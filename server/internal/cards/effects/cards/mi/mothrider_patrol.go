package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mothrider Patrol — Creature — Fox Warrior {W}, 1/1 (Neon Dynasty):
//
//	"Flying
//	 {3}{W}, {T}: Tap target creature."
//
// Flying is the printed keyword; the tapper is Arashin Sunshield's body
// with a larger mana cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c4073754-7d92-4976-8e7d-d6eb85b91608",
		Name:            "Mothrider Patrol",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{W}, {T}: Tap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{3}{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
