package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragon Mage — Creature — Dragon Wizard {5}{R}{R}, 5/5 (EDHREC rank
// 2944):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, each
//	 player discards their hand, then draws seven cards."
//
// Wheel of Fortune on a Dragon. Flying rides PrintedKeywords; the
// wheel is one trigger per player the Mage connects with
// (combatDamageToPlayerBy narrowed to the Mage as the dealer), and
// its body is Wheel of Fortune's b10EachPlayerWheels — every
// discard before any draw, so discard payoffs queue while the
// effect is still resolving.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "eab71e4e-27c3-4c41-b95f-259c1d14b97a",
		Name:            "Dragon Mage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Source == source.InstanceID && combatDamageToPlayerBy(ev, source.Controller, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Dragon Mage — each player discards their hand, then draws seven cards",
					b10EachPlayerWheels)
			},
		}},
	})
}
