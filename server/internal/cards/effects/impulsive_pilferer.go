package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Impulsive Pilferer — 1/1 Creature — Goblin Pirate for {R}:
//
//	"When this creature dies, create a Treasure token."
//
// A one-drop that turns into mana when it trades or gets sacrificed
// — the ramp half of a deck that wants artifacts entering.
//
// Encore {3}{R} is an alternative cast path from the graveyard
// (S29) and isn't modelled.
func init() {
	Register(Spec{
		OracleID: "7d9fc9e7-d80b-49c3-871c-ed25b3059ae8",
		Name:     "Impulsive Pilferer",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Impulsive Pilferer — create a Treasure",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{
							Controller: item.Controller,
							Template:   TreasureToken(),
							N:          1,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
