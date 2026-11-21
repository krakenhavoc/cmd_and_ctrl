package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Urza's Factory — Land — Urza's (EDHREC rank 4474):
//
//	"{T}: Add {C}.
//	 {7}, {T}: Create a 2/2 colorless Assembly-Worker artifact
//	 creature token."
//
// A colourless land that turns into a creature factory once the game
// goes long. Seven mana for a 2/2 is a terrible rate and that is not
// the point: it is a MANA SINK on a land, so a colourless or
// artifact deck's flooded draws stop being dead, and it dodges
// every sweeper that only sees creatures.
//
// Two abilities that look alike and are not. The {T}: Add {C} is a
// CR 605 mana ability — it does not use the stack and cannot be
// responded to. The {7}, {T} token-maker is an ordinary activated
// ability that goes on the stack, so the table gets a window. Both
// tap the land, so only one is available per untap.
//
// The subtype on the type line is "Urza's", which is the same
// subtype the Tron lands carry — nothing in this file cares, but it
// means a Factory really does count as an Urza's land for anything
// that eventually looks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a620d20-f14e-43d0-8e57-c2a197e2ec51",
		Name:         "Urza's Factory",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{7}, {T}: Create a 2/2 colorless Assembly-Worker artifact creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{7}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.CreateToken{
					Controller: item.Controller,
					Template:   effects.TokenCard("2/2 colorless Assembly-Worker artifact"),
					N:          1,
				}.Apply(ctx)
			},
		}},
	})
}
