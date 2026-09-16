package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coastal Piracy — Enchantment for {2}{U}{U}:
//
//	"Whenever a creature you control deals combat damage to an
//	opponent, you may draw a card."
//
// S19 sub-PR 7. Same trigger shape as Bident of Thassa with one
// wrinkle: "to an opponent", so combat damage a creature deals to
// its own controller (a redirected or goaded oddity) doesn't count.
func init() {
	Register(Spec{
		OracleID:     "8a05ec32-7b0c-4f23-a4f7-413301c2a70a",
		Name:         "Coastal Piracy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Target != source.Controller && combatDamageToPlayerBy(ev, source.Controller, g)
			}, "Coastal Piracy — draw a card", Do(DrawCards{N: 1})), "Coastal Piracy — draw a card?"),
		},
	})
}
