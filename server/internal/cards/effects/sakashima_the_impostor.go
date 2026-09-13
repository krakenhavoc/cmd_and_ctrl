package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sakashima the Impostor — "You may have Sakashima the Impostor
// enter as a copy of any creature on the battlefield, except its
// name is Sakashima the Impostor, it's legendary in addition to its
// other types, and it has '{2}{U}{U}: Return Sakashima the Impostor
// to its owner's hand at the beginning of the next end step.'"
//
// The name exception is the one that proves the model. A copy takes
// the copied card's NAME along with everything else — that is what
// makes a Clone of your own commander die to the CR 704.5j legend
// rule — and Sakashima is printed the way it is precisely to dodge
// that. Keeping the name means the copy is a different legendary
// permanent from the thing it copied, so both survive; the added
// supertype is what puts it under the legend rule against a SECOND
// Sakashima.
//
// Declared simplification: the granted "{2}{U}{U}: return it at the
// beginning of the next end step" activated ability is not carried.
// A copy effect replaces the card-carried ability slices with the
// copied card's, and the catalog's own activated abilities key on
// oracle ID, which after the copy is the copied card's — so the
// granted ability has nowhere to live until an "except" clause can
// APPEND an ability shape rather than edit printed values. Nothing
// else about the card is affected, and the missing ability is only
// ever a bonus escape hatch for its controller.
func init() {
	Register(Spec{
		OracleID: "a7243d25-22a2-4df5-adaf-1f40f5330ec1",
		Name:     "Sakashima the Impostor",
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf(
				"Sakashima the Impostor",
				func(g *game.Game, _ uuid.UUID, self uuid.UUID) []uuid.UUID {
					return copyCandidates(g, self, func(c game.Card) bool {
						return c.IsCreature()
					})
				},
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.SetName("Sakashima the Impostor")
					v.AddSupertype("Legendary")
				},
			),
		},
	})
}
