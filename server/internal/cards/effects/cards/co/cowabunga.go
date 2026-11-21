package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cowabunga! — {G} Sorcery (Teenage Mutant Ninja Turtles, #4457):
//
//	"Look at the top four cards of your library. You may reveal a Mutant,
//	 Ninja, Turtle, or land card from among them and put it into your hand.
//	 Put the rest on the bottom of your library in a random order."
//
// Silundi Vision's look-and-take. A Mutant, Ninja or Turtle card is any card
// with one of those subtypes, read as words of the spell's text; a land card
// is any card with the land type.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1ddf985-1646-468c-b1a4-ab6bb8e9dd7c",
		Name:         "Cowabunga!",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.LookAtTopThenMayTakeToHand(4, effects.Or(effects.AnySubtype("Mutant", "Ninja", "Turtle"), effects.Land()), 1,
				"Cowabunga! — you may reveal a Mutant, Ninja, Turtle, or land card from among them and put it into your hand")(ctx.Game, item)
		},
	})
}
