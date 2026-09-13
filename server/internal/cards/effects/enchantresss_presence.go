package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Enchantress's Presence — Enchantment, {2}{G} (EDHREC rank 985):
//
//	"Whenever you cast an enchantment spell, draw a card."
//
// The original enchantress, on an enchantment rather than a body so
// it dodges creature removal. Mesa Enchantress's trigger without the
// "may" — mandatory, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "795b096a-2bce-4588-a2c9-abc5ea40dc0c",
		Name:         "Enchantress's Presence",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventCast},
			AppliesTo: b08EnchantmentSpellCastByYou,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Enchantress's Presence — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
