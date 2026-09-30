package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Clever Impersonator — Creature — Shapeshifter {2}{U}{U}, 0/0
// (slice 296-m):
//
//	"You may have this creature enter as a copy of any nonland
//	 permanent on the battlefield."
//
// Clone's own construction (EntersAsCopyOf) with the candidate set
// widened from creatures to every nonland permanent — an artifact, an
// enchantment, a planeswalker, another player's, anything but a land.
// No "except" clause, same as Clone.
//
// "Any nonland permanent on the battlefield" is not targeting: the
// choice is made as the permanent enters (CR 707.2), so hexproof,
// shroud and protection do not stop it. Declining, or finding nothing
// to copy, leaves Clever Impersonator as its printed 0/0, and it then
// dies to the toughness state-based action, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d4ec0df9-a4a3-48cf-a0aa-b1aea5c49142",
		Name:         "Clever Impersonator",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf(
				"Clever Impersonator",
				func(g *game.Game, _ uuid.UUID, self uuid.UUID) []uuid.UUID {
					return copyCandidates(g, self, func(c game.Card) bool { return !c.IsLand() })
				},
				nil,
			),
		},
	})
}
