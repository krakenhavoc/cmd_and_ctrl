package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jace's Projection — Creature — Wizard Illusion {2}{U}{U}, 2/2 (War of the
// Spark):
//
//	"Whenever you draw a card, put a +1/+1 counter on this creature.
//	 {3}{U}: Put a loyalty counter on target Jace planeswalker."
//
// One trigger per card drawn (CR 121.2), counters only while it is still the same
// permanent; the activated ability is Gideon's Company's, over a planeswalker
// with the subtype Jace.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff8db8aa-1d5a-42f8-8056-8da0da9846bc",
		Name:         "Jace's Projection",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouDraw("Jace's Projection — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}: Put a loyalty counter on target Jace planeswalker.",
			Cost:    effects.ManaCost("{3}{U}"),
			Targets: effects.TargetPermanent("target Jace planeswalker", effects.Planeswalker(), effects.OfSubtype("Jace")),
			Effect:  effects.OldB167LoyaltyCounterOnTheTarget,
		}},
	})
}
