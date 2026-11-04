package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// April, Reporter of the Weird — {2}{U} Legendary Creature — Human Detective
// 2/2 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Whenever April deals combat damage to a player, draw that many cards,
//	 then discard a card."
//
// "That many" is the damage the trigger saw (Zombie Cheese Magician's draw).
// The discard is the draw's continuation, so it is chosen from the hand after
// every card has been drawn (the loot order, drawThenDiscard).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1c371003-e4f0-4d7c-b021-273176772f97",
		Name:         "April, Reporter of the Weird",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("April, Reporter of the Weird — draw that many cards, then discard a card",
				func(g *game.Game, item *game.StackItem) error {
					n := item.Trigger.Event.Amount
					if n <= 0 {
						return nil
					}
					return drawThenDiscard(g, item.Controller, item.SourceCardID, n, 1, "")
				}),
		},
	})
}
