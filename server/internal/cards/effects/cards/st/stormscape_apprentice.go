package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stormscape Apprentice — Creature — Human Wizard {U}, 1/1 (Invasion, #5616):
//
//	"{W}, {T}: Tap target creature.
//	 {B}, {T}: Target player loses 1 life."
//
// Two printed activated abilities, each needing the {T}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf2bacd3-b1ac-4863-b9c7-1e0a8d4854f3",
		Name:         "Stormscape Apprentice",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{W}, {T}: Tap target creature.",
				Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.TapChosenPermanent,
			},
			effects.S82OldB659DrainOne(),
		},
	})
}
