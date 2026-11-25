package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gideon's Company — Creature — Human Soldier {3}{W}, 3/3 (War of the Spark):
//
//	"Whenever you gain life, put two +1/+1 counters on this creature.
//	 {3}{W}: Put a loyalty counter on target Gideon planeswalker."
//
// The trigger fires once per life-gain event of its controller (CR 119.9), so
// two events are two triggers; the counters go on only while it is still the same
// permanent. The activated ability targets a planeswalker with the subtype Gideon
// (CR 205.3j), any player's, and adds one loyalty counter to it, which is not a
// loyalty ability and so can be used in any timing the ability allows (it is an
// ordinary activated ability).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "083dcc19-3357-48e4-94e8-8da5bbfba210",
		Name:         "Gideon's Company",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouGainLife("Gideon's Company — put two +1/+1 counters on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{W}: Put a loyalty counter on target Gideon planeswalker.",
			Cost:    effects.ManaCost("{3}{W}"),
			Targets: effects.TargetPermanent("target Gideon planeswalker", effects.Planeswalker(), effects.OfSubtype("Gideon")),
			Effect:  effects.OldB167LoyaltyCounterOnTheTarget,
		}},
	})
}
