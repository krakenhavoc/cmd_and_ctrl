package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aragorn, Hornburg Hero — Legendary Creature — Human Soldier {1}{R}{G}{W},
// 4/4 (Tales of Middle-earth Commander, #5948):
//
//	"Attacking creatures you control have first strike and renown 1. (When a
//	 creature with renown 1 deals combat damage to a player, if it isn't
//	 renowned, put a +1/+1 counter on it and it becomes renowned.)
//	 Whenever a renowned creature you control deals combat damage to a player,
//	 double the number of +1/+1 counters on it."
//
// The first line is a layer-6 keyword grant to the attacking creatures you
// control (AttackingCreaturesHave, Ohran Frostfang's shape), granting first
// strike and the renown keyword the engine runs (CR 702.112a). The second is
// a trigger per creature that deals combat damage to a player: it fires only
// for a creature that is already renowned when it deals the damage, so the
// creature that becomes renowned from this same damage is not doubled until
// the next time. It doubles the +1/+1 counters on that creature as it
// resolves, by adding as many as it then has (Doubling Season and Hardened
// Scales apply to the added counters); a creature that has left the
// battlefield, or none with counters, does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "07326335-45dd-4ebd-8ae6-5191b6eb2928",
		Name:         "Aragorn, Hornburg Hero",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.AttackingCreaturesHave("Attacking creatures you control have first strike and renown 1.", nil, "first strike", "renown 1"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || ev.Actor != source.Controller {
					return false
				}
				return g.PlayerByIDForEffect(ev.Target) != nil && g.IsRenowned(ev.Source)
			},
			Key:    "Aragorn, Hornburg Hero — double the number of +1/+1 counters on the renowned creature",
			Effect: aragornDouble,
		}},
	})
}

// aragornDouble doubles the +1/+1 counters on the creature that dealt the
// damage, as the trigger resolves.
func aragornDouble(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.Source
	if !effects.OnBattlefield(g, id) {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	n := c.Counters[game.CounterPlusOne]
	if n <= 0 {
		return nil
	}
	return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: n}.Apply(effects.NewContext(g, item))
}
