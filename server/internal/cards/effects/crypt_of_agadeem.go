package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crypt of Agadeem — Land:
//
//	"This land enters tapped.
//	 {T}: Add {B}.
//	 {2}, {T}: Add {B} for each black creature card in your graveyard."
//
// Cabal Coffers's counted mana ability, over a graveyard instead of the
// battlefield. The count is made after the {2} is paid and the land is
// tapped (CR 605.3b), and reads each card's colour as it sits in the
// graveyard (CR 109.3): the card's own printed colours, so a multicolour
// creature that includes black counts. With none it adds nothing and still
// taps. The enters-tapped clause is the CR 614 self-replacement.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4fe8af73-c84a-44bd-9739-ee5c8b027874",
		Name:         "Crypt of Agadeem",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{B}",
				Label:    "Add {B}",
			},
			{
				Cost: ManaAbilityCost{Tap: true, Mana: "{2}"},
				ProducedAmount: func(g *game.Game, controller, source uuid.UUID) game.ManaAmount {
					black := "B"
					if src, ok := g.LookupCardForEffect(source); ok {
						black = src.TextColor("B")
					}
					n := len(graveyardIDs(g, controller, func(c game.Card) bool {
						return c.IsCreature() && c.HasColor(black)
					}))
					return game.ManaAmount{Options: []string{"B"}, N: n}
				},
				Label: "{2}, {T}: Add {B} for each black creature card in your graveyard",
			},
		},
	})
}
