package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Satyr Enchanter — Creature — Satyr Druid {1}{G}{W}, 2/2 (EDHREC
// rank 2246):
//
//	"Whenever you cast an enchantment spell, draw a card."
//
// The enchantress that does not ask. Mesa Enchantress's cast trigger
// without the "you may": the spell's type is read off the stack
// (b08EnchantmentSpellCastByYou), and the draw resolves before the
// enchantment does (LIFO), as in paper.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aa321138-b1a7-4b8e-a2ca-b9ce65704e92",
		Name:         "Satyr Enchanter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventCast},
			AppliesTo: b08EnchantmentSpellCastByYou,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Satyr Enchanter — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
