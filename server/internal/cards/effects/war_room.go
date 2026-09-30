package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// War Room — Land:
//
//	"{T}: Add {C}.
//	 {3}, {T}, Pay life equal to the number of colors in your
//	 commanders' color identity: Draw a card."
//
// A colourless rock with a card-draw tax that scales with the deck's
// colours — a mono-colour deck pays 1 life, a four-colour partner pair
// pays 4, a colourless commander pays nothing.
//
// The draw ability shipped caveated in slice 294-c (#1592): the life
// amount is computed, and AbilityCost.Life was a fixed int. #1594 gave
// the cost a registered count (AbilityCost.LifeFrom, ADR 0020 Decision
// 47): LifeEqualToCommanderColors is read once, at announce (CR
// 601.2f–g via CR 602.2b), over every commander the activator owns, and
// the amount is what is charged — a commander that changes afterwards
// changes nothing about the activation already on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "71c52bf5-2a5d-488e-8b15-7ef290e4b77d",
		Name:         "War Room",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{3}, {T}, Pay life equal to the number of colors in your commanders' color identity: Draw a card",
			Cost:  Plus(ManaCost("{3}"), TapCost(), PayLifeCount(LifeEqualToCommanderColors)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
