package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wall of Blossoms — Creature — Plant Wall {1}{G}, 0/4 (EDHREC rank
// 2218):
//
//	"Defender
//	 When this creature enters, draw a card."
//
// The cantrip wall. Defender rides PrintedKeywords; the ETB is a
// mandatory draw on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ef4d5fb3-70a3-433d-a9d3-18b2beb8d79f",
		Name:            "Wall of Blossoms",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Wall of Blossoms — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
