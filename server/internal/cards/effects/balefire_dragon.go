package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Balefire Dragon — Creature — Dragon {5}{R}{R}, 6/6 (EDHREC rank
// 1266):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, it deals
//	 that much damage to each creature that player controls."
//
// A one-sided wipe on connection. The trigger is the Professional
// Face-Breaker shape narrowed to the Dragon itself, capturing the
// damaged player and the amount by value in Build; the rider deals
// the Dragon's damage to each creature that player controls, from a
// set snapshotted before the first point lands (CR 608.2). The
// Dragon's own damage goes through the ordinary damage path, so
// prevention and indestructible apply as they would to any source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3d783beb-9ca6-4681-9276-fc3ad13b993f",
		Name:            "Balefire Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Source == source.InstanceID && combatDamageToPlayerBy(ev, source.Controller, g)
			},
			Key: "Balefire Dragon — that much damage to each creature that player controls",
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				return b11DamageEachCreatureControlledBy(g, item, item.Trigger.Event.Target, item.Trigger.Event.Amount)
			},
		}},
	})
}
