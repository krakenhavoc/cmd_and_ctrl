package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mudbutton Torchrunner — Creature — Goblin Warrior {2}{R}, 1/1 (Jumpstart
// 2022, #5948):
//
//	"When this creature dies, it deals 3 damage to any target."
//
// Pitchburn Devils' dies trigger: the damage comes from the Torchrunner as
// it last existed on the battlefield (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "045b2b84-96bd-46fb-8aae-a33968e741c7",
		Name:         "Mudbutton Torchrunner",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisDies("Mudbutton Torchrunner — 3 damage to any target", effects.SourceDealsDamageToEachLegalTarget(3)),
				effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 3))),
		},
	})
}
