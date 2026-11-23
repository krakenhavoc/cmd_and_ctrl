package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sliver Hive — Land (EDHREC rank 5018):
//
//	"{T}: Add {C}.
//	 {T}: Add one mana of any color. Spend this mana only to cast a
//	 Sliver spell.
//	 {5}, {T}: Create a 1/1 colorless Sliver creature token. Activate
//	 only if you control a Sliver."
//
// The Sliver deck's rainbow land. The coloured mana carries the
// Eldrazi Temple / Delighted Halfling restriction tags (cast only, a
// Sliver spell), and the auto-tapper leaves a restricted ability
// alone, so it is a deliberate click. "Activate only if you control a
// Sliver" is the token ability's activation condition (CR 602.1b,
// #743): any permanent with the Sliver subtype, effective, so a
// changeling counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e7286688-ffbe-4d25-ad55-27990f005368",
		Name:         "Sliver Hive",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color (Sliver spells only)",
				Restrictions: []string{
					effects.ManaRestrictCast,
					effects.ManaRestrictSubtype("Sliver"),
				},
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}, {T}: Create a 1/1 colorless Sliver creature token. Activate only if you control a Sliver.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost()),
			// ControlsAtLeast lends the source's text (CR 612.2, #3984).
			Condition: effects.ControlsAtLeast(1, func(c game.Card) bool {
				return c.HasSubtypeWord("Sliver")
			}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 colorless Sliver"), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
