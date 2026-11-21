package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kami of the Crescent Moon — Legendary Creature — Spirit {U}{U}, 1/3:
//
//	"At the beginning of each player's draw step, that player draws an
//	 additional card."
//
// Dictate of Kruphix's trigger on a body: EACH player's draw step, the
// drawer is the event's Actor, and it resolves after the turn-based draw
// (CR 504.1 does not use the stack).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "38cfe4a0-6ead-435b-9f86-928205c0a52c",
		Name:         "Kami of the Crescent Moon",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginDrawStep},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil
			},
			Key: "Kami of the Crescent Moon — draw an additional card",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{Player: item.Trigger.Event.Actor, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
