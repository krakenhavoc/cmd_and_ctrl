package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Waterlogged Grove — Land (EDHREC rank 1142):
//
//	"{T}, Pay 1 life: Add {G} or {U}.
//	 {1}, {T}, Sacrifice this land: Draw a card."
//
// The Simic member of the Horizon Canopy cycle, Fiery Islet's shape
// with the colours swapped: the life is a COST (validated before the
// land taps, refused at zero life), the two printed colours are not
// narrowed to the commander's identity, and the cash-in is Buried
// Ruin's three-component cost with a draw. See fiery_islet.go.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "70fa2eba-565e-4fed-adc9-7f5d9fcbf1fa",
		Name:         "Waterlogged Grove",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Life: 1},
			Produced: "{G|U}",
			Label:    "{T}, Pay 1 life: Add {G} or {U}",
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
