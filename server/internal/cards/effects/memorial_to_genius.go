package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "81763d7d-3897-4be9-bbf6-f6f5dee366ff",
		Name:         "Memorial to Genius",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{4}{U}, {T}, Sacrifice this land: Draw two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{4}{U}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
			},
		}},
	})
}
