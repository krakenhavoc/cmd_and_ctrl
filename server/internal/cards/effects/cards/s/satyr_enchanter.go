package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "aa321138-b1a7-4b8e-a2ca-b9ce65704e92",
		Name:         "Satyr Enchanter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.EnchantmentSpellCastByYou, "Satyr Enchanter — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
