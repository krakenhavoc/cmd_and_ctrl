package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spirited Companion — Enchantment Creature — Dog, {1}{W}, 1/1
// (EDHREC rank 913):
//
//	"When this creature enters, draw a card."
//
// A cantrip on a body that is also an enchantment, which is why an
// enchantress deck plays it: it draws once on cast (Sythis, Mesa
// Enchantress) and once on entry, and it turns on constellation.
// Mulldrifter's trigger with a one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9c5f0d91-9d86-4e66-94fd-4af93ad01838",
		Name:         "Spirited Companion",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Spirited Companion — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
