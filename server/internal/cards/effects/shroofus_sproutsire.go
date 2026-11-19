package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shroofus Sproutsire — Legendary Creature — Saproling {2}{G}, 1/1
// (Foundations Jumpstart, #5929):
//
//	"Trample
//	 Whenever a Saproling you control deals combat damage to a player,
//	 create that many 1/1 green Saproling creature tokens."
//
// One trigger per damage event (Ingenious Infiltrator's shape): two
// Saprolings connecting are two triggers, each making as many tokens as
// the damage that creature actually dealt (CR 120.3), so Shroofus's own
// 1 trample damage makes one. "Saproling" is read through the source's
// text (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e0cebb09-d10d-4b70-96be-82df0e5d52d3",
		Name:            "Shroofus Sproutsire",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b43CreatureOfSubtypeYouControlDealtCombatDamageToPlayer(ev, source, g, "Saproling")
			}, "Shroofus Sproutsire — create that many 1/1 green Saproling creature tokens",
				func(g *game.Game, item *game.StackItem) error {
					n := item.Trigger.Event.Amount
					if n <= 0 {
						return nil
					}
					return CreateToken{Template: TokenCard("1/1 green Saproling"), N: n}.Apply(NewContext(g, item))
				}),
		},
	})
}
