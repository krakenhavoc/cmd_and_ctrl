package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Horizon of Progress — Land (Modern Horizons 3 Commander, #5468):
//
//	"{T}, Pay 1 life: Add one mana of any type that a land you control
//	 could produce.
//	 {3}, {T}: You may put a land card from your hand onto the
//	 battlefield tapped.
//	 {1}, {T}, Sacrifice this land: Draw a card."
//
// The mana ability is Reflecting Pool's derivation with a life payment
// in the cost: any TYPE, so {C} counts, read from the lands you control
// at activation through the engine's recursion guard, so a lone Horizon
// makes nothing (it can produce no type of its own). The life is a cost
// (CR 119.4), so the ability cannot be activated at a life total below 1.
//
// The land drop is Jefferies Tube's optional pick: any land card in
// hand, entering tapped, and declining is a real answer. It is not a
// land play, so it uses none of the turn's land drops. The draw row is
// Horizon Canopy's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "59a82f57-fe2f-4834-a4ee-4b948eef1e12",
		Name:         "Horizon of Progress",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:                    ManaAbilityCost{Tap: true, Life: 1},
			DerivedMatch:            DerivedFromOwnLands(),
			DerivedColorsOnly:       false,
			DerivesFromOtherSources: true,
			Label:                   "{T}, Pay 1 life: Add one mana of any type that a land you control could produce",
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "{3}, {T}: You may put a land card from your hand onto the battlefield tapped.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{3}"), TapCost()),
				Effect: Do(PutFromHandOntoBattlefield{
					Match:    Land(),
					Tapped:   true,
					Optional: true,
					Label:    "Horizon of Progress — you may put a land card from your hand onto the battlefield tapped",
				}),
			},
			{
				Label:   "{1}, {T}, Sacrifice this land: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{1}"), TapCost(), SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
