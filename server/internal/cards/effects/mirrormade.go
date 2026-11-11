package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mirrormade — Enchantment {1}{U}{U} (Duskmourn Commander, #5368):
//
//	"You may have this enchantment enter as a copy of any artifact or
//	 enchantment on the battlefield."
//
// Copy Artifact's shape without an except clause: an "enters as a copy"
// replacement (EntersAsCopyOf, CR 614.1c, CR 707.2) over every artifact and
// every enchantment on the battlefield, whoever controls it. Not targeting,
// so hexproof and shroud do not stop it (CR 115.10a). Declined, or with
// nothing to copy, it enters as a plain Mirrormade.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "79a71fe5-38ce-4bb5-aea0-c9b9a856d397",
		Name:         "Mirrormade",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf(
				"Mirrormade",
				func(g *game.Game, _ uuid.UUID, self uuid.UUID) []uuid.UUID {
					return copyCandidates(g, self, func(c game.Card) bool { return c.IsArtifact() || c.IsEnchantment() })
				},
				nil,
			),
		},
	})
}
