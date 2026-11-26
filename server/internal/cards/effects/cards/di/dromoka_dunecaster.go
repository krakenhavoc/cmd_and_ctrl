package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dromoka Dunecaster — Creature — Human Wizard {W}, 0/2 (Dragons of Tarkir):
//
//	"{1}{W}, {T}: Tap target creature without flying."
//
// Flying is read through the layers as the target is chosen and again as the
// ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ac44e0a-782f-49ee-b263-2fb65359faaa",
		Name:         "Dromoka Dunecaster",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, {T}: Tap target creature without flying.",
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature without flying", effects.Not(effects.HasKeyword("flying"))),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
