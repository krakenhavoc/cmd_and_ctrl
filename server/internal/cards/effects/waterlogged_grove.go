package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "70fa2eba-565e-4fed-adc9-7f5d9fcbf1fa",
		Name:         "Waterlogged Grove",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:                    ManaAbilityCost{Tap: true, Life: 1},
			Produced:                "{G|U}",
			Label:                   "{T}, Pay 1 life: Add {G} or {U}",
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
