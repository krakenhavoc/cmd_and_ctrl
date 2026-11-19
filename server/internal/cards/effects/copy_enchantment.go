package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Copy Enchantment — Enchantment {2}{U} (Enchanting Tales):
//
//	"You may have this enchantment enter as a copy of any enchantment on the
//	 battlefield."
//
// Copy Artifact's shape without an except clause: an "enters as a copy"
// replacement (EntersAsCopyOf, CR 614.1c and CR 707.9) over the enchantments
// on the battlefield, any controller. Not targeting, so a hexproof or shroud
// enchantment can be copied. A copy of an Aura chooses what it enchants as it
// enters (CR 303.4f). With no enchantment to copy, or when the player
// declines, it enters as a plain Copy Enchantment that does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "87bc0b1f-49ef-4c06-9049-2c148b467882",
		Name:         "Copy Enchantment",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf(
				"Copy Enchantment",
				func(g *game.Game, _ uuid.UUID, self uuid.UUID) []uuid.UUID {
					return copyCandidates(g, self, func(c game.Card) bool { return c.IsEnchantment() })
				},
				nil,
			),
		},
	})
}
