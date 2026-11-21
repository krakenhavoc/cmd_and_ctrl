package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Commando — Creature — Goblin {4}{R}, 2/2 (Starter 1999, #5620):
//
//	"When this creature enters, it deals 2 damage to target creature."
//
// Firemaw Kavu's enters row: the Commando is the source of the damage,
// and a target that became illegal is dealt nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76c02534-35e3-4950-b4b3-90c679cdf6a7",
		Name:         "Goblin Commando",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Goblin Commando — 2 damage to target creature", effects.SourceDealsDamageToEachLegalTarget(2)),
				effects.TargetCreature("target creature")), effects.ForTargets(effects.DamageToTarget(0, 2))),
		},
	})
}
