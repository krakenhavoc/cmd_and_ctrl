package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fire Imp — Creature — Imp {2}{R}, 2/1 (Portal, #5630):
//
//	"When this creature enters, it deals 2 damage to target creature."
//
// Flametongue Kavu's enters trigger, with 2 damage. With no creature
// to target, the trigger is removed from the stack (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5bd806e7-3f9b-4bb4-9708-f87a578f531e",
		Name:         "Fire Imp",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Fire Imp — 2 damage to target creature", effects.SourceDealsDamageToEachLegalTarget(2)),
				effects.TargetCreature("target creature")), effects.ForTargets(effects.DamageToTarget(0, 2))),
		},
	})
}
