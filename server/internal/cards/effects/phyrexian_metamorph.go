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
// The Phyrexian mana symbol's "or 2 life" half is the cost engine's
// business, not this card's: #787 landed the announce
// (CastSpellParams.PhyrexianLife, CR 107.4c) and #916 gave the cast
// prompt a stepper for it, so a cast from hand can pay {3} and two
// life.
func init() {
	Register(Spec{
		OracleID:     "340bbe8b-e987-4c3e-ab4e-9dee63e57d4f",
		Name:         "Phyrexian Metamorph",
		Completeness: CompletenessFull,
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
