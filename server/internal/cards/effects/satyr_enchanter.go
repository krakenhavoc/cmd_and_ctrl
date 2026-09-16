package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Satyr Enchanter — Creature — Satyr Druid {1}{G}{W}, 2/2 (EDHREC
// rank 2246):
//
//	"Whenever you cast an enchantment spell, draw a card."
//
// The enchantress that does not ask. Mesa Enchantress's cast trigger
// without the "you may": the spell's type is read off the stack
// (enchantmentSpellCastByYou), and the draw resolves before the
// enchantment does (LIFO), as in paper.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aa321138-b1a7-4b8e-a2ca-b9ce65704e92",
		Name:         "Satyr Enchanter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, enchantmentSpellCastByYou, "Satyr Enchanter — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
