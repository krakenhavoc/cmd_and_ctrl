package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Copy Land — Enchantment {2}{U} (Modern Horizons 3 Commander, #5948):
//
//	"You may have this enchantment enter as a copy of any land on the
//	 battlefield, except it's an enchantment in addition to its other
//	 types."
//
// Copy Artifact's shape over lands (EntersAsCopyOf, CR 614.1c and CR 707.9):
// the except clause ADDS the Enchantment type (CR 707.9b), so a copy of a
// Forest is an Enchantment Land — Forest, which taps for {G}, and a copy of
// a creature land is an Enchantment Land Creature. Because it says "in
// addition to its other types", a characteristic-defining ability that sets
// the land's types is still copied (CR 707.9d). The copy takes the land's
// colours (a land has none), so it is no longer blue. Not targeting, so a
// hexproof land can be copied. With no land on the battlefield, or when the
// player declines, it enters as a plain enchantment that does nothing. The
// copy does not enter tapped unless the copied land says so.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "caa208bd-92e9-40ec-95c0-86e89d4facbb",
		Name:         "Copy Land",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf(
				"Copy Land",
				func(g *game.Game, _ uuid.UUID, self uuid.UUID) []uuid.UUID {
					return effects.CopyCandidates(g, self, func(c game.Card) bool { return c.IsLand() })
				},
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.AddCardType("Enchantment")
				},
			),
		},
	})
}
