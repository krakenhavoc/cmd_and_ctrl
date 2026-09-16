package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kher Keep — Legendary Land (EDHREC rank 1704):
//
//	"{T}: Add {C}.
//	 {1}{R}, {T}: Create a 0/1 red Kobold creature token named
//	 Kobolds of Kher Keep."
//
// A colorless land that makes a chump blocker — or, in the deck
// that plays it, a sacrifice-fodder body every turn. The mana half
// is a plain {C}; the token half is a CR 602 activation with a mana
// and a tap component, so it uses the stack. The token's printed
// name matters (Rohgahh of Kher Keep reads it), so the template
// carries it. Legendary, so the legend rule applies to a second copy.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "79638767-fbc7-451a-b29f-d93f2ac6f102",
		Name:         "Kher Keep",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{1}{R}, {T}: Create a 0/1 red Kobold creature token named Kobolds of Kher Keep.",
			Cost:  Plus(ManaCost("{1}{R}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: TokenCard("0/1 red Kobolds of Kher Keep"), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
