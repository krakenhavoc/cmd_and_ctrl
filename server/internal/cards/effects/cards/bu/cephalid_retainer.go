package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cephalid Retainer — Creature — Octopus {2}{U}{U}, 2/3 (Odyssey, #5613):
//
//	"{U}{U}: Tap target creature without flying."
//
// Arashin Sunshield's tapper with no tap symbol, so it can be used again
// and again, and on the turn the Retainer arrives. A creature that gains
// flying in response is no longer a legal target (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9a0a926c-59ab-4bd3-a238-13988bdd96a5",
		Name:         "Cephalid Retainer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}{U}: Tap target creature without flying.",
			Cost:    effects.ManaCost("{U}{U}"),
			Targets: effects.TargetCreature("target creature without flying", effects.Not(effects.HasKeyword("flying"))),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
