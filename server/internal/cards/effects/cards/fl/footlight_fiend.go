package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Footlight Fiend — Creature — Devil {B/R}, 1/1 (Ravnica: Clue Edition):
//
//	"When this creature dies, it deals 1 damage to any target."
//
// Dynamite Diver's death trigger (damage by last-known information,
// CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77a5305a-2611-4a64-adb4-da131dcea233",
		Name:         "Footlight Fiend",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisDies("Footlight Fiend — 1 damage to any target", effects.SourceDealsDamageToEachLegalTarget(1)),
				effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 1))),
		},
	})
}
