package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Akoum Boulderfoot — Creature — Giant Warrior {4}{R}{R}, 4/5 (Rise of the
// Eldrazi, S82 OLD batch 483, #5581):
//
//	"When this creature enters, it deals 1 damage to any target."
//
// Blisterstick Shaman's trigger: the damage is dealt by the Boulderfoot.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c16367eb-1c0a-4197-8e2f-2d74e4381024",
		Name:         "Akoum Boulderfoot",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Akoum Boulderfoot — 1 damage to any target",
				effects.SourceDealsDamageToEachLegalTarget(1)), effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 1))),
		},
	})
}
