package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunbaked Canyon — Land (EDHREC rank 1484):
//
//	"{T}, Pay 1 life: Add {R} or {W}.
//	 {1}, {T}, Sacrifice this land: Draw a card."
//
// The Boros member of the Horizon Canopy cycle, Fiery Islet's shape
// with the colours swapped: the life is a COST (validated before the
// land taps, refused at zero life), the two printed colours are not
// narrowed to the commander's identity, and the cash-in is Buried
// Ruin's three-component cost with a draw. See fiery_islet.go.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f97fd068-b83a-4621-bf8c-cc96e880ce90",
		Name:         "Sunbaked Canyon",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Life: 1},
			Produced: "{R|W}",
			Label:    "{T}, Pay 1 life: Add {R} or {W}",
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
