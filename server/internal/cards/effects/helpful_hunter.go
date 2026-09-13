package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Helpful Hunter — Creature — Cat {1}{W}, 1/1 (EDHREC rank 1958):
//
//	"When this creature enters, draw a card."
//
// The two-mana cantrip Cat. An ETB draw on the stack, so it can be
// responded to and re-fires on a blink, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c0864adb-e9aa-40b6-91ff-a0646193e887",
		Name:         "Helpful Hunter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Helpful Hunter — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
