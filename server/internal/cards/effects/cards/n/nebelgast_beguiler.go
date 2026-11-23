package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nebelgast Beguiler — Creature — Spirit {4}{W}, 2/5 (Innistrad: Crimson
// Vow):
//
//	"{W}, {T}: Tap target creature."
//
// Arashin Sunshield's tap ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b1e76ce9-6f39-419c-99e3-ede2dfa4858c",
		Name:         "Nebelgast Beguiler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Tap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
