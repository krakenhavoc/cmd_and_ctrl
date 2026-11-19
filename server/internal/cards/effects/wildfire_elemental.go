package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "76ff6653-f9b9-4bfe-92a9-52bda275b340",
		Name:         "Wildfire Elemental",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventDealDamage, bnsB42OpponentDealtNoncombatDamage,
				"Wildfire Elemental — creatures you control get +1/+0 until end of turn",
				Do(BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Label: "Wildfire Elemental — +1/+0"}))),
		},
	})
}
