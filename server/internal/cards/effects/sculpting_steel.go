package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sculpting Steel — Artifact {3}:
//
//	"You may have this artifact enter as a copy of any artifact on the
//	 battlefield."
//
// Copy Artifact's shape with no except clause: an "enters as a copy"
// replacement (EntersAsCopyOf, CR 614.1c and CR 707.9) over the
// artifacts on the battlefield. Not targeting, so a hexproof artifact
// can be copied. With no artifact to copy, or when the player declines,
// it enters as the plain Sculpting Steel.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6c85271c-c711-49b0-a72e-9e576c33714d",
		Name:         "Sculpting Steel",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf(
				"Sculpting Steel",
				func(g *game.Game, _ uuid.UUID, self uuid.UUID) []uuid.UUID {
					return copyCandidates(g, self, func(c game.Card) bool { return c.IsArtifact() })
				},
				nil,
			),
		},
	})
}
