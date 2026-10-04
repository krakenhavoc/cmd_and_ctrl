package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Copy Artifact — Enchantment {1}{U}:
//
//	"You may have this enchantment enter as a copy of any artifact on
//	 the battlefield, except it's an enchantment in addition to its
//	 other types."
//
// Phyrexian Metamorph's shape with the types swapped: an "enters as a
// copy" replacement (EntersAsCopyOf, CR 614.1c and CR 707.9) over the
// artifacts on the battlefield, whose except clause ADDS a card type.
// Adding, not setting (CR 707.9b): a copy of a Sol Ring is an Artifact
// Enchantment, and a copy of a Wurmcoil Engine is an Artifact Creature
// Enchantment, as the 2004 ruling says. Because the clause says "in
// addition to its other types", a characteristic-defining ability that
// sets the copied artifact's types is still copied (CR 707.9d).
//
// The copy takes the artifact's colours, so it is no longer blue (the
// ruling). Not targeting, so a hexproof or shroud artifact can be
// copied. With no artifact on the battlefield, or when the player
// declines, it enters as a plain enchantment that does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "80bc56a9-40e0-48da-ae86-190e39c8a4a3",
		Name:         "Copy Artifact",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf(
				"Copy Artifact",
				func(g *game.Game, _ uuid.UUID, self uuid.UUID) []uuid.UUID {
					return copyCandidates(g, self, func(c game.Card) bool { return c.IsArtifact() })
				},
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.AddCardType("Enchantment")
				},
			),
		},
	})
}
