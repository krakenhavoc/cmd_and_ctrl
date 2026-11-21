package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Firemaw Kavu — Creature — Kavu, {5}{R}, 4/2:
//
//	"Echo {5}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters, it deals 2 damage to target creature.
//	 When this creature leaves the battlefield, it deals 4 damage to target creature."
//
// "Leaves the battlefield" is any zone change from the battlefield, not
// only dying.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7ac09735-f05c-4e28-83d6-262cddc9eeb3",
		Name:         "Firemaw Kavu",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Echo("Firemaw Kavu", "{5}{R}"),
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Firemaw Kavu — 2 damage to target creature", effects.SourceDealsDamageToEachLegalTarget(2)),
				effects.TargetCreature("target creature")), effects.ForTargets(effects.DamageToTarget(0, 2))),
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisLeaves("Firemaw Kavu — 4 damage to target creature", effects.SourceDealsDamageToEachLegalTarget(4)),
				effects.TargetCreature("target creature")), effects.ForTargets(effects.DamageToTarget(0, 4))),
		},
	})
}
