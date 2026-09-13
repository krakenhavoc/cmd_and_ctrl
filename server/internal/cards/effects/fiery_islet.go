package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "026f4a4b-eedd-44e1-9d37-ca4fb8d6db98",
		Name:         "Fiery Islet",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:                    ManaAbilityCost{Tap: true, Life: 1},
			Produced:                "{U|R}",
			Label:                   "{T}, Pay 1 life: Add {U} or {R}",
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
