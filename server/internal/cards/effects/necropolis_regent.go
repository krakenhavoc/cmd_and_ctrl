package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Necropolis Regent — Creature — Vampire {3}{B}{B}{B}, 6/5 (EDHREC
// rank 4068):
//
//	"Flying
//	 Whenever a creature you control deals combat damage to a
//	 player, put that many +1/+1 counters on it."
//
// Six mana for a 6/5 flier that makes every creature you control
// permanently bigger the moment it connects. In a wide board the
// Regent turns one unanswered attack step into a board that ends the
// game the next one; against a single blocker it grows itself to a
// 12/11 in one swing. It is the card the +1/+1 counters payoffs — a
// Cathars' Crusade, an All Will Be One, a Shalai and Hallar — want on
// the battlefield.
//
// "THAT MANY" IS THE DAMAGE ACTUALLY DEALT, read off the damage
// event, not the creature's power. The difference is observable and
// it goes both ways: a creature whose damage was partly prevented
// grows by the amount that got through, and a doubled damage event
// (Gratuitous Violence) grows it by the doubled figure. Both are
// what CR 119.3 and the printed "that much" mean.
//
// "ON IT" IS THE CREATURE THAT DEALT THE DAMAGE, not the Regent. The
// damaging creature's ID is closed over as the trigger is built, so
// a creature that traded and died before the trigger resolves simply
// gets no counters (the counter primitive finds nothing) rather than
// the counters landing on the wrong permanent.
//
// ONE TRIGGER PER CREATURE, not one per combat: the printed text is
// singular with no "one or more", so a five-creature alpha strike
// puts five separate triggers on the stack and each grows its own
// creature. The Regent triggers off its own connection too — "a
// creature you control" includes it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "33c3a6be-8cae-4de8-9d9b-43216d727e93",
		Name:            "Necropolis Regent",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			},
			Key: "Necropolis Regent — that many +1/+1 counters",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				grew, amount := ev.Source, ev.Amount
				return game.NewTriggeredItem(source, "Necropolis Regent — that many +1/+1 counters",
					func(g *game.Game, item *game.StackItem) error {
						return AddCounter{
							Target: grew,
							Kind:   game.CounterPlusOne,
							N:      amount,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
