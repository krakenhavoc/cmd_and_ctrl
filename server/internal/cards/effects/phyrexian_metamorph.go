package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phyrexian Metamorph — "({U/P} can be paid with either {U} or 2
// life.) You may have this creature enter as a copy of any artifact
// or creature on the battlefield, except it's an artifact in
// addition to its other types."
//
// Two things beyond Clone:
//
//   - the candidate set is wider (any ARTIFACT or creature), which
//     is what makes it the format staple: copying an opposing mana
//     rock or an equipment is usually better than copying a body.
//   - the except clause ADDS a card type. Adding rather than
//     replacing matters — a copy of a Wurmcoil Engine is an artifact
//     creature, and a copy of a Sol Ring is just an artifact, so the
//     clause cannot simply set the type line.
//
// Declared simplification: the Phyrexian mana symbol in the cost
// parses (mana_cost.go reads {U/P} and flags HasPhyrexian) but the
// "or 2 life" half of the payment is the cost engine's business,
// not this card's. Paying {U} works today.
func init() {
	Register(Spec{
		OracleID:     "340bbe8b-e987-4c3e-ab4e-9dee63e57d4f",
		Name:         "Phyrexian Metamorph",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Phyrexian mana isn't supported — you must pay the {U}, you can't pay 2 life instead."},
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf(
				"Phyrexian Metamorph",
				func(g *game.Game, _ uuid.UUID, self uuid.UUID) []uuid.UUID {
					return copyCandidates(g, self, func(c game.Card) bool {
						return c.IsCreature() || c.IsArtifact()
					})
				},
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.AddCardType("Artifact")
				},
			),
		},
	})
}
