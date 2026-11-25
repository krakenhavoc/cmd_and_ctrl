package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dreadhorde Butcher — Creature — Zombie Warrior {B}{R}, 1/1 (War of the Spark):
//
//	"Haste
//	 Whenever this creature deals combat damage to a player or planeswalker, put
//	 a +1/+1 counter on this creature.
//	 When this creature dies, it deals damage equal to its power to any target."
//
// Two printed triggers, two registered. The counter goes on only while the
// Butcher is still the same permanent. The death trigger reads the Butcher's
// last-known power (CR 603.10a, 608.2h), counters included, and the Butcher is
// the source of the damage as it last existed (CR 609.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d254514b-0688-4f05-9473-3090a2ce7ced",
		Name:            "Dreadhorde Butcher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayerOrPlaneswalker,
				"Dreadhorde Butcher — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			effects.EerieDiesWithPower("Dreadhorde Butcher — it deals damage equal to its power to any target", effects.TargetAny(),
				effects.OldB167SourceDealsItsLastKnownPower),
		},
	})
}
