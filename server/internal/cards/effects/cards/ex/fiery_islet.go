package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fiery Islet — Land (EDHREC rank 857):
//
//	"{T}, Pay 1 life: Add {U} or {R}.
//	 {1}, {T}, Sacrifice this land: Draw a card."
//
// The Horizon Canopy cycle: a painless-looking dual that costs a
// life per tap and cashes in for a card when the mana stops
// mattering. The life is a COST (Mana Confluence's shape — validated
// before the land taps, refused at zero life), not a painland's
// rider; the two printed colours are not narrowed to the commander's
// identity. The cash-in is Buried Ruin's three-component cost with
// a draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "026f4a4b-eedd-44e1-9d37-ca4fb8d6db98",
		Name:         "Fiery Islet",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Life: 1},
			Produced: "{U|R}",
			Label:    "{T}, Pay 1 life: Add {U} or {R}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice this land: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
