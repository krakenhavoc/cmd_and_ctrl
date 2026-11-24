package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blighted Cataract — Land (Commander 2021):
//
//	"{T}: Add {C}.
//	 {5}{U}, {T}, Sacrifice this land: Draw two cards."
//
// A colourless land with a sacrifice-for-cards ability: the sacrifice is a
// cost (CR 602.2b), so the land is gone before the ability resolves and
// the draw does not depend on it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4d722521-0396-48f9-88b2-08b6249f970d",
		Name:         "Blighted Cataract",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{U}, {T}, Sacrifice this land: Draw two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 2},
			Cost:    effects.Plus(effects.ManaCost("{5}{U}"), effects.TapCost(), effects.SacrificeThis()),
			Effect:  effects.Do(effects.DrawCards{N: 2}),
		}},
	})
}
