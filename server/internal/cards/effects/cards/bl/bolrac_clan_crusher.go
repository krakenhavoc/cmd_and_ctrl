package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bolrac-Clan Crusher — Creature — Ogre Warrior {3}{R}{G}, 4/4 (Ravnica Allegiance):
//
//	"{T}, Remove a +1/+1 counter from a creature you control: This creature deals
//	 2 damage to any target."
//
// The counter is removed as a cost, from any creature the controller has one on,
// this creature included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dbfa5409-5706-4de8-8152-e74b8d93910a",
		Name:         "Bolrac-Clan Crusher",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Remove a +1/+1 counter from a creature you control: This creature deals 2 damage to any target.",
			Cost:    effects.Plus(effects.TapCost(), effects.RemoveCountersFrom(game.CounterPlusOne, 1, "a creature you control", effects.Creature())),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect:  effects.DealDamageToTheTarget(2),
		}},
	})
}
