package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sparkmage Apprentice — Creature — Human Wizard {1}{R}, 1/1 (Ravnica: City
// of Guilds, #5600):
//
//	"When this creature enters, it deals 1 damage to any target."
//
// Skeleton Archer's targeted enters trigger: the creature is the source of
// the 1 damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "69003cbc-8e5a-41dc-ac53-18f6e776217b",
		Name:         "Sparkmage Apprentice",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Sparkmage Apprentice — 1 damage to any target",
				effects.SourceDealsDamageToEachLegalTarget(1)), effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 1))),
		},
	})
}
