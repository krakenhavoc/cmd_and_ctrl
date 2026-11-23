package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wildfire Elemental — Creature — Elemental {2}{R}{R}, 3/3:
//
//	"Whenever an opponent is dealt noncombat damage, creatures you
//	 control get +1/+0 until end of turn."
//
// One trigger per instance of damage to an opponent (Master of Barbs'
// shape): combat damage and damage to a creature or planeswalker are not
// it. "Creatures you control" is the set when the trigger resolves.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76ff6653-f9b9-4bfe-92a9-52bda275b340",
		Name:         "Wildfire Elemental",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventDealDamage, effects.BnsB42OpponentDealtNoncombatDamage,
				"Wildfire Elemental — creatures you control get +1/+0 until end of turn",
				effects.Do(effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Label: "Wildfire Elemental — +1/+0"}))),
		},
	})
}
