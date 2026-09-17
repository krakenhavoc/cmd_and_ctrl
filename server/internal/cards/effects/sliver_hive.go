package effects

import (
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
	Register(Spec{
		OracleID:     "e7286688-ffbe-4d25-ad55-27990f005368",
		Name:         "Sliver Hive",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:                    ManaAbilityCost{Tap: true},
				Produced:                "{W|U|B|R|G}",
				Label:                   "Add one mana of any color (Sliver spells only)",
				IgnoreCommanderIdentity: true,
				Restrictions: []string{
					ManaRestrictCast,
					ManaRestrictSubtype("Sliver"),
				},
			},
		},
		Activated: []ActivatedAbility{{
			Label: "{5}, {T}: Create a 1/1 colorless Sliver creature token. Activate only if you control a Sliver.",
			Cost:  Plus(ManaCost("{5}"), TapCost()),
			Condition: ControlsAtLeast(1, func(c game.Card) bool {
				return c.HasSubtype("Sliver")
			}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 colorless Sliver"), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
