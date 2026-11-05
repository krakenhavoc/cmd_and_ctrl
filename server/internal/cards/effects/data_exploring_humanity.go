package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Data, Exploring Humanity — Legendary Artifact Creature — Android
// Scientist {4}{U}, 4/4 (Star Trek Commander, #4974):
//
//	"Seek out — Whenever you draw your second card each turn, discover 4."
//
// "Seek out" is an ability word (CR 207.2c). The trigger is
// b41DrewYourSecondCardThisTurn (once a turn), and discover is the
// engine's keyword action (CR 701.57, DiscoverN): exile until a nonland
// card with mana value 4 or less, cast it free or put it into your hand,
// the rest on the bottom in a random order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "93bfb762-e69e-4d8e-a5b7-e13dd12cd42a",
		Name:         "Data, Exploring Humanity",
		Completeness: CompletenessFull,
		Discovers:    true,
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Data, Exploring Humanity — discover 4", DiscoverN(4)),
		},
	})
}
