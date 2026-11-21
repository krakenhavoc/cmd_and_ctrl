package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roaring Furnace // Steaming Sauna — Enchantment — Room (CR 709.5):
//
//	Roaring Furnace {1}{R}: "When you unlock this door, this Room deals
//	damage equal to the number of cards in your hand to target creature an
//	opponent controls."
//	Steaming Sauna {3}{U}{U}: "You have no maximum hand size." and "At the
//	beginning of your end step, draw a card."
func init() {
	effects.Register(effects.Room(effects.RoomSpec{
		OracleID:     "d5f31713-d380-42ba-8052-4b8d9beb3958",
		Name:         "Roaring Furnace // Steaming Sauna",
		Completeness: effects.CompletenessFull,
		Left: effects.Door{Triggered: []game.TriggeredAbility{effects.RoomDamageToOpponentCreature(game.DoorLeft,
			"Roaring Furnace — damage equal to cards in your hand to target creature an opponent controls",
			func(g *game.Game, item *game.StackItem) int { return effects.B14HandSize(g, item.Controller) })}},
		Right: effects.Door{
			NoMaxHandSize: true,
			Triggered: []game.TriggeredAbility{
				effects.AtYourEndStep("Steaming Sauna — draw a card", effects.Do(effects.DrawCards{N: 1})),
			},
		},
	}))
}
