package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Roaring Furnace // Steaming Sauna — Enchantment — Room (CR 709.5):
//
//	Roaring Furnace {1}{R}: "When you unlock this door, this Room deals
//	damage equal to the number of cards in your hand to target creature an
//	opponent controls."
//	Steaming Sauna {3}{U}{U}: "You have no maximum hand size." and "At the
//	beginning of your end step, draw a card."
func init() {
	Register(Room(RoomSpec{
		OracleID:     "d5f31713-d380-42ba-8052-4b8d9beb3958",
		Name:         "Roaring Furnace // Steaming Sauna",
		Completeness: CompletenessFull,
		Left: Door{Triggered: []game.TriggeredAbility{roomDamageToOpponentCreature(game.DoorLeft,
			"Roaring Furnace — damage equal to cards in your hand to target creature an opponent controls",
			func(g *game.Game, item *game.StackItem) int { return b14HandSize(g, item.Controller) })}},
		Right: Door{
			NoMaxHandSize: true,
			Triggered: []game.TriggeredAbility{
				AtYourEndStep("Steaming Sauna — draw a card", Do(DrawCards{N: 1})),
			},
		},
	}))
}
