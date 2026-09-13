package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nurturing Peatland — Land (EDHREC rank 1599):
//
//	"{T}, Pay 1 life: Add {B} or {G}.
//	 {1}, {T}, Sacrifice this land: Draw a card."
//
// The Golgari member of the Horizon Canopy cycle, Fiery Islet's shape
// with the colours swapped: the life is a COST (validated before the
// land taps, refused at zero life), the two printed colours are not
// narrowed to the commander's identity, and the cash-in is Buried
// Ruin's three-component cost with a draw. See fiery_islet.go.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8ed932ff-986c-4592-ad70-53b3fac80d69",
		Name:         "Nurturing Peatland",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:                    ManaAbilityCost{Tap: true, Life: 1},
			Produced:                "{B|G}",
			Label:                   "{T}, Pay 1 life: Add {B} or {G}",
			IgnoreCommanderIdentity: true,
		}},
		Activated: []ActivatedAbility{{
			Label: "{1}, {T}, Sacrifice this land: Draw a card.",
			Cost:  Plus(ManaCost("{1}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
