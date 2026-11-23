package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Auriok Transfixer — Creature — Human Scout {W}, 1/1 (Mirrodin, #5606):
//
//	"{W}, {T}: Tap target artifact."
//
// Arashin Sunshield's tapper on an artifact clause: any artifact, an
// artifact creature or an opponent's included, and Staff of Domination's
// body (tapChosenPermanent).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0ad5f4c5-f6d2-4e96-af65-5488d9711113",
		Name:         "Auriok Transfixer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Tap target artifact.",
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target artifact", effects.Artifact()),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
