package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Memorial to Genius — Land:
//
//	"This land enters tapped.
//	 {T}: Add {U}.
//	 {4}{U}, {T}, Sacrifice this land: Draw two cards."
//
// The blue Memorial: a tapped Island that cashes in for two cards late.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "81763d7d-3897-4be9-bbf6-f6f5dee366ff",
		Name:         "Memorial to Genius",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{U}, {T}, Sacrifice this land: Draw two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{4}{U}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{Player: item.Controller, N: 2}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
