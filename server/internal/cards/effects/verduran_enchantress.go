package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Verduran Enchantress — Creature — Human Druid {1}{G}{G}, 0/2
// (EDHREC rank 2317):
//
//	"Whenever you cast an enchantment spell, you may draw a card."
//
// The original enchantress. Mesa Enchantress's trigger in green: the
// spell's type is read off the stack (b08EnchantmentSpellCastByYou),
// "you may" is the prompt, and the draw resolves before the
// enchantment does (LIFO), as in paper.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cd98a31b-cc7e-43f9-982e-109ad9850908",
		Name:         "Verduran Enchantress",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventCast, b08EnchantmentSpellCastByYou, "Verduran Enchantress — draw a card", Do(DrawCards{N: 1})), "Verduran Enchantress — draw a card?"),
		},
	})
}
